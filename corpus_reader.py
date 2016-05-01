import sqlite3
from nltk.tokenize import word_tokenize, sent_tokenize

class reader:

	def prepare(self):

		# connect to database
		con = sqlite3.connect("lyrics_net")
		cur = con.cursor()

		# return the connection
		return con, cur

	def artists(self):

		# get the connection
		con, cur = self.prepare()

		# execute the query
		cur.execute("select name from artists")
		artists = (item[0] for item in cur.fetchall())

		# close the connection
		con.close()

		return artists

	def albums(self, artist=None):

		# get the connection
		con, cur = self.prepare()

		# set the query
		query = "select title from albums"

		if artist: query += " where artist_name=\"%s\";" % artist

		# execute the query
		cur.execute(query)
		albums = (item[0] for item in cur.fetchall())

		# close the connection
		con.close()

		return albums

	def songs(self, artist=None, album=None):

		# get the connection
		con, cur = self.prepare()

		# set the query TODO exploits here
		query = "select title from songs"

		if   album:  query += " where album_title=\"%s\";" % album
		elif artist: query += " where album_title in (select title from albums where artist_name=\"%s\");" % artist

		# execute the query
		cur.execute(query)
		songs = (item[0] for item in cur.fetchall())

		# close the connection
		con.close()

		return songs

	def lyrics(self, artist=None, album=None, song=None):

		# get the connection
		con, cur = self.prepare()

		# set the query TODO exploits here
		query = "select lyrics from songs"

		if   song:   query += " where title=\"%s\";" % song
		elif album:  query += " where album_title=\"%s\";" % album
		elif artist: query += " where album_title in (select title from albums where artist_name=\"%s\");" % artist

		# execute the query
		cur.execute(query)
		text = (item[0] for item in cur.fetchall())

		# close the connection
		con.close()

		return text

	def words(self, artist=None, album=None, song=None):

		# get the connection
		con, cur = self.prepare()

		# initialize list of words
		word_list = list()

		# tokenize all the text
		for text in self.lyrics(artist, album, song): word_list += word_tokenize(text)

		# close the connection
		con.close()

		return word_list

	def sents(self, artist=None, album=None, song=None):

		# get the connection
		con, cur = self.prepare()

		# initialize list of sentences
		sent_list = list()

		# tokenize all the text
		for text in self.lyrics(artist, album, song): sent_list += sent_tokenize(text)

		# close the connection
		con.close()

		return sent_list

	def paras(self): pass
	def tagged_words(self): pass
	def tagged_sents(self): pass
	def tagged_paras(self): pass
	def chunked_sents(self): pass
	def parsed_sents(self): pass
	def parsed_paras(self): pass
	def xml(self): pass
	def raw(self): pass

test = reader()

print test.words()
