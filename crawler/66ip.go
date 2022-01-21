package crawler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

type daili66 struct {
	page int
}

var daili66Instance = &daili66{
	page: 3,
}

func (daili66) Name() string {
	return "66ip"
}

func (c daili66) Fetch(ctx context.Context) (*Result, error) {
	errGroup, ctx := errgroup.WithContext(ctx)
	result := &Result{}
	for i := 1; i <= c.page; i++ {

		errGroup.Go()
		url := fmt.Sprintf("http://www.66ip.cn/%d.html", i)
		doc, err := fetch(url, c.Name())
		if err != nil {
			return result, err
		}
		doc.Find(".containerbox table tr").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				return
			}
			tds := s.Children()
			port, err := strconv.Atoi(tds.Eq(1).Text())
			if err != nil {
				return
			}

			proxy := &Proxy{
				Source:   c.Name(),
				Type:     "http",
				Host:     tds.First().Text(),
				Port:     port,
				Location: tds.Eq(2).Text(),
			}
			result.Proxies = append(result.Proxies, proxy)
		})
	}

	return result, nil
}
