package canvas

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db, _ = sql.Open("sqlite3", "lyrics.net.db")

func Init() {
	initArtists()
	initAlbums()
	initSongs()
}
