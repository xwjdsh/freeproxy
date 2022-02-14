package parser

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/log"
	"github.com/xwjdsh/freeproxy/proxy"
)

type Result struct {
	Source     string
	SourceDone bool
	Proxy      proxy.Proxy
	Err        error
}

type Executor interface {
	Execute(ctx context.Context, linkchan chan<- *linkResp) error
	Name() string
}

type executorAndConfig struct {
	Executor
	cfg *config.ParserExecutor
}

type Handler struct {
	executors map[string]*executorAndConfig
	cfg       *config.ParserConfig
}

type linkResp struct {
	Source string
	Link   string
}

var executorsMap = map[string]Executor{}

func init() {
	for _, e := range []Executor{cfmemInstance, freefqSSInstance, freefqSSRInstance, freefqVmessInstance, feedburnerInstance} {
		executorsMap[e.Name()] = e
	}
}

func Init(cfg *config.ParserConfig) (*Handler, error) {
	h := &Handler{
		cfg:       cfg,
		executors: map[string]*executorAndConfig{},
	}
	for _, e := range cfg.Executors {
		if h.executors[e.Name] != nil {
			return nil, fmt.Errorf("parser: registered executor: %s", e.Name)
		}

		executor, ok := executorsMap[e.Name]
		if !ok {
			return nil, fmt.Errorf("parser: invalid executor name: %s", e.Name)
		}
		h.executors[e.Name] = &executorAndConfig{
			Executor: executor,
			cfg:      e,
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
			ctx, cancel := context.WithTimeout(ctx, e.cfg.Timeout)
			defer func() {
				cancel()
				wg.Done()
				log.L().Debug("parser: executor end", zap.String("name", e.Name()))
			}()

			log.L().Debug("parser: executor start", zap.String("name", e.Name()))
			r := &Result{Source: e.Name(), SourceDone: true}
			r.Err = e.Execute(ctx, linkChan)
			if r.Err != nil {
				log.L().Debug("parser: executor error", zap.Error(r.Err))
			}
			ch <- r
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
			r := &Result{Source: lr.Source}
			r.Proxy, r.Err = proxy.NewProxyByLink(lr.Link)
			if r.Err != nil {
				continue
			}

			r.Proxy.GetBase().Source = lr.Source
			ch <- r
		case <-ctx.Done():
			return
		}
	}
}
