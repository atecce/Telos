package canvas

import (
	"fmt"
	"log"
	"sync"

	"github.com/de-nova-stella/rest"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/html"
)

type Artist struct {
	Url  string
	Name string
}

func initArtists() {

	if _, err := db.Exec(`create table if not exists artists (

				      name text not null,

				      primary key (name))`); err != nil {
		logger.Emerg("initializing artists")
		log.Fatal("sqlite: ", err)
	}
}

func FetchLatestArtist() (*Artist, error) {

	if db == nil {
		initDb()
	}

	rows, err := db.Query("select name from artists order by name desc")
	if err != nil {
		logger.Debug(fmt.Sprintf("failed to get latest artist name %v", err))
		return nil, err
	}
	defer rows.Close()

	var name string
	rows.Next()
	err = rows.Scan(&name)
	if err != nil {
		logger.Err(fmt.Sprintf("failed to scan rows %v", err))
		return nil, err
	}
	logger.Info(fmt.Sprintf("got latest %s", name))

	return &Artist{
		Name: name,
	}, nil
}

func (artist *Artist) Parse() {

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
								if dorothy := album.Parse(); dorothy {
									no_place(album, z)
								}
							}
						}
					}
				}
			}
		}
	}
}

func no_place(album *Album, z *html.Tokenizer) {

	// parse album from artist page
	for {
		z.Next()
		t := z.Token()
		switch t.Data {

		// check for finished album
		case "div":

			for _, a := range t.Attr {
				if a.Key == "class" && a.Val == "clearfix" {
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
					go song.Parse()
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
		logger.Err(fmt.Sprintf("preparing stmt %v for artist %v with err %v", stmt, artist, err))
		return
	}
	defer stmt.Close()

	if res, err := stmt.Exec(artist.Name); err != nil {
		logger.Err(fmt.Sprintf("execing stmt %v for artist %v with res %v and err %v", stmt, artist, res, err))
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Err(fmt.Sprintf("committing tx %v for artist %v with err %v", tx, artist, err))
		return
	}
}
