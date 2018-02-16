package main

import (
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

type Album struct {
	Artist *Artist

	Url  string
	Name string
}

func initAlbums() {
	_, err := db.Exec(`create table if not exists albums (

				     artist	 text not null,

				     name        text not null,

				     primary key (name, artist),
				     foreign key (artist) references artists (name))`)
	if err != nil {
		panic(err)
	}

}

func (album *Album) Parse() bool {

	root, b, err := parse(album.Url)
	if err != nil {
		fmt.Printf("parsing album url %s\n", album.Url)
		return false
	}
	defer b.Close()

	// check for home page
	if _, dorothy := scrape.Find(root, func(n *html.Node) bool {
		return n.Data == "body" && scrape.Attr(n, "id") == "s4-page-homepage"
	}); dorothy {
		return true
	}

	songs := scrapeSongs(root)
	if len(songs) == 0 {
		return true
	}

	// scrape links
	for _, link := range songs {

		// title is first child
		if link.FirstChild != nil {

			u := *domain
			u.Path = scrape.Attr(link, "href")

			song := &Song{
				Album: album,

				Url:  u.String(),
				Name: link.FirstChild.Data,
			}
			go song.Parse()
		}
	}

	return false
}

func scrapeSongs(root *html.Node) []*html.Node {
	return scrape.FindAll(root, func(n *html.Node) bool {
		if n.Parent != nil {
			return n.Parent.Data == "strong" && n.Data == "a"
		}
		return false
	})
}

func (album *Album) put() {

	tx := begin()
	if tx == nil {
		return
	}

	stmt, err := tx.Prepare("insert or replace into albums (artist, name) values (?, ?)")
	if err != nil {
		fmt.Printf("preparing stmt %v for album %v with err %v\n", stmt, album, err)
	}

	_, err = stmt.Exec(album.Artist.Name, album.Name)
	if err != nil {
		fmt.Printf("execing stmt %v for album %v with err %v\n", stmt, album, err)
	}

	if err := tx.Commit(); err != nil {
		fmt.Printf("committing tx %v for artist %v with err %v\n", tx, album, err)
	}

	if err := stmt.Close(); err != nil {
		fmt.Printf("closing stmt %v for album %v with err %v\n", stmt, album, err)
	}
}
