package parser

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var freefqInstance = &freefqExecutor{
	host: "https://freefq.com",
}

type freefqExecutor struct {
	host string
}

func (c *freefqExecutor) Name() string {
	return "freefq"
}

func (c *freefqExecutor) Execute(ctx context.Context, linkChan chan<- *linkResp) error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.parseSS(ctx, linkChan)
	}()

	wg.Wait()
	return nil
}

func (c *freefqExecutor) parseSS(ctx context.Context, linkChan chan<- *linkResp) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://freefq.com/free-ss/", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("parser: [freefq] http.Get error: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("parser: [freefq] invalid error code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return fmt.Errorf("parser: [freefq] goquery.NewDocumentFromReader error: %w", err)
	}

	pageLink, ok := doc.Find(".news_list table").Eq(1).Find("a").First().Attr("href")
	if !ok {
		return fmt.Errorf("page link not found")
	}
	pageLink = c.host + pageLink
	fileLink, err := c.getFileLinkByPageLink(ctx, pageLink)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return c.fetchFile(ctx, fileLink, linkChan)
}

func (c *freefqExecutor) fetchFile(ctx context.Context, fileLink string, linkChan chan<- *linkResp) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fileLink, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return err
	}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		text := scanner.Text()
		linkChan <- &linkResp{
			Source: c.Name(),
			Link:   text,
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (c *freefqExecutor) getFileLinkByPageLink(ctx context.Context, pageLink string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, pageLink, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("parser: [freefq] http request error: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("parser: [freefq] invalid error code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("parser: [freefq] goquery.NewDocumentFromReader error: %w", err)
	}

	fileLink, ok := doc.Find("fieldset a").Attr("href")
	if !ok {
		return "", fmt.Errorf("parser: [freefq] file link not found")
	}

	return fileLink, nil
}
