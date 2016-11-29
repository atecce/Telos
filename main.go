package main

import (
	"log"
	"net/http"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

func main() {

	// get initial page
	resp, err := http.Get("https://en.wikipedia.org/wiki/Machine_learning")
	if err != nil {
		panic(err)
	}

	// get root
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

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
					log.Println("\t", scrape.Attr(link, "href"))
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
