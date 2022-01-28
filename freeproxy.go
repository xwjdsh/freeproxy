package freeproxy

import (
	"fmt"
	"sync"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/parser"
	"github.com/xwjdsh/freeproxy/validator"
)

type Handler struct {
	parser    *parser.Parser
	validator *validator.Validator
}

func New(cfg *config.Config) *Handler {
	parser := parser.New("./files")
	return &Handler{
		parser:    parser,
		validator: validator.New(parser.Chan(), cfg.Validator),
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
				if r.Error == nil {
					fmt.Printf("%s %s(%s) delay: %d\n", r.Proxy.GetBase().Server, r.Country, r.CountryEmoji, r.Delay)
				}
			}
		}
	}()

	wg.Wait()
}
