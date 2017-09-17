package canvas

import (
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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

func PutSong(album, name, lyrics string) {

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	var failed bool
	for {

		stmt, err := tx.Prepare("insert or replace into songs (album, name, lyrics) values (?, ?, ?)")
		if err != nil {
			failed = true
			log.Println("Error in .Prepare: Failed to add song", name, "in album", album+":", err)
			time.Sleep(time.Second)
			continue
		}
		defer stmt.Close()

		_, err = stmt.Exec(album, name, lyrics)
		if err != nil {
			failed = true
			log.Println("Error in .Exec: Failed to add song", name, "in album", album+":", err)
			time.Sleep(time.Second)
			continue
		}
		tx.Commit()

		if failed {
			log.Println("Successfully added song", name, "in album", album)
		}
		return
	}
}
