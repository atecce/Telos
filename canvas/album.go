package canvas

import (
	"log"

	"github.com/kr/pretty"
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

	tx := begin()
	if tx == nil {
		return
	}

	stmt, err := tx.Prepare("insert or replace into albums (artist, name) values (?, ?)")
	if err != nil {
		log.Println("[ERROR] preparing stmt for album", album)
		pretty.Logln("[DEBUG] on tx", tx, "with err", err)
	}

	_, err = stmt.Exec(album.Artist, album.Name)
	if err != nil {
		log.Println("[ERROR] execing stmt for album", album)
		pretty.Logln("[DEBUG] on stmt", stmt, "with err", err)
	}

	if err := tx.Commit(); err != nil {
		log.Println("[ERROR] committing tx for album", album)
		pretty.Logln("[DEBUG] on tx", tx, "with err", err)
	}

	if err := stmt.Close(); err != nil {
		log.Println("[ERROR] closing stmt for album", album)
		pretty.Logln("[DEBUG] on stmt", stmt, "with err", err)
	}
}
