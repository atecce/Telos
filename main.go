package main

import (
	"github.com/de-nova-stella/investigations/www"
)

type job interface {
	scrape()
	exec()
}

func main() {

	wittgenstein := www.New()
	wittgenstein.Run()

	const MAX_QUEUE = 256

	queue := make(chan payload, MAX_QUEUE)

	for {
		select {
		case job := <-queue:
			job.scrape()
			job.exec()
		}
	}
}
