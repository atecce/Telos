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
	songs := getSongs(root)
	if len(songs) == 0 {
		return true
	}

	// scrape links
	for _, link := range songs {

		// title is first child
		if link.FirstChild != nil {

			u := *domain
			u.Path = scrape.Attr(link, "href")

			wg.Add(1)
			song := &Song{
				Album: album,

				Url:  u.String(),
				Name: link.FirstChild.Data,
			}
			go song.Parse(wg)
		}
	}

	wg.Wait()
	return false
}

func getSongs(root *html.Node) []*html.Node {
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
