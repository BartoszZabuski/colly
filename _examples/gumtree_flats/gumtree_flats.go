package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

const (
	prefix = "https://www.gumtree.pl"
)

func main() {
	fName := "cryptocoinmarketcap.csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	// writer.Write([]string{"Name", "Symbol", "Price (USD)", "Volume (USD)", "Market capacity (USD)", "Change (1h)", "Change (24h)", "Change (7d)"})

	// Instantiate default collector
	// c := colly.NewCollector()

	toBeParsed := make(chan string, 1000)
	listingsReader := newListingReader(file)
	urlFilter := NewUrlFilter(toBeParsed)

	// get listing details
	listingsReader.OnHTML(".vip-header-and-details ", func(e *colly.HTMLElement) {
		fmt.Println(strings.TrimRight(e.ChildText(".vip-content-header .price .value"), " zł"))
		fmt.Println(e.ChildText(".vip-details .description span"))
		// data dodania
		// fmt.Println(strings.TrimSpace(e.ChildText(".selMenu li .attribute .value")))
	})

	// get listing detail pages
	listingsReader.OnHTML(".result-link .container .title", func(e *colly.HTMLElement) {
		urlFilter.Add(e.ChildAttr("a", "href"))
	})

	// get listing pages
	listingsReader.OnHTML(".pagination .after", func(e *colly.HTMLElement) {
		urlFilter.Add(e.ChildAttr("a", "href"))
	})

	listingsReader.Visit(prefix + "/s-mieszkania-i-domy-sprzedam-i-kupie/wroclaw/v1c9073l3200114p1?df=ownr&nr=10")

	log.Printf("Scraping finished, check file %q for results\n", fName)

	var waitTime = 0
loop:
	for {
		select {
		case url := <-toBeParsed:
			fmt.Println("url: ", prefix+url)
			listingsReader.Visit(prefix + url)

		default:
			time.Sleep(1000 * time.Millisecond)
			waitTime++
			if waitTime > 4 {
				break loop
			}
		}
	}
	fmt.Println("done waiting")
	log.Printf("number of urls %d", urlFilter.GetUrlsNumber())

}

type listingsReader struct {
	*colly.Collector
	file *os.File
}

func newListingReader(file *os.File) *listingsReader {
	return &listingsReader{colly.NewCollector(), file}
}

type urlFilter struct {
	urlMap         map[string]struct{}
	processingChan chan string
	mx             sync.RWMutex
}

func NewUrlFilter(processingChan chan string) *urlFilter {
	return &urlFilter{
		urlMap:         make(map[string]struct{}),
		processingChan: processingChan,
	}
}

func (filter *urlFilter) Add(url string) {
	if url == "" {
		return
	}
	filter.mx.Lock()
	defer filter.mx.Unlock()
	if _, ok := filter.urlMap[url]; !ok {
		filter.urlMap[url] = struct{}{}
		filter.processingChan <- url
	}
}

func (filter *urlFilter) GetUrlsNumber() int {
	filter.mx.RLock()
	defer filter.mx.RUnlock()
	return len(filter.urlMap)

}
