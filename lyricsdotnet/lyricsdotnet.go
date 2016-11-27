package lyricsdotnet

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/atecce/investigations/db"

	_ "github.com/mattn/go-sqlite3" // need this to declare sqlite3 pointer
	"golang.org/x/net/html"
)

// set wait group
var wg sync.WaitGroup

// get url
const url = "http://www.lyrics.net"

// constant flags
const href = "href"
const strong = "strong"

// set caught up variable
var caughtUp bool

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
		log.Println("INFO", url, resp.Status)

		// check status codes
		switch resp.StatusCode {

		// cases which are returned
		case 200:
			return false, resp.Body
		case 403:
			return true, resp.Body
		case 404:
			return true, resp.Body

		// cases which are retried
		case 503:
			time.Sleep(30 * time.Minute)
		case 504:
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

// Investigate called to start web scrape
func Investigate(start string) {

	// initiate db
	canvas := db.InitiateDB("lyrics_net")
	defer canvas.Close()

	// use specified start letter
	var expression string
	if start == "0" || start == "" {
		expression = "^/artists/[0A-Z]$"
	} else if inASCIIupper(start) {
		expression = "^/artists/[" + string(start[0]) + "-Z]$"
	} else {
		log.Println("Invalid start character.")
		return
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
					if a.Key == href {
						if letters.MatchString(a.Val) {

							// concatenate the url
							letterURL := url + a.Val + "/99999"

							// get artists
							getArtists(start, letterURL, canvas)
						}
					}
				}
			}
		}
	}
}

func getArtists(start, letterURL string, canvas *sql.DB) {

	// set caught up expression
	expression, _ := regexp.Compile("^" + start + ".*$")
	if start == "0" {
		caughtUp = true
	}

	// set regular expression for letter suburls
	artists, _ := regexp.Compile("^artist/.*$")

	// set body
	skip, b := communicate(letterURL)
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
			if z.Token().Data == strong {
				z.Next()
				for _, a := range z.Token().Attr {
					if a.Key == href {
						if artists.MatchString(a.Val) {

							// concatenate the url
							artistURL := url + "/" + a.Val

							// next token is artist name
							z.Next()
							artistName := z.Token().Data

							// check if caught up
							if expression.MatchString(artistName) {
								caughtUp = true
							}
							if !caughtUp {
								continue
							}

							// parse the artist
							parseArtist(artistURL, artistName, canvas)
						}
					}
				}
			}
		}
	}
}

func parseArtist(artistURL, artistName string, canvas *sql.DB) {

	// initialize artist flag
	var artistAdded bool

	// set body
	skip, b := communicate(artistURL)
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
							db.AddArtist(artistName, canvas)
							artistAdded = true
						}

						// album links are next token
						var albumURL string
						z.Next()
						for _, albumAttribute := range z.Token().Attr {
							if albumAttribute.Key == href {
								albumURL = url + albumAttribute.Val
							}
						}

						// album titles are the next token
						z.Next()
						albumTitle := z.Token().Data

						// add album
						db.AddAlbum(artistName, albumTitle, canvas)

						// parse album
						dorothy := parseAlbum(albumURL, albumTitle, canvas)

						// handle dorothy
						if dorothy {
							noPlace(albumTitle, z, canvas)
						}
					}
				}
			}
		}
	}
}

func noPlace(albumTitle string, z *html.Tokenizer, canvas *sql.DB) {

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
		case strong:

			z.Next()

			for _, a := range z.Token().Attr {
				if a.Key == href {

					// concatenate the url
					songURL := url + a.Val

					// next token is artist name
					z.Next()
					songTitle := z.Token().Data

					// parse song
					wg.Add(1)
					go parseSong(songURL, songTitle, albumTitle, canvas)
				}
			}
		}
	}
}

func parseAlbum(albumURL, albumTitle string, canvas *sql.DB) bool {

	// initialize flag that checks for songs
	var hasSongs bool

	// set body
	skip, b := communicate(albumURL)
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
			return !hasSongs

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
			case strong:
				z.Next()
				for _, a := range z.Token().Attr {
					if a.Key == href {

						// mark that the page has songs
						hasSongs = true

						// concatenate the url
						songURL := url + a.Val

						// next token is artist name
						z.Next()
						songTitle := z.Token().Data

						// parse song
						wg.Add(1)
						go parseSong(songURL, songTitle, albumTitle, canvas)
					}
				}
			}
		}
	}
}

func parseSong(songURL, songTitle, albumTitle string, canvas *sql.DB) {

	// set body
	skip, b := communicate(songURL)
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
			wg.Done()
			return

		// catch start tags
		case html.StartTagToken:

			// find pre tokens
			if z.Token().Data == "pre" {

				// next token is lyrics
				z.Next()
				lyrics := z.Token().Data

				// add song to db
				db.AddSong(albumTitle, songTitle, lyrics, canvas)
			}
		}
	}
}
