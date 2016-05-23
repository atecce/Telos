#!/usr/bin/env python
#
# I should not like my writing to spare other people the trouble of thinking.
# But, if possible, to stimulate someone to thoughts of their own.
#

import scrapy
from db import canvas

class lyrics_net(scrapy.Spider):

    name = "lyrics_net"

    start_urls = [

        "http://www.lyrics.net/"
    ]

    canvas = canvas(name)

    def parse(self, response):

        for suburl in response.xpath("//div[@id='page-letter-search']//@href").re("^/artists/[A-Z0]$"): 

            url = response.urljoin(suburl+'/99999')

            yield scrapy.Request(url, callback=self.parse_letter)

    def parse_letter(self, response):

        for item in response.xpath("//tr//strong"):

		print item.xpath("//a/text()").extract()
		print item.xpath("//@href").extract()

#        for artist in response.xpath("//tr//a/text()").extract(): 
#            
#            self.canvas.add_artist(artist)
#
#        for suburl in response.xpath("//tr//@href").extract(): 
#            
#            url = response.urljoin(suburl)
#
#            yield scrapy.Request(url, callback=self.parse_artist)
#
#    def parse_artist(self, response):
#
#        for item in response.xpath("//div[@class='clearfix']//h3//a"): 
#
#		print item.xpath("text()").extract()
#		print item.xpath("@href").extract()
#
#                    # set the album information
#                    album_title = item.h3.a.text
#                    album_url   = urljoin(self.url, item.h3.a.get('href'))
#
#                    if self.verbose:
#
#                        print '\t', album_title
#                        print '\t', album_url
#                        print
#
#                    # add the album to the canvas
#                    self.canvas.add_album(artist_name, album_title)
#
#                    # get the soup
#                    album_soup = self.communicate(album_url)
#
#                    # handle Dorothy (which do not return the proper status code)
#                    if album_soup.find_all('body', {'id': 's4-page-homepage'}): 
#
#                        # extract the song data
#                        song_data = ((trace.a.text, urljoin(self.url, trace.a.get('href'))) \
#                                      for trace in item.find_all('tr') 	   		    \
#                                                if trace.a)
#
#                    # otherwise
#                    else:
#
#                        # extract the song data
#                        song_data = ((song_tag.a.text, urljoin(self.url, song_tag.a.get('href'))) \
#                                      for song_tag in album_soup.find_all('strong') 	   	  \
#                                                   if song_tag.a)
#
#                    # for each song
#                    for song_title, song_url in song_data:
#
#                        # fork
#                        self.multitask('songs', self.meditate, song_title, (album_title, song_title, song_url,))
#
#    def meditate(self, album_title, song_title, song_url):
#
#        # make some soup
#        song_soup = self.communicate(song_url)
#
#        # sometimes there's nothing to meditate on
#        try: 
#
#            lyrics = song_soup.find_all('pre', {'id': 'lyric-body-text'})[0].text
#
#            if self.verbose:
#
#                for line in lyrics.splitlines(): print '\t\t\t', line
#
#                print
#
#        except IndexError: return
#
#        # add song to canvas
#        self.canvas.add_song(album_title, song_title, lyrics)
#
#if __name__ == '__main__':
#
#    # declare parser
#    parser  = argparse.ArgumentParser()
#
#    # add arguments
#    start   = parser.add_argument("-s", "--start",   help="specify the start character",  default='0')
#    branch  = parser.add_argument("-b", "--branch",  help="specify the branching factor", default=2, 
#                                 type=int)
#    verbose = parser.add_argument("-v", "--verbose", help="specify the verbosity",        default=False,
#                                 action='store_true')
#
#    # parse arguments
#    args = parser.parse_args()
#
#    # start the investigation
#    investigation = lyrics_net(args.start, args.branch, args.verbose)
#    investigation.investigate()
#
#    # shut down instance after finished
#    system("sudo shutdown -h now")
