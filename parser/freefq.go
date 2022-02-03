package parser

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/sync/errgroup"

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
	g := new(errgroup.Group)
	for _, url := range []string{"https://freefq.com/free-ss/", "https://freefq.com/free-ssr/", "https://freefq.com/v2ray/"} {
		url := url
		g.Go(func() error {
			return c.parse(ctx, url, linkChan)
		})
	}

	return g.Wait()
}

func (c *freefqExecutor) parse(ctx context.Context, url string, linkChan chan<- *linkResp) error {
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
			Source: c.Name(),
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
