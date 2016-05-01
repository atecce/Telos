#!/usr/bin/env python

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

	def honor(self, response):

		for suburl in response.xpath("//tr//@href").extract(): 

			print '\t', suburl

			yield scrapy.Request(response.urljoin(suburl), self.admire)

	def admire(self, response):

		for suburl in response.xpath("//h3[@class='artist-album-label']//@href").extract(): 

			print '\t\t', suburl

			yield scrapy.Request(response.urljoin(suburl), self.experience)

	def experience(self, response):

		for suburl in response.xpath("//strong//@href").extract(): 

			print '\t\t\t', suburl

			yield scrapy.Request(response.urljoin(suburl), self.meditate)

	def meditate(self, response):

		for lyrics in response.xpath("//pre[@id='lyric-body-text']/text()").extract():

			for line in lyrics.splitlines():

				print '\t\t\t\t', line
