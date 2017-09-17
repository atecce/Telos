package www

import (
	"log"
	"regexp"
	"sync"

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

	// TODO shared ref
	wg *sync.WaitGroup
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
	investigator.wg = new(sync.WaitGroup)

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

		if artist_suburl := scrape.Attr(link, "href"); artists.MatchString(artist_suburl) {

			// extract artist name
			var artist_name string
			if link.FirstChild != nil {
				artist_name = link.FirstChild.Data
			}

			artist := &canvas.Artist{
				Url:  domain + "/" + artist_suburl,
				Name: artist_name,
			}

			// parse the artist
			artist.Parse(investigator.wg)
		}
	}
}

func getArtistLinks(root *html.Node) []*html.Node {
	return scrape.FindAll(root, func(n *html.Node) bool {
		if n.Parent != nil {
			return n.Parent.Data == "strong" && n.Data == "a"
		}
		return false
	})
}
