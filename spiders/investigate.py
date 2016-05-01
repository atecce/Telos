#!/usr/bin/env python
#
# I should not like my writing to spare other people the trouble of thinking.
# But, if possible, to stimulate someone to thoughts of their own.
#

import scrapy

class investigation(scrapy.Spider):

	name = 'investigation'

	start_urls = [

		"http://www.lyrics.net/"
	]

	def parse(self, response):

		for suburl in response.xpath("//div[@id='page-letter-search']//@href").re("^/artists/[A-Z0]$"): 

			print suburl

			yield scrapy.Request(response.urljoin(suburl), self.honor)

#		# extract alphabet urls
#		alphabet_urls = (self.url + re.match('^/artists/[A-Z0]$', link.get('href')).group(0) + '/99999' \
#				 for link in soup.find_all('div', {'id': 'page-letter-search'})[0]      	\
#				          if re.match('^/artists/[A-Z0]$', str(link.get('href'))))
#
#		# extract artist tags
#		artist_tags = (trace.strong.a 		  		     		  \
#			       for alphabet_url in alphabet_urls 	     	 	  \
#			       for trace in self.communicate(alphabet_url).find_all('tr') \
#			       		 if trace.strong)
#
#		# get artist data
#		artist_data = ((artist_tag.text, self.url + artist_tag.get('href')) for artist_tag in artist_tags)
#
#		# for each artist
#		for artist_name, artist_url in artist_data: 
#				
#			print artist_name
#			print
#
#			if artist_name == last_artist: caught_up = True
#
#			if not caught_up: continue
#
#			self.multitask('artists', self.honor, artist_name, (artist_name, artist_url,))

	def honor(self, response):

		for suburl in response.xpath("//tr//@href").extract(): 

			print '\t', suburl

			yield scrapy.Request(response.urljoin(suburl), self.admire)

#	def honor(self, artist_name, artist_url):
#
#		# get the soup
#		artist_soup = self.communicate(artist_url)
#
#		# add the artist to the canvas
#		canvas.add_artist(artist_name)
#
#		# get the album labels
#		album_labels = artist_soup.find_all('h3', {'class': 'artist-album-label'})
#
#		# for each album label
#		for album_label in album_labels:
#
#			# extract the data
#			album_title = album_label.a.text
#			album_url   = self.url + album_label.a.get('href')
#
#			print '\t', album_title
#			print
#
#			self.multitask('albums', self.experience, album_title, (artist_name, artist_url, album_title, album_url,))

	def admire(self, response):

		for suburl in response.xpath("//h3[@class='artist-album-label']//@href").extract(): 

			print '\t\t', suburl

			yield scrapy.Request(response.urljoin(suburl), self.experience)

#	def experience(self, artist_name, artist_url, album_title, album_url):
#
#		# get the soup
#		album_soup = self.communicate(album_url)
#
#		# TODO handle the redirects
#		if album_soup.find_all('body', {'id': 's4-page-homepage'}): 
#
#			# keep track of the redirects
#			with open('home_page_albums.txt', 'a') as f: f.write(artist_name.encode('utf-8') + ', ' + album_title.encode('utf-8') +'\n')
#
#			return
#
#		# add the album to the canvas
#		canvas.add_album(artist_name, album_title)
#
#		# extract the song data
#		song_data = ((song_tag.a.text, self.url + song_tag.a.get('href'))  \
#			     for song_tag in album_soup.find_all('strong') 	   \
#			     		  if song_tag.a)
#
#		# for each song
#		for song_title, song_url in song_data:
#
#			print '\t\t', song_title
#			print
#
#			self.multitask('songs', self.meditate, song_title, (album_title, song_title, song_url,))

	def experience(self, response):

		for suburl in response.xpath("//strong//@href").extract(): 

			print '\t\t\t', suburl

			yield scrapy.Request(response.urljoin(suburl), self.meditate)

#	def meditate(self, album_title, song_title, song_url):
#
#		# make some soup
#		song_soup = self.communicate(song_url)
#
#		# TODO handle redirects
#		if song_soup.find_all('body', {'id': 's4-page-homepage'}): 
#			
#			# track redirects
#			with open('home_page_songs.txt','a') as f: 
#
#				f.write(album_title.encode('utf-8') + ', ' + song_title.encode('utf-8') + '\n')
#
#			return
#
#		# sometimes there's nothing to meditate on
#		try: 
#
#			lyrics = song_soup.find_all('pre', {'id': 'lyric-body-text'})[0].text
#
#			for line in lyrics.splitlines(): print '\t\t\t', line
#
#			print
#
#		except IndexError: return
#
#		# add song to canvas
#		canvas.add_song(album_title, song_title, lyrics)

	def meditate(self, response):

		for lyrics in response.xpath("//pre[@id='lyric-body-text']/text()").extract():

			for line in lyrics.splitlines():

				print '\t\t\t\t', line
