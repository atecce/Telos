package canvas

import (
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Artist struct {
	Url  string
	Name string
}

func initArtists() {
	_, err := db.Exec(`create table if not exists artists (

				      name text not null,

				      primary key (name))`)
	if err != nil {
		panic(err)
	}
}

func PutArtist(artist Artist) {

	tx, err := db.Begin()
	stmt, err := tx.Prepare("insert or replace into artists (name) values (?)")
	defer stmt.Close()
	_, err = stmt.Exec(artist.Name)
	tx.Commit()

	if err != nil {
		log.Println("Failed to add artist", artist.Name+":", err)
	}
}
