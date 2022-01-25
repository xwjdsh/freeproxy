package validator

import (
	"sync"

	"github.com/xwjdsh/proxypool/parser"
	"github.com/xwjdsh/proxypool/proxy"
)

type Result struct {
	Proxy     proxy.Proxy
	Available bool
}

type Validator struct {
	parserChan <-chan *parser.Result
	ch         chan *Result
}

func New(parserChan <-chan *parser.Result) *Validator {
	return &Validator{
		ch:         make(chan *Result),
		parserChan: parserChan,
	}
}

func (v *Validator) Validate() {
	wg := sync.WaitGroup{}
	for {
		select {
		case r := <-v.parserChan:
			if r == nil {
				wg.Wait()
				v.ch <- nil
				return
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				v.ch <- v.ValidateOne(r.Proxy)
			}()
		}
	}
}

func (v *Validator) ValidateOne(proxy proxy.Proxy) *Result {
	return &Result{Proxy: proxy}
}

func (v *Validator) Chan() <-chan *Result {
	return v.ch
}
