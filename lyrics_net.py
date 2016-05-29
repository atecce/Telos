#!/usr/bin/env python
#
# I should not like my writing to spare other people the trouble of thinking.
# But, if possible, to stimulate someone to thoughts of their own.
#

import scrapy
from db import canvas

class lyrics_net(scrapy.Spider):

    # spider attributes
    name = "lyrics_net"
    start_urls = ["http://www.lyrics.net/"]
    handle_httpstatus_list = [302]

    # database
    canvas = canvas(name)

    def parse(self, response):

        # go through letters
        for suburl in response.xpath("//div[@id='page-letter-search']//@href").re("^/artists/[A-Z0]$"): 
            url = response.urljoin(suburl+'/99999')
            yield scrapy.Request(url, callback=self.parse_letter)

    def parse_letter(self, response):

        # go through artists
        for suburl in response.xpath("//tr//@href").extract(): 
            url = response.urljoin(suburl)
            request = scrapy.Request(url, callback=self.parse_artist)
            request.meta['artist_url'] = url # pass artist url down
            yield request

    def parse_artist(self, response):

        # get artist name
        artist_name = response.xpath("//div[@id='content-body']//h3//strong/text()").extract_first()
        if artist_name == None: return # artist entry without any content
        self.canvas.add_artist(artist_name) # add to db

        # go through albums
        for item in response.xpath("//div[@class='clearfix']//h3//a"): 
            album_title = item.xpath("text()").extract_first()
            album_url = response.urljoin(item.xpath("@href").extract_first())
            self.canvas.add_album(artist_name, album_title) # add to db
            request = scrapy.Request(album_url, callback=self.parse_album)
            request.meta['artist_url'] = response.meta['artist_url'] # pass artist url down
            request.meta['album_title'] = album_title # pass album title down
            yield request
                
    def parse_album(self, response):

        # TODO handle redirects
        if response.status == 302: 
            
            # set artist url to backtrack
            artist_url = response.request.headers.get('Referer')
            request = scrapy.Request(artist_url, callback=self.handle_dorothy, dont_filter=True)
            request.meta['album_title'] = response.meta['album_title']

        else: 

            # go through the songs
            for item in response.xpath("//strong//a"): 
                song_url = response.urljoin(item.xpath("@href").extract_first())
                request = scrapy.Request(song_url, callback=self.parse_song)
                request.meta['album_title'] = response.meta['album_title'] # pass album title down

        yield request

    def handle_dorothy(self, response):

        print
        print response.meta['album_title']
        print

        for item in response.xpath("//tr"): print item

#        album_items = response.xpath("//div[@class='clearfix']") #//h3[@class='artist-album-label']//a/text()").extract():
#
#        print album_items.xpath("//h3[@class='artist-album-label']//a")
#
#        for album_item in album_items.xpath("//h3[@class='artist-album-label']//a"):
#
#            print album_item
#
#            if response.meta['album_title'] == album_title:
#
#                print album_title 
#                print
#
#                for song_item in item.xpath("//tr"): print song_item
#
#                print

        print

    def parse_song(self, response):

        # get song info
        song_title = response.xpath("//h2[@id='lyric-title-text']/text()").extract_first()
        lyrics = response.xpath("//pre[@id='lyric-body-text']/text()").extract_first()
        self.canvas.add_song(response.meta['album_title'], song_title, lyrics) # add to db

#                    # handle Dorothy (which do not return the proper status code)
#                    if album_soup.find_all('body', {'id': 's4-page-homepage'}): 
#
#                        # extract the song data
#                        song_data = ((trace.a.text, urljoin(self.url, trace.a.get('href'))) \
#                                      for trace in item.find_all('tr') 	   	     \
#                                                if trace.a)
#
#                    # otherwise
#                    else:
#
#                        # extract the song data
#                        song_data = ((song_tag.a.text, urljoin(self.url, song_tag.a.get('href'))) \
#                                      for song_tag in album_soup.find_all('strong') 	   	   \
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
