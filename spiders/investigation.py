import scrapy

def multitask(self, processes, process_target, process_args):

	# don't wait for your neighbors
	import multiprocessing

	# and focus
	branching_factor = 8

	# start
	process = multiprocessing.Process(target=process_target, args=process_args)
	process.start()

	# keep track of your progress
	processes.append(process)
	alive = [process for process in processes if process.is_alive()]

	# don't overextend
	while len(alive) >= branching_factor:

	# pace yourself
	alive = [process for process in processes if process.is_alive()]
	self.time.sleep(1)

class investigation(scrapy.Spider):

	name = 'investigation'

	start_urls = [

		"http://www.lyrics.net/"
	]

	def parse(self, response):

		artists = list()

		for suburl in response.xpath("//div[@id='page-letter-search']//@href").re("^/artists/[A-Z0]$"): 

			self.honor(response.urljoin(suburl))

	def honor(self, response):

		for suburl in response.xpath("//tr//@href").extract(): 

			self.admire(scrapy.Request(response.urljoin(suburl)))

	def admire(self, response):

		for suburl in response.xpath("//h3[@class='artist-album-label']//@href").extract(): 

			self.experience(scrapy.Request(response.urljoin(suburl)))

	def experience(self, response):

		for suburl in response.xpath("//strong//@href").extract(): 

			self.meditate((scrapy.Request(response.urljoin(suburl))))

	def meditate(self, response):

		for lyrics in response.xpath("//pre[@id='lyric-body-text']/text()").extract():

			for line in lyrics.splitlines():

				print line
