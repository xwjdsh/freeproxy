package parser

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/log"
	"github.com/xwjdsh/freeproxy/proxy"
)

type Result struct {
	Proxy proxy.Proxy
	Err   error
}

type Executor interface {
	Execute(ctx context.Context, linkchan chan<- *linkResp) error
	Name() string
}

type Handler struct {
	executors map[string]Executor
	cfg       *config.ParserConfig
}

type linkResp struct {
	Source string
	Link   string
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
		case "freefq":
			h.executors[e.Name] = freefqInstance
		case "feedburner":
			h.executors[e.Name] = feedburnerInstance
		default:
			return nil, fmt.Errorf("parser: invalid executor name: %s", e.Name)
		}
	}

	return h, nil
}

func (h *Handler) Parse(ctx context.Context, ch chan<- *Result) {
	wg := sync.WaitGroup{}
	wg.Add(len(h.executors))
	linkChan := make(chan *linkResp)

	for _, e := range h.executors {
		e := e
		go func() {
			defer func() {
				wg.Done()
				log.L().Debug("parser: executor end", zap.String("name", e.Name()))
			}()
			log.L().Debug("parser: executor start", zap.String("name", e.Name()))
			if err := e.Execute(ctx, linkChan); err != nil {
				log.L().Debug("parser: executor error", zap.Error(err))
			}
		}()
	}

	go func() {
		wg.Wait()
		close(linkChan)
	}()

	for {
		select {
		case lr, ok := <-linkChan:
			if !ok {
				return
			}
			r := new(Result)
			r.Proxy, r.Err = proxy.NewProxyByLink(lr.Link)
			if errors.Is(r.Err, proxy.ErrInvalidLink) {
				continue
			}

			if r.Err == nil {
				r.Proxy.GetBase().Source = lr.Source
			}

			ch <- r
		case <-ctx.Done():
			return
		}
	}
}
