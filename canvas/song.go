package canvas

import (
	"log"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/de-nova-stella/rest"
	"github.com/kr/pretty"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

const debug = false

type Song struct {
	Album *Album

	Url    string
	Name   string
	Lyrics string
}

func initSongs() {
	_, err := db.Exec(`create table if not exists songs (

				     album 	 text not null,

				     name        text not null,
				     lyrics      text,

				     primary key (album, name),
				     foreign key (album) references albums (name))`)

	if err != nil {
		log.Println("Failed to create tables:", err)
	}
}

func (song *Song) Parse(wg *sync.WaitGroup) {

	defer wg.Done()

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
				wg.Add(1)
				song.Parse(wg)
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
		song.put()
	}
}

func (song *Song) put() {

	tx := begin()
	if tx == nil {
		return
	}

	var failed bool
	for {

		stmt, err := tx.Prepare("insert or replace into songs (album, name, lyrics) values (?, ?, ?)")
		if err != nil {
			failed = true
			log.Println("[ERROR] preparing", song.Name, "for album", song.Album.Name+":", err)
			time.Sleep(time.Second)
			continue
		}
		defer stmt.Close()

		if debug {
			pretty.Logln("[DEBUG] execing stmt", stmt, "for song", song)
		}
		_, err = stmt.Exec(song.Album.Name, song.Name, song.Lyrics)
		if err != nil {
			failed = true
			log.Println("[ERROR] execing", song.Name, "for album", song.Album.Name+":", err)
			time.Sleep(time.Second)
			continue
		}
		tx.Commit()

		if failed {
			log.Println("[INFO] successfully added", song.Name, "to album", song.Album.Name)
		}
		return
	}
}
