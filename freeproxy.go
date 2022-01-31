package freeproxy

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/exporter"
	"github.com/xwjdsh/freeproxy/internal/progressbar"
	"github.com/xwjdsh/freeproxy/log"
	"github.com/xwjdsh/freeproxy/parser"
	"github.com/xwjdsh/freeproxy/storage"
	"github.com/xwjdsh/freeproxy/validator"
)

type Handler struct {
	cfg       *config.Config
	parser    *parser.Handler
	validator *validator.Validator
	storage   *storage.Handler
	exporter  *exporter.Exporter
}

func Init(cfg *config.Config) (*Handler, error) {
	log.Init(cfg.Log)

	h, err := storage.Init(cfg.Storage)
	if err != nil {
		return nil, err
	}
	p, err := parser.Init(cfg.Parser)
	if err != nil {
		return nil, err
	}
	return &Handler{
		cfg:       cfg,
		parser:    p,
		validator: validator.New(cfg.Validator),
		storage:   h,
		exporter:  exporter.New(cfg.Exporter),
	}, nil
}

func (h *Handler) Tidy(ctx context.Context) error {
	ps, err := h.storage.GetProxies(ctx)
	if err != nil {
		return err
	}

	if err := h.validator.CheckNetwork(ctx); err != nil {
		return err
	}

	g := new(errgroup.Group)
	pb := progressbar.New(len(ps))
	for _, p := range ps {
		p := p
		g.Go(func() error {
			defer func() {
				pb.SetSuffix("done: %d", p.ID)
				pb.Incr()
			}()
			pp, err := p.Restore(p.Config)
			if err != nil {
				return err
			}
			if err := h.validator.Validate(ctx, pp); err != nil {
				return h.storage.Remove(ctx, p.ID)
			}

			_, err = h.storage.Store(ctx, pp)
			return err
		})
	}

	pb.Wait()
	return g.Wait()
}

func (h *Handler) Fetch(ctx context.Context) error {
	parserResultChan := make(chan *parser.Result)
	g := new(errgroup.Group)

	g.Go(func() error {
		h.parser.Parse(ctx, parserResultChan)
		close(parserResultChan)
		return nil
	})

	g.Go(func() error {
		ng := new(errgroup.Group)
		for {
			select {
			case r, ok := <-parserResultChan:
				if !ok {
					return ng.Wait()
				}
				if r.Err != nil {
					continue
				}

				ng.Go(func() error {
					if err := h.validator.Validate(ctx, r.Proxy); err != nil {
						return nil
					}
					_, err := h.storage.Store(ctx, r.Proxy)
					return err
				})
			}
		}
	})

	return g.Wait()
}

func (h *Handler) Export(ctx context.Context) error {
	ps, err := h.storage.GetProxies(ctx)
	if err != nil {
		return nil
	}

	return h.exporter.Export(ps)
}
