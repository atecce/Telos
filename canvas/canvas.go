package canvas

import (
	"database/sql"
	"log"

	"github.com/kr/pretty"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Init() {
	database, err := sql.Open("sqlite3", "lyrics.net.db")
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
	tx, err := db.Begin()
	if err != nil {
		log.Println("[ERROR] beginning tx")
		pretty.Logln("[DEBUG] on db", db, "with err", err)
		return nil
	}
	return tx
}
