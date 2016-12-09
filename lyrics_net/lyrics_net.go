package lyrics_net

import (
	"database/sql"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/atecce/investigations/db"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

// set wait group
var wg sync.WaitGroup

// get url
var url string = "http://www.lyrics.net"

// set caught up variable
var caught_up bool

func communicate(url string) (bool, io.ReadCloser) {

	// never stop trying
	for {

		// get url
		resp, err := http.Get(url)

		// catch error
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second)
			continue
		}

		// write status to output
		log.Println(url, resp.Status)

		// check status codes
		switch resp.StatusCode {

		// cases which are returned
		case http.StatusOK:
			return false, resp.Body
		case http.StatusForbidden:
			return true, resp.Body
		case http.StatusNotFound:
			return true, resp.Body

			// cases which are retried
		case http.StatusServiceUnavailable:
			time.Sleep(10 * time.Minute)
		case http.StatusGatewayTimeout:
			time.Sleep(time.Minute)
		default:
			time.Sleep(time.Minute)
		}
	}
}

func inASCIIupper(start string) bool {
	for _, char := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		if string(char) == string(start[0]) {
			return true
		}
	}
	return false
}

func Investigate(start string) {

	// initiate db
	canvas := db.InitiateDB("lyrics_net")
	defer canvas.Close()

	// use specified start letter
	var expression string
	if inASCIIupper(start) {
		expression = "^/artists/[" + string(start[0]) + "-Z]$"
	} else {
		expression = "^/artists/[0A-Z]$"
	}

	// set regular expression for letter suburls
	letters, _ := regexp.Compile(expression)

	// set body
	skip, b := communicate(url)
	defer b.Close()

	// check for skip
	if skip {
		return
	}

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

			// look for matching letter suburl
			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key == "href" {
						if letters.MatchString(a.Val) {

							// concatenate the url
							letter_url := url + a.Val + "/99999"

							// get artists
							getArtists(start, letter_url, canvas)
						}
					}
				}
			}
		}
	}
}

func getArtists(start, letter_url string, canvas *sql.DB) {

	// set caught up expression
	expression, _ := regexp.Compile("^" + start + ".*$")
	if start == "0" {
		caught_up = true
	}

	// set regular expression for letter suburls
	artists, _ := regexp.Compile("^artist/.*$")

	// set body
	skip, b := communicate(letter_url)
	defer b.Close()

	// check for skip
	if skip {
		return
	}

	// parse page
	z := html.NewTokenizer(b)
	for {
		switch z.Next() {

		// end of document
		case html.ErrorToken:
			return

		// catch start tags
		case html.StartTagToken:

			// find artist urls
			if z.Token().Data == "strong" {
				z.Next()
				for _, a := range z.Token().Attr {
					if a.Key == "href" {
						if artists.MatchString(a.Val) {

							// concatenate the url
							artist_url := url + "/" + a.Val

							// next token is artist name
							z.Next()
							artist_name := z.Token().Data

							// check if caught up
							if expression.MatchString(artist_name) {
								caught_up = true
							}
							if !caught_up {
								continue
							}

							// parse the artist
							parseArtist(artist_url, artist_name, canvas)
						}
					}
				}
			}
		}
	}
}

func parseArtist(artist_url, artist_name string, canvas *sql.DB) {

	// initialize artist flag
	var artistAdded bool

	// set body
	skip, b := communicate(artist_url)
	defer b.Close()

	// check for skip
	if skip {
		return
	}

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
							db.AddArtist(artist_name, canvas)
							artistAdded = true
						}

						// album links are next token
						var album_url string
						z.Next()
						for _, album_attribute := range z.Token().Attr {
							if album_attribute.Key == "href" {
								album_url = url + album_attribute.Val
							}
						}

						// album titles are the next token
						z.Next()
						album_title := z.Token().Data

						// add album
						db.AddAlbum(artist_name, album_title, canvas)

						// parse album
						dorothy := parseAlbum(album_url, album_title, canvas)

						// handle dorothy
						if dorothy {
							no_place(album_title, z, canvas)
						}
					}
				}
			}
		}
	}
}

func no_place(album_title string, z *html.Tokenizer, canvas *sql.DB) {

	// parse album from artist page
	for {
		z.Next()
		t := z.Token()
		switch t.Data {

		// check for finished album
		case "div":

			for _, a := range t.Attr {
				if a.Key == "class" && a.Val == "clearfix" {
					wg.Wait()
					return
				}
			}

		// check for song links
		case "strong":

			z.Next()

			for _, a := range z.Token().Attr {
				if a.Key == "href" {

					// concatenate the url
					song_url := url + a.Val

					// next token is artist name
					z.Next()
					song_title := z.Token().Data

					// parse song
					wg.Add(1)
					go parseSong(song_url, song_title, album_title, canvas)
				}
			}
		}
	}
}

func parseAlbum(album_url, album_title string, canvas *sql.DB) bool {

	// initialize flag that checks for songs
	var has_songs bool

	// set body
	skip, b := communicate(album_url)
	defer b.Close()

	// check for skip
	if skip {
		return false
	}

	// parse page
	z := html.NewTokenizer(b)
	for {
		switch z.Next() {

		// end of html document
		case html.ErrorToken:
			wg.Wait()
			return !has_songs

		// catch start tags
		case html.StartTagToken:

			// check token
			t := z.Token()
			switch t.Data {

			// check for home page
			case "body":
				for _, a := range t.Attr {
					if a.Key == "id" && a.Val == "s4-page-homepage" {
						return true
					}
				}

			// find song links
			case "strong":
				z.Next()
				for _, a := range z.Token().Attr {
					if a.Key == "href" {

						// mark that the page has songs
						has_songs = true

						// concatenate the url
						song_url := url + a.Val

						// next token is artist name
						z.Next()
						song_title := z.Token().Data

						// parse song
						wg.Add(1)
						go parseSong(song_url, song_title, album_title, canvas)
					}
				}
			}
		}
	}
}

func parseSong(song_url, song_title, album_title string, canvas *sql.DB) {

	// finish job at the end of function call
	defer wg.Done()

	// set body
	skip, b := communicate(song_url)
	defer b.Close()

	// check for skip
	if skip {
		return
	}

	// parse page
	root, err := html.Parse(b)
	if err != nil {
		if operr, ok := err.(*net.OpError); ok {
			if operr.Err.Error() == syscall.ECONNRESET.Error() {
				wg.Add(1)
				parseSong(song_url, song_title, album_title, canvas)
				return
			}
		}
		panic(err)
	}

	// get root of lyrics element
	lyrics_root, ok := scrape.Find(root, func(n *html.Node) bool {
		return n.Data == "pre" && scrape.Attr(n, "id") == "lyric-body-text"
	})

	// means no lyrics are listed
	if !ok {
		return
	}

	// extract lyrics
	var lyrics string
	for n := lyrics_root.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.TextNode {
			lyrics += n.Data
		} else {
			if n.FirstChild != nil {
				lyrics += n.FirstChild.Data
			}
		}
	}

	// add song to db
	db.AddSong(album_title, song_title, lyrics, canvas)
}
