import sqlite3

class canvas:

	def __init__(self, name):

		# set the canvas
		self.name = name
		canvas = sqlite3.connect(self.name)
		brush  = canvas.cursor()

		try: 

			brush.execute("create database ?", (self.name,))
			canvas.close()

		except: canvas.close()

		# sketch the outline
		canvas, brush = self.prepare()

		brush.execute("""create table if not exists artists (
									  
					name varchar(255) not null, 	  
								  
					primary key (name))""")

		brush.execute("""create table if not exists albums ( 		
										
					title 	    varchar(255) not null, 			
					artist_name varchar(255) not null,  		
										
					primary key (title, artist_name), 				
					foreign key (artist_name) references artists (name))""")

		brush.execute("""create table if not exists songs ( 	    	       
									       
					title 	    varchar(255) not null, 	    	       
					album_title varchar(255) not null, 	    	       
					lyrics 	    text, 			    	       
									       
					primary key (album_title, title),
					foreign key (album_title) references albums (title))""")

		canvas.close()

	def prepare(self):

		# 
		canvas = sqlite3.connect(self.name)
		brush  = canvas.cursor()

		return canvas, brush

	def add_artist(self, artist_name):

		canvas, brush = self.prepare()

		brush.execute("""insert or replace into artists (name) values (?)""", (artist_name,))

		canvas.commit()
		canvas.close()

	def add_album(self, artist_name, album_title):

		canvas, brush = self.prepare()

		brush.execute("""insert or replace into albums (artist_name, title) values (?, ?)""", (artist_name, album_title))

		canvas.commit()
		canvas.close()

	def add_song(self, album_title, song_title, lyrics):

		canvas, brush = self.prepare()

		brush.execute("""insert or replace into songs (album_title, title, lyrics) values (?, ?, ?)""", (album_title, song_title, lyrics))

		canvas.commit()
		canvas.close()

	# TODO 
	def get_artists(self):

		canvas, brush = self.prepare()

		brush.execute("select name from artists")

		artists = [item[0] for item in brush.fetchall()]

		canvas.close()

		return artists
