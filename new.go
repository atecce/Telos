package main

type job interface {
	scrape()
	exec()
}

func main() {

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
