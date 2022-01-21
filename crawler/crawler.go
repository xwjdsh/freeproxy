package crawler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Proxy struct {
	Type     string
	Source   string
	Host     string
	Port     int
	Location string
}

type Result struct {
	Proxies []*Proxy
	Err     error
}

type Crawler interface {
	Fetch() *Result
}

type Handler struct {
	crawlers []Crawler
}

func New() *Handler {
	return &Handler{
		crawlers: []Crawler{daili66Instance},
	}
}

func (h *Handler) Crawl(ctx context.Context) *Result {
	for _, c := range h.crawlers {
		c.Fetch()
	}

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
