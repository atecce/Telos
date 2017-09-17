package www

import (
	"log"
	"net"
	"regexp"
	"sync"
	"syscall"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"

	"github.com/de-nova-stella/investigations/canvas"
	"github.com/de-nova-stella/rest"
)

const domain = "http://www.lyrics.net"

// set regular expression for letter suburls
var artists = regexp.MustCompile("^artist/.*$")

type Investigator struct {
	expression string

	wg sync.WaitGroup
}

func inASCIIupper(str string) bool {
	for _, char := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		if string(char) == string(str[0]) {
			return true
		}
	}
	return false
}

func New(start string) *Investigator {

	investigator := new(Investigator)
	canvas.Init()

	if inASCIIupper(start) {
		investigator.expression = "^/artists/[" + string(start[0]) + "-Z]$"
	} else {
		investigator.expression = "^/artists/[0A-Z]$"
	}

	return investigator
}

func (investigator *Investigator) Run() {

	b, ok := rest.Get(domain)
	if !ok {
		return
	}
	defer b.Close()

	z := html.NewTokenizer(b)
	for {
		switch z.Next() {

		case html.ErrorToken:
			return

		case html.StartTagToken:
			if letterLink, ok := investigator.getLetterLink(z.Token()); ok {
				investigator.getArtists(letterLink)
			}
		}
	}
}

func (investigator *Investigator) getLetterLink(t html.Token) (string, bool) {

	letters, _ := regexp.Compile(investigator.expression)

	if t.Data == "a" {
		for _, a := range t.Attr {
			if a.Key == "href" && letters.MatchString(a.Val) {
				return domain + a.Val + "/99999", true
			}
		}
	}
	return "", false
}

func getArtistLinks(root *html.Node) []*html.Node {
	return scrape.FindAll(root, func(n *html.Node) bool {
		if n.Parent != nil {
			return n.Parent.Data == "strong" && n.Data == "a"
		}
		return false
	})
}

func (investigator *Investigator) getArtists(letter_url string) {

	// set body
	b, ok := rest.Get(letter_url)
	if !ok {
		return
	}
	defer b.Close()

	// parse page
	root, err := html.Parse(b)
	if err != nil {
		log.Fatal(err)
	}

	// find artist urls
	for _, link := range getArtistLinks(root) {

		artist_suburl := scrape.Attr(link, "href")

		if artists.MatchString(artist_suburl) {

			// concatenate artist url
			artist_url := domain + "/" + artist_suburl

			// extract artist name
			var artist_name string
			if link.FirstChild != nil {
				artist_name = link.FirstChild.Data
			}

			artist := &canvas.Artist{
				Url:  artist_url,
				Name: artist_name,
			}

			// parse the artist
			investigator.parseArtist(artist)
		}
	}
}

func (investigator *Investigator) parseArtist(artist *canvas.Artist) {

	// initialize artist flag
	var artistAdded bool

	// get body
	b, ok := rest.Get(artist.Url)
	if !ok {
		return
	}
	defer b.Close()

	// parse page
	z := html.NewTokenizer(b)
	for {
		switch z.Next() {

		// end of html document
		case html.ErrorToken:
			return

		// catch start tags
		case html.StartTagToken:

			// set token
			t := z.Token()

			// look for artist album labels
			if t.Data == "h3" {
				for _, a := range t.Attr {
					if a.Key == "class" && a.Val == "artist-album-label" {

						// add artist
						if !artistAdded {
							artist.Put()
							artistAdded = true
						}

						// album links are next token
						var album_url string
						z.Next()
						for _, album_attribute := range z.Token().Attr {
							if album_attribute.Key == "href" {
								album_url = domain + album_attribute.Val
							}
						}

						// album titles are the next token
						z.Next()
						album := &canvas.Album{
							Artist: artist,

							Url:  album_url,
							Name: z.Token().Data,
						}

						// add album
						album.Put()

						// parse album
						dorothy := investigator.parseAlbum(album)

						// handle dorothy
						if dorothy {
							investigator.no_place(album, z)
						}
					}
				}
			}
		}
	}
}

func (investigator *Investigator) no_place(album *canvas.Album, z *html.Tokenizer) {

	// parse album from artist page
	for {
		z.Next()
		t := z.Token()
		switch t.Data {

		// check for finished album
		case "div":

			for _, a := range t.Attr {
				if a.Key == "class" && a.Val == "clearfix" {
					investigator.wg.Wait()
					return
				}
			}

		// check for song links
		case "strong":

			z.Next()

			for _, a := range z.Token().Attr {
				if a.Key == "href" {

					// concatenate the url
					song_url := domain + a.Val

					// next token is artist name
					z.Next()
					song_title := z.Token().Data

					song := &canvas.Song{
						Url:  song_url,
						Name: song_title,
					}

					// parse song
					investigator.wg.Add(1)
					go investigator.parseSong(song)
				}
			}
		}
	}
}

func getSongLinks(root *html.Node) []*html.Node {
	return scrape.FindAll(root, func(n *html.Node) bool {
		if n.Parent != nil {
			return n.Parent.Data == "strong" && n.Data == "a"
		}
		return false
	})
}

func (investigator *Investigator) parseAlbum(album *canvas.Album) bool {

	// get body
	b, ok := rest.Get(album.Url)
	if !ok {
		return false
	}
	defer b.Close()

	// parse page
	root, err := html.Parse(b)
	if err != nil {
		log.Fatal(err)
	}

	// check for home page
	if _, dorothy := scrape.Find(root, func(n *html.Node) bool {
		return n.Data == "body" && scrape.Attr(n, "id") == "s4-page-homepage"
	}); dorothy {
		return true
	}

	// find song links
	song_links := getSongLinks(root)
	if len(song_links) == 0 {
		return true
	}

	// scrape links
	for _, link := range song_links {
		song_url := domain + scrape.Attr(link, "href")

		// title is first child
		var song_title string
		if link.FirstChild != nil {
			song_title = link.FirstChild.Data
		} else {
			panic(err)
		}

		// parse songs
		investigator.wg.Add(1)
		song := &canvas.Song{
			Url:  song_url,
			Name: song_title,
		}
		go investigator.parseSong(song)
	}

	// wait for songs
	investigator.wg.Wait()
	return false
}

func (investigator *Investigator) parseSong(song *canvas.Song) {

	// finish job at the end of function call
	defer investigator.wg.Done()

	// get body
	b, ok := rest.Get(song.Url)
	if !ok {
		return
	}
	defer b.Close()

	// parse page
	root, err := html.Parse(b)
	if err != nil {
		if operr, ok := err.(*net.OpError); ok {
			if operr.Err.Error() == syscall.ECONNRESET.Error() {
				investigator.wg.Add(1)
				investigator.parseSong(song)
				return
			}
		}
		panic(err)
	}

	// extract lyrics
	if lyrics_root, ok := scrape.Find(root, func(n *html.Node) bool {
		return n.Data == "pre" && scrape.Attr(n, "id") == "lyric-body-text"
	}); ok {
		song.Lyrics = scrape.Text(lyrics_root)
		song.Put()
	}
}
