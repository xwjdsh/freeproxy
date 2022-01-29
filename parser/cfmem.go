package parser

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

var cfmemInstance = new(cfmemExecutor)

type cfmemExecutor struct{}

func (c *cfmemExecutor) Execute(ctx context.Context, linkchan chan<- string) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.cfmem.com/search/label/free", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("parser: [cfmem] http.Get error: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("parser: [cfmem] invalid error code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return fmt.Errorf("parser: [cfmem] goquery.NewDocumentFromReader error: %w", err)
	}

	wg := sync.WaitGroup{}
	doc.Find("article .entry-title a").Each(func(i int, s *goquery.Selection) {
		if post, ok := s.Attr("href"); ok {
			wg.Add(1)

			go func() {
				defer wg.Done()
				c.parsePage(ctx, post, linkchan)
			}()
		}
	})

	wg.Wait()
	return nil
}

func (c *cfmemExecutor) parsePage(ctx context.Context, post string, linkchan chan<- string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, post, nil)
	if err != nil {
		zap.L().Debug("parser: [cfmem] http.NewRequestWithContext error", zap.Error(err), zap.String("post", post))
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.L().Debug("parser: [cfmem] get post error", zap.Error(err), zap.String("post", post))
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		zap.L().Debug("parser: [cfmem] get post error", zap.Int("statusCode", res.StatusCode), zap.String("post", post))
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		zap.L().Debug("parser: [cfmem] goquery.NewDocumentFromReader error", zap.Error(err), zap.String("post", post))
		return
	}

	doc.Find("pre span[role=presentation]").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if text != "" {
			linkchan <- text
		}
	})
}