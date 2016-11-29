package main

import (
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

type NavPage struct {
	root *html.Node
	wg   sync.WaitGroup
}

func (p *NavPage) investigate() {

	// find all pages with content
	contentPages := scrape.FindAll(p.root, func(n *html.Node) bool {
		if n.Parent != nil {
			return n.Parent.Data == "ul" &&
				scrape.Attr(n.Parent, "class") == "mw-allpages-chunk"
		}
		return false
	})

	// concurrently investigate all the pages
	for _, page := range contentPages {
		p.wg.Add(1)
		go func(page *html.Node) {
			defer p.wg.Done()
			if page.FirstChild != nil {

				url, err := url.Parse(scrape.Attr(page.FirstChild, "href"))
				if err != nil {
					panic(err)
				}

				root := communicate(url)
				content := scrape.FindAll(root, func(n *html.Node) bool {
					return n.Data == "p"
				})

				for _, node := range content {
					if node.FirstChild != nil {
						log.Println("\t", node.FirstChild.Data)
					}
				}
			}
		}(page)
	}

	// wait for them to wrap up
	p.wg.Wait()
}

func (p *NavPage) next() *NavPage {

	// attempt to find element with next page link
	nextPage, ok := scrape.Find(p.root, func(n *html.Node) bool {
		if n.Parent != nil {
			return n.Parent.Data == "div" &&
				scrape.Attr(n.Parent, "class") == "mw-allpages-nav" &&
				n == n.Parent.LastChild
		}
		return false
	})

	// return the link if found otherwise return nil
	if ok {
		url, err := url.Parse(scrape.Attr(nextPage, "href"))
		if err != nil {
			panic(err)
		}
		return newNavPage(url)
	}
	return nil
}

func communicate(url *url.URL) *html.Node {

	// fix sub urls
	if url.Scheme == "" {
		url.Scheme = "https"
	}
	if url.Host == "" {
		url.Host = "en.wikipedia.org"
	}
	log.Println(url)

	// make request
	resp, err := http.Get(url.String())
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}

	// return the root of DOM
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}
	return root
}

func newNavPage(url *url.URL) *NavPage {

	// construct new page
	p := new(NavPage)
	p.root = communicate(url)
	return p
}

func main() {

	// initialize url
	initialURL, err := url.Parse("https://en.wikipedia.org/w/index.php?title=Special:AllPages")
	if err != nil {
		panic(err)
	}

	// singly linked list
	for p := newNavPage(initialURL); p != nil; p = p.next() {
		p.investigate()
	}
}
