package main

import "log"

func main() {

	crawlers := make([]Crawler, 0, 5)

	crawlers = append(crawlers, newPropertyPalCrawler())
	crawlers = append(crawlers, newPropertyIeCrawler())

	for _, c := range crawlers {
		defer c.close()
	}

	var checks int
	var matches []string
	for _, c := range crawlers {
		c, m, err := c.run()
		if err != nil {
			log.Fatal(err)
		}

		checks += c
		matches = append(matches, m...)
	}

	log.Println("+++++++++++++++++++++++++++++++++++++++++++++++")
	log.Println("Finished Crawling!")
	log.Printf("Number of items crawled: %d\n", checks)
	log.Printf("Number of matches: %d\n", len(matches))
	log.Println()

	for _, v := range matches {
		log.Println(v)
	}

}
