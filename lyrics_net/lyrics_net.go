package lyrics_net

import (
	"db"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
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
			log.Println("\n", err, "\n")
			time.Sleep(time.Second)
			continue
		}

		// write status to output
		fmt.Println(time.Now(), url, resp.Status)

		// check status codes
		switch resp.StatusCode {

			// cases which are returned
			case 200: return false, resp.Body 
			case 403: return true,  resp.Body 
			case 404: return true,  resp.Body 

			// cases which are retried
			case 503: time.Sleep(30 * time.Minute) 
			case 504: time.Sleep(time.Minute) 
			default:  time.Sleep(time.Minute)
		}
	}
}

func inASCIIupper(start string) bool {
	for _, char := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" { if string(char) == string(start[0]) { return true } }
	return false
}

func Investigate(start string) {

	// initiate db
	db.InitiateDB("lyrics_net")

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
	if skip { return }

	// parse page
	z := html.NewTokenizer(b)
	for { switch z.Next() {

		// end of html document
		case html.ErrorToken: return

		// catch start tags
		case html.StartTagToken:

			// set token
			t := z.Token()

			// look for matching letter suburl
			if t.Data == "a" { for _, a := range t.Attr { if a.Key == "href" { if letters.MatchString(a.Val) {

				// concatenate the url
				letter_url := url + a.Val + "/99999"

				// get artists
				getArtists(start, letter_url)
			}}}}
		}
	}
}

func getArtists(start, letter_url string) {

	// set caught up expression
	expression, _ := regexp.Compile("^" + start + ".*$")
	if start == "0" { caught_up = true }

	// set regular expression for letter suburls
	artists, _ := regexp.Compile("^artist/.*$")

	// set body
	skip, b := communicate(letter_url)
	defer b.Close()

	// check for skip
	if skip { return }

	// parse page
	z := html.NewTokenizer(b)
	for { switch z.Next() {

		// end of document
		case html.ErrorToken: return

		// catch start tags
		case html.StartTagToken:

			// find artist urls
			if z.Token().Data == "strong" { z.Next(); for _, a := range z.Token().Attr { if a.Key == "href" { if artists.MatchString(a.Val) {

				// concatenate the url
				artist_url := url + "/" + a.Val

				// next token is artist name
				z.Next(); artist_name := z.Token().Data

				// check if caught up
				if expression.MatchString(artist_name) { caught_up = true }
				if !caught_up { continue }

				// parse the artist
				parseArtist(artist_url, artist_name)
			}}}}
		}
	}
}

func parseArtist(artist_url, artist_name string) {

	// initialize artist flag
	var artistAdded bool

	// set body
	skip, b := communicate(artist_url)
	defer b.Close()

	// check for skip
	if skip { return }

	// parse page
	z := html.NewTokenizer(b)
	for { switch z.Next() {

		// end of html document
		case html.ErrorToken: 
			return

		// catch start tags
		case html.StartTagToken:

			// set token
			t := z.Token()

			// look for artist album labels
			if t.Data == "h3" { for _, a := range t.Attr { if a.Key == "class" && a.Val == "artist-album-label" {

				// add artist
				if !artistAdded { db.AddArtist(artist_name); artistAdded = true }

				// album links are next token
				var album_url string
				z.Next()
				for _, album_attribute := range z.Token().Attr { if album_attribute.Key == "href" { album_url = url + album_attribute.Val } }

				// album titles are the next token
				z.Next(); album_title := z.Token().Data

				// add album
				db.AddAlbum(artist_name, album_title)

				// parse album
				dorothy := parseAlbum(album_url, album_title)

				// handle dorothy
				if dorothy { no_place(album_title, z) }
			}}}
		}
	}
}

func no_place(album_title string, z *html.Tokenizer) {

	// parse album from artist page
	for { z.Next(); t := z.Token(); switch t.Data {

		// check for finished album
		case "div": 
		
			for _, a := range t.Attr { if a.Key == "class" && a.Val == "clearfix" { 
				wg.Wait()
				return
			}}

		// check for song links
		case "strong": 
		
			z.Next()
			
			for _, a := range z.Token().Attr { if a.Key == "href" {

				// concatenate the url
				song_url := url + a.Val

				// next token is artist name
				z.Next(); song_title := z.Token().Data

				// parse song
				wg.Add(1)
				go parseSong(song_url, song_title, album_title)
			}}
	}}
}

func parseAlbum(album_url, album_title string) bool {

	// initialize flag that checks for songs
	var has_songs bool

	// set body
	skip, b := communicate(album_url)
	defer b.Close()

	// check for skip
	if skip { return false }

	// parse page
	z := html.NewTokenizer(b)
	for { switch z.Next() {

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
				case "body": for _, a := range t.Attr { if a.Key == "id" && a.Val =="s4-page-homepage" { return true } }

				// find song links
				case "strong": z.Next(); for _, a := range z.Token().Attr { if a.Key == "href" {

					// mark that the page has songs
					has_songs = true

					// concatenate the url
					song_url := url + a.Val

					// next token is artist name
					z.Next(); song_title := z.Token().Data

					// parse song
					wg.Add(1)
					go parseSong(song_url, song_title, album_title)
				}}
			}
		}
	}
}

func parseSong(song_url, song_title, album_title string) {

	// set body
	skip, b := communicate(song_url)
	defer b.Close()

	// check for skip
	if skip { return }

	// parse page
	z := html.NewTokenizer(b)
	for { switch z.Next() {

		// end of html document
		case html.ErrorToken:
			wg.Done()
			return

		// catch start tags
		case html.StartTagToken:

			// find pre tokens
			if z.Token().Data == "pre" {

				// next token is lyrics
				z.Next(); lyrics := z.Token().Data

				// add song to db
				db.AddSong(album_title, song_title, lyrics)
			}
		}
	}
}
