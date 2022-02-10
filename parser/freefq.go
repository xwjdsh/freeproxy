package parser

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	freefqSSInstance = &baseFreefqExecutor{
		name:    "freefq_ss",
		address: "https://freefq.com/free-ss/",
	}

	freefqSSRInstance = &baseFreefqExecutor{
		name:    "freefq_ssr",
		address: "https://freefq.com/free-ssr/",
	}

	freefqVmessInstance = &baseFreefqExecutor{
		name:    "freefq_v2ray",
		address: "https://freefq.com/v2ray/",
	}
)

type baseFreefqExecutor struct {
	address string
	name    string
}

func (c *baseFreefqExecutor) Name() string {
	return c.name
}

func (c *baseFreefqExecutor) Execute(ctx context.Context, linkChan chan<- *linkResp) error {
	return c.parse(ctx, c.address, linkChan)
}

func (c *baseFreefqExecutor) parse(ctx context.Context, url string, linkChan chan<- *linkResp) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	host := "https://freefq.com"
	pageLink = host + pageLink
	fileLink, err := c.getFileLinkByPageLink(ctx, pageLink)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return c.fetchFile(ctx, fileLink, linkChan)
}

func (c *baseFreefqExecutor) fetchFile(ctx context.Context, fileLink string, linkChan chan<- *linkResp) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, fileLink, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("parser: [freefq] http status code error: %d, url: %s", res.StatusCode, fileLink)
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(doc.Text()))
	for scanner.Scan() {
		text := scanner.Text()
		if link, ok := getSSRlink(text); ok {
			text = link
		} else if link, ok := getV2raylink(text); ok {
			text = link
		}

		linkChan <- &linkResp{
			Source: c.name,
			Link:   text,
		}
	}

	return scanner.Err()
}

func getSSRlink(line string) (string, bool) {
	result := regexp.MustCompile(`(?U)data="ssr://(?P<result>.+)"`).FindStringSubmatch(line)
	if len(result) != 2 {
		return line, false
	}
	return "ssr://" + result[1], true
}

func getV2raylink(line string) (string, bool) {
	result := regexp.MustCompile(`(?U)data="vmess://(?P<result>.+)"`).FindStringSubmatch(line)
	if len(result) != 2 {
		return line, false
	}
	return "vmess://" + result[1], true
}

func (c *baseFreefqExecutor) getFileLinkByPageLink(ctx context.Context, pageLink string) (string, error) {
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
