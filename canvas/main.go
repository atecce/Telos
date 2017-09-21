package canvas

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"

	"github.com/de-nova-stella/rest"
	"github.com/kr/pretty"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/html"
)

var (
	db     *sql.DB
	domain *url.URL
)

func initDb() {

	domain, _ = url.Parse("http://www.lyrics.net")

	database, err := sql.Open("sqlite3", "/keybase/private/atec/lyrics.net.db")
	if err != nil {
		pretty.Logln("[FATAL] failed to initialize db")
		log.Fatal(err)
	}
	db = database

	initArtists()
	initAlbums()
	initSongs()
}

func begin() *sql.Tx {

	if db == nil {
		initDb()
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println("[ERROR] beginning tx")
		pretty.Logln("[DEBUG] on db", db, "with err", err)
		return nil
	}
	return tx
}

func parse(url string) (*html.Node, io.ReadCloser, error) {

	b, ok := rest.Get(url)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("failed to get url: %s", url))
	}

	root, err := html.Parse(b)
	if err != nil {
		b.Close()
		return nil, nil, err
	}

	return root, b, nil
}