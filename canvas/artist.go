package canvas

import (
	"log"

	"github.com/kr/pretty"
	_ "github.com/mattn/go-sqlite3"
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

func (artist *Artist) Put() {

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
