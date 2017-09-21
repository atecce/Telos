package www

import (
	"log"
	"net/url"
	"path"
	"sync"
	"unicode"

	"github.com/kr/pretty"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"

	"github.com/de-nova-stella/investigations/canvas"
	"github.com/de-nova-stella/rest"
)

const domain = "http://www.lyrics.net"

// set regular expression for letter suburls
var artists = regexp.MustCompile("^artist/.*$")

type Investigator struct {
	expression string

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

func New(start string) *Investigator {

	investigator := new(Investigator)
	canvas.Init()
	investigator.wg = new(sync.WaitGroup)

	latest, err := canvas.FetchLatestArtist()
	if err != nil {
		log.Fatal("failed to fetch latest artist")
	}
	pretty.Logln("got latest artist", latest)

	first := rune(latest.Name[0])
	pretty.Logln("got first letter", first)

	if inAlphabet(first) {
		investigator.expression = "^/artists/[" + string(unicode.ToUpper(first)) + "-Z]$"
	} else {
		investigator.expression = "^/artists/[0A-Z]$"
	}
	pretty.Logln("set regex", investigator.expression)

	return investigator
}

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
