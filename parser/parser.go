package parser

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/proxy"
)

type Result struct {
	Proxy proxy.Proxy
	Err   error
}

type Executor interface {
	Execute(ctx context.Context, linkchan chan<- string) error
}

type Handler struct {
	executors map[string]Executor
	cfg       *config.ParserConfig
}

func Init(cfg *config.ParserConfig) (*Handler, error) {
	h := &Handler{
		cfg:       cfg,
		executors: map[string]Executor{},
	}
	for _, e := range cfg.Executors {
		if h.executors[e.Name] != nil {
			return nil, fmt.Errorf("parser: registered executor: %s", e.Name)
		}
		switch e.Name {
		case "cfmem":
			h.executors[e.Name] = cfmemInstance
		default:
			return nil, fmt.Errorf("parser: invalid executor name: %s", e.Name)
		}
	}

	return h, nil
}

func (h *Handler) Parse(ctx context.Context, ch chan<- *Result) {
	wg := sync.WaitGroup{}
	wg.Add(len(h.executors))
	linkChan := make(chan string)

	for _, e := range h.executors {
		e := e
		go func() {
			defer wg.Done()
			e.Execute(ctx, linkChan)
		}()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		wg.Wait()
		cancel()
	}()

	for {
		select {
		case link := <-linkChan:
			r := new(Result)
			switch {
			case strings.HasPrefix(link, "ss://"):
				r.Proxy, r.Err = proxy.NewShadowsocksByLink(link)
			default:
				continue
			}
			ch <- r
		case <-ctx.Done():
			return
		}
	}
}
