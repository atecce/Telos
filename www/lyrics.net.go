package www

import (
	"log"
	"regexp"
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

func (investigator *Investigator) Run() {

	b, ok := rest.Get(domain)
	if !ok {
		return
	}
	defer b.Close()

	z := html.NewTokenizer(b)
	for {
		switch z.Next() {

		case html.ErrorToken:
			return

		case html.StartTagToken:
			if letterLink, ok := investigator.getLetterLink(z.Token()); ok {
				investigator.getArtists(letterLink)
			}
		}
	}
}

func (investigator *Investigator) getLetterLink(t html.Token) (string, bool) {

	letters, _ := regexp.Compile(investigator.expression)

	if t.Data == "a" {
		for _, a := range t.Attr {
			if a.Key == "href" && letters.MatchString(a.Val) {
				return domain + a.Val + "/99999", true
			}
		}
	}
	return "", false
}

func (investigator *Investigator) getArtists(letter_url string) {

	// set body
	b, ok := rest.Get(letter_url)
	if !ok {
		return
	}
	defer b.Close()

	// parse page
	root, err := html.Parse(b)
	if err != nil {
		log.Fatal(err)
	}

	// find artist urls
	for _, link := range getArtistLinks(root) {

		if artist_suburl := scrape.Attr(link, "href"); artists.MatchString(artist_suburl) {

			// extract artist name
			var artist_name string
			if link.FirstChild != nil {
				artist_name = link.FirstChild.Data
			}

			artist := &canvas.Artist{
				Url:  domain + "/" + artist_suburl,
				Name: artist_name,
			}

			// parse the artist
			artist.Parse(investigator.wg)
		}
	}
}

func getArtistLinks(root *html.Node) []*html.Node {
	return scrape.FindAll(root, func(n *html.Node) bool {
		if n.Parent != nil {
			return n.Parent.Data == "strong" && n.Data == "a"
		}
		return false
	})
}
