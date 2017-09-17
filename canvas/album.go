package canvas

import (
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Album struct {
	Artist *Artist

	Url  string
	Name string
}

func initAlbums() {
	_, err := db.Exec(`create table if not exists albums (

				     artist	 text not null,

				     name        text not null,

				     primary key (name, artist),
				     foreign key (artist) references artists (name))`)
	if err != nil {
		panic(err)
	}

}

func PutAlbum(album Album) {

	tx, err := db.Begin()
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
