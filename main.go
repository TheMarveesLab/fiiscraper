package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"runtime"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

const (
	nomeSelector     = "#carbon_fields_fiis_header-2 > div > div > div.headerTicker__content > div.headerTicker__content__name > h1"
	dySelector       = "#carbon_fields_fiis_header-2 > div > div > div.headerTicker__content > div.headerTicker__content__info > div > div:nth-child(1) > p:nth-child(1) > b"
	pvpSelector      = "#carbon_fields_fiis_header-2 > div > div > div.headerTicker__content > div.headerTicker__content__info > div > div:nth-child(4) > p:nth-child(1) > b"
	segmentoSelector = "#carbon_fields_fiis_informations-2 > div.moreInfo.wrapper > p:nth-child(6) > b"
)

type Resource struct {
	Tickers []Ticker `json:"tickers"`
}

type Ticker struct {
	Nome     string `json:"nome"`
	DY       string `json:"dy"`
	PVP      string `json:"pvp"`
	Segmento string `json:"segmento"`
}

func main() {
	content, err := fetch("https://fiis.com.br/lista-de-fundos-imobiliarios/")
	if err != nil {
		fmt.Println(err)
		return
	}

	doc, err := goquery.NewDocumentFromReader(content)
	if err != nil {
		fmt.Println(err)
	}

	urls := []string{}
	doc.Find(".tickerBox__link_ticker").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("href")
		if exists {
			urls = append(urls, url)
		}
	})

	workers := runtime.NumCPU()
	chunks := int(math.Ceil(float64(len(urls)) / float64(workers)))

	var wg sync.WaitGroup
	var mu sync.Mutex

	tickers := []Ticker{}
	for i := 0; i < len(urls); i += chunks {
		end := i + chunks
		if end > len(urls) {
			end = len(urls)
		}

		chunk := urls[i:end]

		// worker
		wg.Add(1)
		go func(urls []string) {
			defer wg.Done()
			for _, url := range urls {
				ticker, _ := fetchTicker(url)
				if ticker != nil {
					mu.Lock()
					tickers = append(tickers, *ticker)
					mu.Unlock()
				}
			}
		}(chunk)
	}

	wg.Wait()

	jsonContent, err := json.Marshal(Resource{Tickers: tickers})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(jsonContent))
}

func fetchTicker(url string) (*Ticker, error) {
	content, err := fetch(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(content)
	if err != nil {
		return nil, err
	}

	ticker := &Ticker{
		Nome:     doc.Find(nomeSelector).Text(),
		DY:       doc.Find(dySelector).Text(),
		PVP:      doc.Find(pvpSelector).Text(),
		Segmento: doc.Find(segmentoSelector).Text(),
	}

	return ticker, nil
}

func fetch(URL string) (io.Reader, error) {
	res, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, err
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(content), nil
}
