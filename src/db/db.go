package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

func InitiateDB(name string) *sql.DB {

	// prepare db
	canvas, err := sql.Open("sqlite3", name + ".db")

	// create tables
	_, err = canvas.Exec(`create table if not exists artists (
								  
				      name text not null, 	  
				              			  
				      primary key (name))`)

	_, err = canvas.Exec(`create table if not exists albums ( 		
									
				     title       text not null, 			
				     artist_name text not null,  		
				           					
				     primary key (title, artist_name), 				
				     foreign key (artist_name) references artists (name))`)

	_, err = canvas.Exec(`create table if not exists songs ( 	    	       
								       
				     title       text not null, 	    	       
				     album_title text not null, 	    	       
				     lyrics      text, 			    	       
				           				       
				     primary key (album_title, title),
				     foreign key (album_title) references albums (title))`)

	// catch error
	if err != nil { log.Println("Failed to create tables:", err) }

	return canvas
}

func AddArtist(artist_name string, canvas *sql.DB) {

	// prepare db
	tx, err := canvas.Begin()

	// insert entry
	stmt, err := tx.Prepare("insert or replace into artists (name) values (?)")
	defer stmt.Close()
	_, err = stmt.Exec(artist_name)
	tx.Commit()

	// catch error
	if err != nil { log.Println("Failed to add artist", artist_name + ":", err) }
}

func AddAlbum(artist_name, album_title string, canvas *sql.DB) {

	// prepare db
	tx, err := canvas.Begin()

	// insert entry
	stmt, err := tx.Prepare("insert or replace into albums (artist_name, title) values (?, ?)")
	defer stmt.Close()
	_, err = stmt.Exec(artist_name, album_title)
	tx.Commit()

	// catch error
	if err != nil { log.Println("Failed to add album", album_title, "by", artist_name+":", err) }
}

func AddSong(album_title, song_title, lyrics string, canvas *sql.DB) {

	// initialized failed flag
	var failed bool

	for {

		// prepare db
		tx, err := canvas.Begin()

		// catch error
		if err != nil {
			failed = true
			log.Println("Error in .Begin: Failed to add song", song_title, "in album", album_title+":", err)
			time.Sleep(time.Second)
			canvas.Close()
			continue
		}

		// prepare statement
		stmt, err := tx.Prepare("insert or replace into songs (album_title, title, lyrics) values (?, ?, ?)")

		// catch error
		if err != nil {
			failed = true
			log.Println("Error in .Prepare: Failed to add song", song_title, "in album", album_title+":", err)
			time.Sleep(time.Second)
			canvas.Close()
			continue
		}

		// close statement
		defer stmt.Close()

		// execute statement
		_, err = stmt.Exec(album_title, song_title, lyrics)

		// catch error
		if err != nil {
			failed = true
			log.Println("Error in .Exec: Failed to add song", song_title, "in album", album_title+":", err)
			time.Sleep(time.Second)
			canvas.Close()
			continue
		}

		// commit changes
		tx.Commit()

		// notify that a previous failure was cleaned up
		if failed { log.Println("Successfully added song", song_title, "in album", album_title) }

		// exit
		return
	}
}
