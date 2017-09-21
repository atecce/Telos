package www

import (
	"log"
	"net/url"
	"path"
	"regexp"
	"sync"
	"unicode"

	"github.com/kr/pretty"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"

	"github.com/de-nova-stella/investigations/canvas"
	"github.com/de-nova-stella/rest"
)

type Investigator struct {
	domain *url.URL

	// TODO shared ref
	wg *sync.WaitGroup
}

func inAlphabet(char rune) bool {
	for _, c := range "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		if c == char {
			return true
		}
	}
	return false
}

func New() *Investigator {

	investigator := new(Investigator)
	investigator.wg = new(sync.WaitGroup)
	investigator.domain, _ = url.Parse("http://www.lyrics.net")

	return investigator
}

func (investigator *Investigator) Run() {

	latest, err := canvas.FetchLatestArtist()
	if err != nil {
		log.Fatal("failed to fetch latest artist")
	}
	pretty.Logln("got latest artist", latest)

	first := rune(latest.Name[0])
	pretty.Logln("got first letter", first)

	var pattern string
	if inAlphabet(first) {
		pattern = "^[" + string(unicode.ToUpper(first)) + "-Z]$"
	} else {
		pattern = "^[0A-Z]$"
	}
	pretty.Logln("set pattern", pattern)

	for _, c := range "0ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		if ok, err := regexp.MatchString(pattern, string(c)); ok {
			u := investigator.domain
			u.Path = path.Join("artists", string(c), "99999")
			investigator.parseArtists(*u)
		} else if err != nil {
			log.Println("error matching alphabet pattern")
		}
	}
}

func (investigator *Investigator) parseArtists(u url.URL) {

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
	for _, link := range scrapeArtists(root) {
		if link.FirstChild != nil {
			u.Path = scrape.Attr(link, "href")
			artist := &canvas.Artist{
				Url:  u.String(),
				Name: link.FirstChild.Data,
			}
			artist.Parse(investigator.wg)
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
