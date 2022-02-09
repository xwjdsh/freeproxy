package parser

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type feedburnerData struct {
	XMLName xml.Name `xml:"feed"`
	Text    string   `xml:",chardata"`
	Entry   []struct {
		Text      string `xml:",chardata"`
		ID        string `xml:"id"`
		Published string `xml:"published"`
		Updated   string `xml:"updated"`
		Category  struct {
			Text   string `xml:",chardata"`
			Scheme string `xml:"scheme,attr"`
			Term   string `xml:"term,attr"`
		} `xml:"category"`
		Title struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"title"`
		Content struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"content"`
		Link []struct {
			Text  string `xml:",chardata"`
			Rel   string `xml:"rel,attr"`
			Type  string `xml:"type,attr"`
			Href  string `xml:"href,attr"`
			Title string `xml:"title,attr"`
		} `xml:"link"`
		Author struct {
			Text  string `xml:",chardata"`
			Name  string `xml:"name"`
			Email string `xml:"email"`
			Image struct {
				Text   string `xml:",chardata"`
				Rel    string `xml:"rel,attr"`
				Width  string `xml:"width,attr"`
				Height string `xml:"height,attr"`
				Src    string `xml:"src,attr"`
			} `xml:"image"`
		} `xml:"author"`
		Thumbnail struct {
			Text   string `xml:",chardata"`
			Media  string `xml:"media,attr"`
			URL    string `xml:"url,attr"`
			Height string `xml:"height,attr"`
			Width  string `xml:"width,attr"`
		} `xml:"thumbnail"`
		Total string `xml:"total"`
	} `xml:"entry"`
}

var feedburnerInstance = &feedburnerExecutor{}

type feedburnerExecutor struct{}

func (c *feedburnerExecutor) Name() string {
	return "feedburner"
}

func (c *feedburnerExecutor) Execute(ctx context.Context, linkChan chan<- *linkResp) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://feeds.feedburner.com/mattkaydiary", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("parser: [%s] http.Get error: %w", c.Name(), err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("parser: [%s] invalid error code: %d", c.Name(), res.StatusCode)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fd := &feedburnerData{}
	if err := xml.Unmarshal(data, fd); err != nil {
		return err
	}

	if len(fd.Entry) == 0 {
		return fmt.Errorf("parser: [%s] xml no entries", c.Name())
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fd.Entry[0].Content.Text))
	if err != nil {
		return fmt.Errorf("parser: [%s] goquery.NewDocumentFromReader error: %w", c.Name(), err)
	}

	doc.Find("div[style='-webkit-text-stroke-width: 0px;']").Children().Eq(4).Children().Last().Find("div").Each(func(i int, s *goquery.Selection) {
		if text := s.Text(); text != "" {
			linkChan <- &linkResp{
				Source: c.Name(),
				Link:   text,
			}
		}
	})
	return nil
}
