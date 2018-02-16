package main

import (
	"errors"
	"fmt"
	"log"
	"sync"

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
				     url 	 text not null,

				     name        text not null,
				     lyrics      text,

				     primary key (album, name),
				     foreign key (album) references albums (name))`)

	if err != nil {
		log.Fatal("failed to create tables:", err)
	}
}

func (song *Song) Parse() {

	root, b, err := parse(song.Url)
	if err != nil {
		fmt.Printf("failed to parse song url %s\n", song.Url)
		return
	}
	defer b.Close()

	song.Lyrics, err = scrapeLyrics(root)
	if err != nil {
		fmt.Printf("%s at %s\n", err.Error(), song.Url)
		return
	}
	song.put()
}

func scrapeLyrics(root *html.Node) (string, error) {
	if n, ok := scrape.Find(root, func(n *html.Node) bool {
		return n.Data == "pre" && scrape.Attr(n, "id") == "lyric-body-text"
	}); ok {
		return scrape.Text(n), nil
	}

	return "", errors.New("failed to scrape lyrics")
}

func (song *Song) put() {

	tx := begin()
	if tx == nil {
		return
	}

	stmt, err := tx.Prepare("insert or replace into songs (album, url, name, lyrics) values (?, ?, ?, ?)")
	if err != nil {
		fmt.Printf("failed preparing song at %s\n", song.Url)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(song.Album.Name, song.Url, song.Name, song.Lyrics)
	if err != nil {
		fmt.Printf("failed execing song at %s\n", song.Url)
		return
	}

	if err := tx.Commit(); err != nil {
		fmt.Printf("failed comitting song at %s\n", song.Url)
	}
}
