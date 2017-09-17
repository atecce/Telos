package canvas

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Canvas struct {
	db   *sql.DB
	name string
}

func New(name string) *Canvas {

	db, err := sql.Open("sqlite3", name+".db")

	_, err = db.Exec(`create table if not exists artists (

				      name text not null,

				      primary key (name))`)

	_, err = db.Exec(`create table if not exists albums (

				     title       text not null,
				     artist	 text not null,

				     primary key (title, artist),
				     foreign key (artist) references artists (name))`)

	_, err = db.Exec(`create table if not exists songs (

				     title       text not null,
				     album 	 text not null,
				     lyrics      text,

				     primary key (album, title),
				     foreign key (album) references albums (title))`)

	if err != nil {
		log.Println("Failed to create tables:", err)
	}

	return &Canvas{
		db:   db,
		name: name,
	}
}

func (canvas *Canvas) AddArtist(artist_name string) {

	tx, err := canvas.db.Begin()
	stmt, err := tx.Prepare("insert or replace into artists (name) values (?)")
	defer stmt.Close()
	_, err = stmt.Exec(artist_name)
	tx.Commit()

	if err != nil {
		log.Println("Failed to add artist", artist_name+":", err)
	}
}

func (canvas *Canvas) AddAlbum(artist_name, album_title string) {

	tx, err := canvas.db.Begin()
	stmt, err := tx.Prepare("insert or replace into albums (artist, title) values (?, ?)")
	defer stmt.Close()
	_, err = stmt.Exec(artist_name, album_title)
	tx.Commit()

	if err != nil {
		log.Fatal("Failed to add album", album_title, "by", artist_name+":", err)
	}
}

func (canvas *Canvas) AddSong(album_title, song_title, lyrics string) {

	tx, err := canvas.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	var failed bool
	for {

		stmt, err := tx.Prepare("insert or replace into songs (album, title, lyrics) values (?, ?, ?)")
		if err != nil {
			failed = true
			log.Println("Error in .Prepare: Failed to add song", song_title, "in album", album_title+":", err)
			time.Sleep(time.Second)
			continue
		}
		defer stmt.Close()

		_, err = stmt.Exec(album_title, song_title, lyrics)
		if err != nil {
			failed = true
			log.Println("Error in .Exec: Failed to add song", song_title, "in album", album_title+":", err)
			time.Sleep(time.Second)
			continue
		}
		tx.Commit()

		if failed {
			log.Println("Successfully added song", song_title, "in album", album_title)
		}
		return
	}
}
