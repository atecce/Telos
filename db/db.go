package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // need this to declare sqlite3 pointer
)

func InitiateDB(name string) *sql.DB {

	// prepare db
	canvas, err := sql.Open("sqlite3", name+".db")

	// create tables
	_, err = canvas.Exec(`create table if not exists artists (

				      name text not null,

				      primary key (name))`)

	_, err = canvas.Exec(`create table if not exists albums (

				     title      text not null,
				     artistName text not null,

				     primary key (title, artistName),
				     foreign key (artistName) references artists (name))`)

	_, err = canvas.Exec(`create table if not exists songs (

				     title      text not null,
				     albumTitle text not null,
				     lyrics     text,

				     primary key (albumTitle, title),
				     foreign key (albumTitle) references albums (title))`)

	// catch error
	if err != nil {
		log.Println("ERROR Failed to create tables:", err)
	}

	return canvas
}

func AddArtist(artistName string, canvas *sql.DB) {

	// prepare db
	tx, err := canvas.Begin()

	// insert entry
	stmt, err := tx.Prepare("insert or replace into artists (name) values (?)")
	defer stmt.Close()
	_, err = stmt.Exec(artistName)
	tx.Commit()

	// catch error
	if err != nil {
		log.Println("ERROR Failed to add artist", artistName+":", err)
	}
}

func AddAlbum(artistName, albumTitle string, canvas *sql.DB) {

	// prepare db
	tx, err := canvas.Begin()

	// insert entry
	stmt, err := tx.Prepare("insert or replace into albums (artistName, title) values (?, ?)")
	defer stmt.Close()
	_, err = stmt.Exec(artistName, albumTitle)
	tx.Commit()

	// catch error
	if err != nil {
		log.Println("ERROR Failed to add album", albumTitle, "by", artistName+":", err)
	}
}

func AddSong(albumTitle, songTitle, lyrics string, canvas *sql.DB) {

	// initialized failed flag
	var failed bool

	for {

		// prepare db
		tx, err := canvas.Begin()

		// catch error
		if err != nil {
			failed = true
			log.Println("ERROR .Begin: Failed to add song", songTitle, "in album", albumTitle+":", err)
			time.Sleep(time.Second)
			canvas.Close()
			continue
		}

		// prepare statement
		stmt, err := tx.Prepare("insert or replace into songs (albumTitle, title, lyrics) values (?, ?, ?)")

		// catch error
		if err != nil {
			failed = true
			log.Println("ERROR .Prepare: Failed to add song", songTitle, "in album", albumTitle+":", err)
			time.Sleep(time.Second)
			canvas.Close()
			continue
		}

		// close statement
		defer stmt.Close()

		// execute statement
		_, err = stmt.Exec(albumTitle, songTitle, lyrics)

		// catch error
		if err != nil {
			failed = true
			log.Println("ERROR .Exec: Failed to add song", songTitle, "in album", albumTitle+":", err)
			time.Sleep(time.Second)
			canvas.Close()
			continue
		}

		// commit changes
		tx.Commit()

		// notify that a previous failure was cleaned up
		if failed {
			log.Println("INFO Successfully added song", songTitle, "in album", albumTitle)
		}

		// exit
		return
	}
}
