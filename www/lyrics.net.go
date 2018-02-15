package www

import (
	"fmt"
	"log"
	"log/syslog"
	"net/url"
	"path"
	"regexp"
	"sync"
	"unicode"

	"golang.org/x/net/html"

	"github.com/de-nova-stella/investigations/canvas"
	"github.com/de-nova-stella/rest"
	"github.com/yhat/scrape"
)

type Investigator struct {
	domain *url.URL
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
	investigator.domain, _ = url.Parse("http://www.lyrics.net")

	return investigator
}

func (investigator *Investigator) Run() {

	pattern := "^[0A-Z]$"

	latest, err := canvas.FetchLatestArtist()
	if err == nil {

		fmt.Printf("got latest artist %s\n", latest)

		first := rune(latest.Name[0])
		fmt.Printf("got first letter %s\n", first)

		if inAlphabet(first) {
			pattern = "^[" + string(unicode.ToUpper(first)) + "-Z]$"
		}
	}
	fmt.Printf("set pattern %s\n", pattern)

	for _, c := range "0ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		if ok, err := regexp.MatchString(pattern, string(c)); ok {
			u := investigator.domain
			u.Path = path.Join("artists", string(c), "99999")
			investigator.parseArtists(*u)
		} else if err != nil {
			fmt.Printf("error matching alphabet pattern\n")
		}
	}
}

func (investigator *Investigator) parseArtists(u url.URL) {

	// set body
	b, ok := rest.Get(u.String())
	if !ok {
		fmt.Printf("failed getting artist url %v\n", u)
		return
	}
	defer b.Close()

	// parse page
	root, err := html.Parse(b)
	if err != nil {
		fmt.Println(err.Error())
	}

	// find artist urls
	for _, link := range scrapeArtists(root) {
		if link.FirstChild != nil {
			u.Path = scrape.Attr(link, "href")
			artist := &canvas.Artist{
				Url:  u.String(),
				Name: link.FirstChild.Data,
			}
			artist.Parse()
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
