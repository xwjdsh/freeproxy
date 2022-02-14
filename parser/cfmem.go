package parser

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

var cfmemInstance = new(cfmemExecutor)

type cfmemExecutor struct{}

func (c *cfmemExecutor) Name() string {
	return "cfmem"
}

func (c *cfmemExecutor) Execute(ctx context.Context, linkchan chan<- *linkResp) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.cfmem.com/search/label/free", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	if postLink, ok := doc.Find("article .entry-title a").First().Attr("href"); ok {
		return c.parsePage(ctx, postLink, linkchan)
	}

	return nil
}

func (c *cfmemExecutor) parsePage(ctx context.Context, post string, linkchan chan<- *linkResp) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, post, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	doc.Find("pre span[role=presentation]").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		linkchan <- &linkResp{
			Source: c.Name(),
			Link:   text,
		}
	})

	return nil
}
