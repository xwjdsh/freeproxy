package crawler

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type ProxyType int

const (
	Unknown ProxyType = iota
	HighAnonymous
	Transparent
)

type Proxy struct {
	Type     ProxyType
	Protocol string
	Source   string
	Host     string
	Port     int
}

type Result struct {
	Proxies []*Proxy
}

type Crawler interface {
	fetch(chan *Proxy)
}

type Handler struct {
	crawlers []Crawler
}

func New() *Handler {
	return &Handler{
		crawlers: []Crawler{daili66Instance, fatezeroInstance},
	}
}

func (h *Handler) Crawl(ch chan *Proxy) {
	wg := sync.WaitGroup{}
	wg.Add(len(h.crawlers))
	for _, c := range h.crawlers {
		c := c
		func() {
			defer wg.Done()
			c.fetch(ch)
		}()
	}
	wg.Wait()
}

func fetch(url, name string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("crawler: [%s] http request error: %w", name, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("crawler: [%s] unexpected response: %d %s", name, res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("crawler: [%s] goquery new document error: %w", name, err)
	}

	return doc, nil
}
