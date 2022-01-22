package crawler

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type daili66 struct {
	page int
}

var daili66Instance = &daili66{
	page: 10,
}

func (daili66) name() string {
	return "66ip"
}

func (c daili66) fetch(ch chan *Proxy) {
	wg := sync.WaitGroup{}
	wg.Add(c.page)
	for i := 1; i <= c.page; i++ {
		i := i
		go func() {
			defer wg.Done()
			url := fmt.Sprintf("http://www.66ip.cn/%d.html", i)
			doc, err := fetch(url, c.name())
			if err != nil {
				return
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

				ch <- &Proxy{
					Source:   c.name(),
					Protocol: "http",
					Host:     tds.First().Text(),
					Port:     port,
				}
			})

		}()
	}
	wg.Wait()
}
