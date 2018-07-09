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
		description := e.ChildText(".vip-details .description span")
		district := getDistrictFor(description)
		writer.Write([]string{
			strings.TrimSpace(strings.TrimRight(e.ChildText(".vip-content-header .price .value"), " zł")),
			// e.ChildText(".vip-details .description span"),
			district,
		})

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

func getDistrictFor(text string) string {
	for key, value := range districtsMap {
		if strings.Contains(strings.ToLower(text), key) {
			return value
		}
	}
	fmt.Printf("no matching districts in '%s'", text)
	return "-"
}

var districtsMap = map[string]string{
	"stare miasto":            "stare miasto",
	"starym mieście":          "stare miasto",
	"przedmieście świdnickie": "przedmieście świdnickie",
	"przedmieściu świdnickim": "przedmieście świdnickie",
	"szczepin":                "szczepin",
	"szczepinie":              "szczepin",
	"śródmieście":             "śródmieście",
	"śródmieściu":             "śródmieście",
	"bartoszowice":            "bartoszowice",
	"bartoszowicach":          "bartoszowice",
	"biskupin":                "biskupin",
	"biskupine":               "biskupin",
	"dąbie":                   "dąbie",
	"dąbiu":                   "dąbie",
	"nadodrze":                "nadodrze",
	"nadodrzu":                "nadodrze",
	"ołbin":                   "ołbin",
	"ołbinie":                 "ołbin",
	"plac grunwaldzki":        "plac grunwaldzki",
	"placu grunwaldzkim":      "plac grunwaldzki",
	"sępolno":                 "sępolno",
	"sępolnie":                "sępolno",
	"zacisze":                 "zacisze",
	"zaciszu":                 "zacisze",
	"zalesie":                 "zalesie",
	"zalesiu":                 "zalesie",
	"szczytniki":              "szczytniki",
	"szczytnikach":            "szczytniki",
	"krzyki":                  "krzyki",
	"krzykach":                "krzyki",
	"bieńkowice":              "bieńkowice",
	"bieńkowicach":            "bieńkowice",
	"bierdzany":               "bierdzany",
	"bierdzanach":             "bierdzany",
	"borek":                   "borek",
	"borku":                   "borek",
	"brochów":                 "brochów",
	"brochowie":               "brochów",
	"dworek":                  "dworek",
	"dworku":                  "dworek",
	"gaj":                     "gaj",
	"gaju":                    "gaj",
	"glinianki":               "glinianki",
	"gliniankach":             "glinianki",
	"huby":                    "huby",
	"hubach":                  "huby",
	"jagodno":                 "jagodno",
	"jagodnie":                "jagodno",
	"klecina":                 "klecina",
	"klecinie":                "klecina",
	"księże małe":             "księże małe",
	"księżu małym":            "księże małe",
	"księże wielkie":          "księże wielkie",
	"księżu wielkim":          "księże wielkie",
	"lamowice stare":          "lamowice stare",
	"lamowicach starych":      "lamowice stare",
	"nowy dom":                "nowy dom",
	"nowym domu":              "nowy dom",
	"ołtaszyn":                "ołtaszyn",
	"ołtaszynie":              "ołtaszyn",
	"opatowice":               "opatowice",
	"opatowicach":             "opatowice",
	"partynice":               "partynice",
	"partynicach":             "partynice",
	"południe":                "południe",
	"południu":                "południe",
	"przedmieście oławskie":   "przedmieście oławskie",
	"przedmieściu oławskim":   "przedmieście oławskie",
	"rakowiec":                "rakowiec",
	"rakowcu":                 "rakowiec",
	"siedlec":                 "siedlec",
	"siedlcach":               "siedlec",
	"świątniki":               "świątniki",
	"świątnikach":             "świątniki",
	"tarnogaj":                "tarnogaj",
	"tarnogaju":               "tarnogaj",
	"wilczy kąt":              "wilczy kąt",
	"wilczym kącie":           "wilczy kąt",
	"wojszyce":                "wojszyce",
	"wojszycach":              "wojszyce",
	"psie pole":               "psie pole",
	"psim polu":               "psie pole",
	"karłowice":               "karłowice",
	"karłowicach":             "karłowice",
	"kleczków":                "kleczków",
	"kleczkowie":              "kleczków",
	"kłokoczyce":              "kłokoczyce",
	"kłokoczycach":            "kłokoczyce",
	"kowale":                  "kowale",
	"kowalach":                "kowale",
	"lesica":                  "lesica",
	"lesicy":                  "lesica",
	"ligota":                  "ligota",
	"ligocie":                 "ligota",
	"lipa piotrowska":         "lipa piotrowska",
	"lipie piotrowskiej":      "lipa piotrowska",
	"mirowiec":                "mirowiec",
	"mirowcu":                 "mirowiec",
	"osobowice":               "osobowice",
	"osobowicach":             "osobowice",
	"pawłowice":               "pawłowice",
	"pawłowicach":             "pawłowice",
	"polanka":                 "polanka",
	"polance":                 "polanka",
	"polanowicach":            "polanowicach",
	"poświętne":               "polanowicach",
	"poświętnach":             "poświętnach",
	"pracze widawskie":        "pracze widawskie",
	"praczach widawskich":     "pracze widawskie",
	"rędzin":                  "rędzin",
	"rędzinie":                "rędzin",
	"różanka":                 "różanka",
	"różance":                 "różanka",
	"sołtysowice":             "sołtysowice",
	"sołtysowicach":           "sołtysowice",
	"strachocin":              "strachocin",
	"strachocinie":            "strachocin",
	"swojczyce":               "swojczyce",
	"swojczycach":             "swojczyce",
	"świniary":                "świniary",
	"świniarach":              "świniary",
	"widawa":                  "widawa",
	"widawach":                "widawach",
	"wojnów":                  "wojnów",
	"wojnowie":                "wojnów",
	"zakrzów":                 "zakrzów",
	"zakrzowie":               "zakrzów",
	"zgorzelisko":             "zgorzelisko",
	"zgorzelisku":             "zgorzelisko",
	"fabryczna":               "fabryczna",
	"fabrycznej":              "fabryczna",
	"gajowice":                "gajowice",
	"gajowicach":              "gajowice",
	"gądów mały":              "gądów mały",
	"gądowie małym":           "gądów mały",
	"grabiszyn":               "grabiszyn",
	"grabiszynie":             "grabiszyn",
	"grabiszynek":             "grabiszynek",
	"grabiszynku":             "grabiszynek",
	"janówek":                 "janówek",
	"janówku":                 "janówek",
	"jarnołtów":               "jarnołtów",
	"jarnołtowie":             "jarnołtów",
	"jerzmanowo":              "jerzmanowo",
	"jerzmanowie":             "jerzmanowo",
	"kozanów":                 "kozanów",
	"kozanowie":               "kozanów",
	"kuźniki":                 "kuźniki",
	"kuźnikiach":              "kuźniki",
	"leśnica":                 "leśnica",
	"leśnicy":                 "leśnica",
	"marszowice":              "marszowice",
	"marszowicach":            "marszowicach",
	"maślice":                 "maślice",
	"maślicach":               "maślice",
	"mokra":                   "mokra",
	"mokrej":                  "mokra",
	"muchobór mały":           "muchobór mały",
	"muchoborze małym":        "muchobór mały",
	"muchobór wielki":         "muchobór wielki",
	"muchoborze wielkim":      "muchobór wielki",
	"nowa karczma":            "nowa karczma",
	"nowej karczmie":          "nowa karczma",
	"nowe domy":               "nowe domy",
	"nowych domach":           "nowe domy",
	"nowy dwór":               "nowy dwór",
	"nowym dworze":            "nowy dwór",
	"oporów":                  "oporów",
	"oporowie":                "oporów",
	"pilczyce":                "pilczyce",
	"pilczycach":              "pilczyce",
	"popowice":                "popowice",
	"popowicach":              "popowice",
	"pracze odrzańskie":       "pracze odrzańskie",
	"praczach odrzańskich":    "pracze odrzańskie",
	"pustki":                  "pustki",
	"pustkach":                "pustki",
	"ratyń":                   "ratyń",
	"ratyniu":                 "ratyń",
	"stabłowice":              "stabłowice",
	"stabłowicach":            "stabłowice",
	"stabłowice nowe":         "stabłowice nowe",
	"stabłowicach nowych":     "stabłowice nowe",
	"strachowice":             "strachowice",
	"strachowicach":           "strachowice",
	"osiniec":                 "osiniec",
	"osiniecu":                "osiniec",
	"złotniki":                "złotniki",
	"złotnikiach":             "złotniki",
	"żar":                     "żar",
	"żarach":                  "żar",
	"żerniki":                 "żerniki",
	"kosmonautów":             "kosmonautów",
}
