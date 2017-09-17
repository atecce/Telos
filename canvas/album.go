package canvas

import (
	"log"
	"sync"

	"github.com/de-nova-stella/rest"
	"github.com/kr/pretty"
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

func (album *Album) Parse(wg *sync.WaitGroup) bool {

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
		wg.Add(1)
		song := &Song{
			Album: album,

			Url:  song_url,
			Name: song_title,
		}
		go song.Parse(wg)
	}

	// wait for songs
	wg.Wait()
	return false
}

func (album *Album) put() {

	tx := begin()
	if tx == nil {
		return
	}

	stmt, err := tx.Prepare("insert or replace into albums (artist, name) values (?, ?)")
	if err != nil {
		log.Println("[ERROR] preparing stmt for album", album)
		pretty.Logln("[DEBUG] on tx", tx, "with err", err)
	}

	_, err = stmt.Exec(album.Artist.Name, album.Name)
	if err != nil {
		log.Println("[ERROR] execing stmt for album", album)
		pretty.Logln("[DEBUG] on stmt", stmt, "with err", err)
	}

	if err := tx.Commit(); err != nil {
		log.Println("[ERROR] committing tx for album", album)
		pretty.Logln("[DEBUG] on tx", tx, "with err", err)
	}

	if err := stmt.Close(); err != nil {
		log.Println("[ERROR] closing stmt for album", album)
		pretty.Logln("[DEBUG] on stmt", stmt, "with err", err)
	}
}
