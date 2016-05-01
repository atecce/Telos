import MySQLdb
from nltk.tokenize import word_tokenize, sent_tokenize

class reader:

	def prepare(self):

		# connect to database
		conn   = MySQLdb.connect('localhost', 'root', db='canvas', charset='utf8')
		cursor = conn.cursor()

		# return the connection
		return conn, cursor

	def artists(self):

		# get the connection
		conn, cursor = self.prepare()

		# execute the query
		cursor.execute("select name from artists")
		artists = (item[0] for item in cursor.fetchall())

		# close the connection
		conn.close()

		return artists

	def albums(self, artist=None):

		# get the connection
		conn, cursor = self.prepare()

		# set the query
		query = "select title from albums"

		if artist: query += " where artist_name=\"%s\";" % artist

		# execute the query
		cursor.execute(query)
		albums = (item[0] for item in cursor.fetchall())

		# close the connection
		conn.close()

		return albums

	def songs(self, artist=None, album=None):

		# get the connection
		conn, cursor = self.prepare()

		# set the query TODO (what happens with duplicate album titles?)
		query = "select title from songs"

		if   album:  query += " where album_title=\"%s\";" % album
		elif artist: query += " where album_title in (select title from albums where artist_name=\"%s\");" % artist

		# execute the query
		cursor.execute(query)
		songs = (item[0] for item in cursor.fetchall())

		# close the connection
		conn.close()

		return songs

	def lyrics(self, artist=None, album=None, song=None):

		# get the connection
		conn, cursor = self.prepare()

		# set the query TODO (what happens with duplicate album titles?)
		query = "select lyrics from songs"

		if   song:   query += " where title=\"%s\";" % song
		elif album:  query += " where album_title=\"%s\";" % album
		elif artist: query += " where album_title in (select title from albums where artist_name=\"%s\");" % artist

		# execute the query
		cursor.execute(query)
		text = (item[0] for item in cursor.fetchall())

		# close the connection
		conn.close()

		return text

	def words(self, artist=None, album=None, song=None):

		# get the connection
		conn, cursor = self.prepare()

		# initialize list of words
		word_list = list()

		# tokenize all the text
		for text in self.lyrics(artist, album, song): word_list += word_tokenize(text)

		# close the connection
		conn.close()

		return word_list

	def sents(self, artist=None, album=None, song=None):

		# get the connection
		conn, cursor = self.prepare()

		# initialize list of sentences
		sent_list = list()

		# tokenize all the text
		for text in self.lyrics(artist, album, song): sent_list += sent_tokenize(text)

		# close the connection
		conn.close()

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
