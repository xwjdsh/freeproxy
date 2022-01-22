package proxypool

import (
	"fmt"
	"sync"

	"github.com/xwjdsh/proxypool/crawler"
	"github.com/xwjdsh/proxypool/validator"
)

type Handler struct {
	crawler   *crawler.Handler
	validator *validator.Validator
}

func New() *Handler {
	return &Handler{
		crawler:   crawler.New(),
		validator: validator.New(),
	}
}

func (h *Handler) Start() {
	crawlerChan := make(chan *crawler.Proxy)
	validatorChan := make(chan *validator.Result)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		h.crawler.Crawl(crawlerChan)
		close(crawlerChan)
	}()

	go func() {
		defer wg.Done()
		h.validator.Validate(crawlerChan, validatorChan)
		close(validatorChan)
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case result, ok := <-validatorChan:
				if !ok {
					return
				}
				if result.Available {
					fmt.Printf("%#v, location:%s, toatl: %s\n", result.Proxy, result.Location, result.Total)
				}
			}
		}
	}()

	wg.Wait()

}
