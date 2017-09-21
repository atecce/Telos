package canvas

import (
	"log"
	"sync"

	"github.com/de-nova-stella/rest"
	"github.com/kr/pretty"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/html"
)

type Artist struct {
	Url  string
	Name string
}

func initArtists() {
	if res, err := db.Exec(`create table if not exists artists (

				      name text not null,

				      primary key (name))`); err != nil {
		pretty.Logln("[FATAL] initializing artists")
		log.Fatal(res, err)
	}
}

func FetchLatestArtist() (*Artist, error) {

	rows, err := db.Query("select name from artists order by name desc")
	if err != nil {
		pretty.Logln("[DEBUG] failed to get latest artist name", rows, err)
		return nil, err
	}
	defer rows.Close()

	var name string
	rows.Next()
	err = rows.Scan(&name)
	if err != nil {
		pretty.Logln("[DEBUG] failed to get scan rows", rows, err)
		return nil, err
	}
	log.Println("[INFO] got latest", name)

	return &Artist{
		Name: name,
	}, nil
}

func (artist *Artist) Parse(wg *sync.WaitGroup) {

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
							artist.put()
							artistAdded = true
						}

						// album links are next token
						z.Next()
						for _, attr := range z.Token().Attr {
							if attr.Key == "href" {

								u := *domain
								u.Path = attr.Val

								// album titles are the next token
								z.Next()
								album := &Album{
									Artist: artist,

									Url:  u.String(),
									Name: z.Token().Data,
								}
								album.put()

								// parse album
								if dorothy := album.Parse(wg); dorothy {
									no_place(album, z, wg)
								}
							}
						}
					}
				}
			}
		}
	}
}

func no_place(album *Album, z *html.Tokenizer, wg *sync.WaitGroup) {

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

					u := *domain
					u.Path = a.Val

					// next token is artist name
					z.Next()
					song := &Song{
						Url:  u.String(),
						Name: z.Token().Data,
					}

					// parse song
					wg.Add(1)
					go song.Parse(wg)
				}
			}
		}
	}
}
func (artist *Artist) put() {

	tx := begin()
	if tx == nil {
		return
	}

	stmt, err := tx.Prepare("insert or replace into artists (name) values (?)")
	if err != nil {
		pretty.Logln("[ERROR] preparing stmt for artist", artist)
		pretty.Logln("[INFO]", stmt, err)
		return
	}
	defer stmt.Close()

	if res, err := stmt.Exec(artist.Name); err != nil {
		pretty.Logln("[ERROR] execing stmt for artist", artist)
		pretty.Logln("[INFO]", res, err)
		return
	}

	if err := tx.Commit(); err != nil {
		pretty.Logln("[ERROR] committing tx for artist", artist)
		pretty.Logln("[INFO]", err)
		return
	}
}
