package canvas

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net/url"

	"github.com/de-nova-stella/rest"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/html"
)

var (
	db     *sql.DB
	domain *url.URL
	logger *syslog.Writer
)

func initDb() {

	syslogger, err := syslog.Dial("", "", syslog.LOG_USER, "")
	if err != nil {
		log.Fatal("syslog:", err)

	}
	logger = syslogger

	domain, _ = url.Parse("http://www.lyrics.net")

	database, err := sql.Open("sqlite3", "/keybase/private/atec/lyrics.net.db")
	if err != nil {
		logger.Emerg(fmt.Sprintf("failed to initialize db %v", err))
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
		logger.Err("beginning tx")
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
