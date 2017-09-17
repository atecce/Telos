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

type Artist struct {
	Url  string
	Name string
}

type Album struct {
	Artist *Artist

	Url  string
	Name string
}

type Song struct {
	Album *Album

	Url    string
	Name   string
	Lyrics string
}

func New(name string) *Canvas {

	db, err := sql.Open("sqlite3", name+".db")

	_, err = db.Exec(`create table if not exists artists (

				      name text not null,

				      primary key (name))`)

	_, err = db.Exec(`create table if not exists albums (

				     artist	 text not null,

				     name        text not null,

				     primary key (name, artist),
				     foreign key (artist) references artists (name))`)

	_, err = db.Exec(`create table if not exists songs (

				     album 	 text not null,

				     name        text not null,
				     lyrics      text,

				     primary key (album, name),
				     foreign key (album) references albums (name))`)

	if err != nil {
		log.Println("Failed to create tables:", err)
	}

	return &Canvas{
		db:   db,
		name: name,
	}
}

func (canvas *Canvas) AddArtist(name string) {

	tx, err := canvas.db.Begin()
	stmt, err := tx.Prepare("insert or replace into artists (name) values (?)")
	defer stmt.Close()
	_, err = stmt.Exec(name)
	tx.Commit()

	if err != nil {
		log.Println("Failed to add artist", name+":", err)
	}
}

func (canvas *Canvas) PutAlbum(album Album) {

	tx, err := canvas.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert or replace into albums (artist, name) values (?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(album.Artist, album.Name)
	if err != nil {
		log.Fatal(res, err)
	}
	tx.Commit()
}

func (canvas *Canvas) AddAlbum(artist, name string) {

	tx, err := canvas.db.Begin()
	stmt, err := tx.Prepare("insert or replace into albums (artist, name) values (?, ?)")
	defer stmt.Close()
	_, err = stmt.Exec(artist, name)
	tx.Commit()

	if err != nil {
		log.Fatal("Failed to add album", name, "by", artist+":", err)
	}
}

func (canvas *Canvas) AddSong(album, name, lyrics string) {

	tx, err := canvas.db.Begin()
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
