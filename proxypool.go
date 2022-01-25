package proxypool

import (
	"fmt"
	"sync"

	"github.com/xwjdsh/proxypool/parser"
	"github.com/xwjdsh/proxypool/validator"
)

type Handler struct {
	parser    *parser.Parser
	validator *validator.Validator
}

func New() *Handler {
	parser := parser.New("./files")
	return &Handler{
		parser:    parser,
		validator: validator.New(parser.Chan()),
	}
}

func (h *Handler) Start() {
	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		h.parser.Parse()
	}()

	go func() {
		defer wg.Done()
		h.validator.Validate()
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case r := <-h.validator.Chan():
				if r == nil {
					return
				}
				fmt.Printf("%+v\n", r.Proxy.GetBase())
			}
		}
	}()

	wg.Wait()

}
