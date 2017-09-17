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

func PutSong(song Song) {

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
