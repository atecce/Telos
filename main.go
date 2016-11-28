package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

const (
	rootURL    = "https://en.wikipedia.org"
	initialURL = "/w/index.php?title=Special:AllPages&hideredirects=1"
)

type Page struct {
	root *html.Node
	wg   sync.WaitGroup
}

func (p *Page) investigate() {

	// find all pages with content
	contentPages := scrape.FindAll(p.root, func(n *html.Node) bool {
		return n.Data == "a"
	})

	// concurrently investigate all the pages
	for _, page := range contentPages {
		p.wg.Add(1)
		go func(page *html.Node) {
			if page.FirstChild != nil {
				_, _ = http.Get(rootURL + scrape.Attr(page, "href"))
				log.Println("\t", scrape.Attr(page, "href"), page.FirstChild.Data)
			}
			p.wg.Done()
		}(page)
	}

	// wait for them to wrap up
	p.wg.Wait()
}

func (p *Page) next() *Page {

	// attempt to find element with next page link
	nextPage, ok := scrape.Find(p.root, func(n *html.Node) bool {
		if n.FirstChild != nil {
			return strings.HasPrefix(n.FirstChild.Data, "Next page")
		}
		return false
	})

	// return the link if found otherwise return nil
	if ok {
		log.Println(scrape.Attr(nextPage, "href"))
		return newPage(scrape.Attr(nextPage, "href"))
	}
	return nil
}

func newPage(subURL string) *Page {

	// make request
	resp, err := http.Get(rootURL + subURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// construct new page
	p := new(Page)
	p.root, err = html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}
	return p
}

func main() {

	// singly linked list
	for p := newPage(initialURL); p != nil; p = p.next() {
		p.investigate()
	}
}
