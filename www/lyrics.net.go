package www

import (
	"fmt"
	"log"
	"log/syslog"
	"net/url"
	"path"
	"sync"

	"golang.org/x/net/html"

	"github.com/de-nova-stella/investigations/canvas"
	"github.com/de-nova-stella/rest"
	"github.com/yhat/scrape"
)

// TODO shared ref
var (
	domain *url.URL
	wg     *sync.WaitGroup
	logger *syslog.Writer
)

func Init() {

	domain, _ = url.Parse("http://www.lyrics.net")
	wg = new(sync.WaitGroup)
	syslogger, err := syslog.Dial("", "", syslog.LOG_USER, "investigations")
	if err != nil {
		log.Fatal("syslog:", err)
	}
	logger = syslogger
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
		logger.Debug(fmt.Sprintf("failed getting artist url %v", u))
		return
	}
	defer b.Close()

	// parse page
	root, err := html.Parse(b)
	if err != nil {
		logger.Emerg(err.Error())
	}

	// find artist urls
	for _, link := range scrapeArtists(root) {
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

func scrapeArtists(root *html.Node) []*html.Node {
	return scrape.FindAll(root, func(n *html.Node) bool {
		if n.Parent != nil {
			return n.Parent.Data == "strong" && n.Data == "a"
		}
		return false
	})
}
