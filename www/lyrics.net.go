package www

import (
	"log"
	"net/url"
	"path"
	"sync"

	"github.com/kr/pretty"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"

	"github.com/de-nova-stella/investigations/canvas"
	"github.com/de-nova-stella/rest"
)

// TODO shared ref
var (
	domain *url.URL
	wg     *sync.WaitGroup
)

func Init() {

	domain, _ = url.Parse("http://www.lyrics.net")
	wg = new(sync.WaitGroup)
}

func Run() {

	for _, c := range "0ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		u := *domain
		u.Path = path.Join("artists", string(c), "99999")
		parseArtists(u)
	}
}

func parseArtists(u url.URL) {

	// set body
	b, ok := rest.Get(u.String())
	if !ok {
		pretty.Logln("[DEBUG] failed getting artist url", u)
		return
	}
	defer b.Close()

	// parse page
	root, err := html.Parse(b)
	if err != nil {
		log.Fatal(err)
	}

	// find artist urls
	for _, link := range getArtists(root) {
		if link.FirstChild != nil {
			u.Path = scrape.Attr(link, "href")
			artist := &canvas.Artist{
				Url:  u.String(),
				Name: link.FirstChild.Data,
			}
			artist.Parse(wg)
		}
	}
}

func getArtists(root *html.Node) []*html.Node {
	return scrape.FindAll(root, func(n *html.Node) bool {
		if n.Parent != nil {
			return n.Parent.Data == "strong" && n.Data == "a"
		}
		return false
	})
}
