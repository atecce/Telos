package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

func communicate(rawURL string) *html.Node {
	resp, err := http.Get(rawURL)
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}
	return root
}

func main() {

	// get initial page
	root := communicate("https://en.wikipedia.org/wiki/Machine_learning")

	// get navigation heads
	navFrames := scrape.FindAll(root, matcher("NavFrame collapsed"))

	// get all class links
	for _, navFrame := range navFrames {
		log.Println(navFrame)
		for _, field := range []string{"NavHead", "NavContent"} {
			navHead, ok := scrape.Find(navFrame, matcher(field))
			if ok {
				links := scrape.FindAll(navHead, linkMatcher())
				for _, link := range links {

					// extract content
					url := url.URL{
						Scheme: "https",
						Host:   "en.wikipedia.org",
						Path:   scrape.Attr(link, "href"),
					}
					root := communicate(url.String())
					paragraphs := scrape.FindAll(root, func(n *html.Node) bool {
						return n.Data == "p"
					})
					var content string
					for _, paragraph := range paragraphs {
						for child := paragraph.FirstChild; child != nil; child = child.NextSibling {
							if child.Type == html.TextNode {
								content += child.Data
							}
						}
					}
					log.Println(content)

					log.Println("\t", url.String())
				}
			}
		}
		println()
	}
}

func linkMatcher() func(*html.Node) bool {
	return func(n *html.Node) bool {
		return n.Data == "a"
	}
}

func matcher(class string) func(*html.Node) bool {
	return func(n *html.Node) bool {
		return n.Data == "div" && scrape.Attr(n, "class") == class
	}
}
