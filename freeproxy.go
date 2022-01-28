package freeproxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/exporter"
	"github.com/xwjdsh/freeproxy/parser"
	"github.com/xwjdsh/freeproxy/proxy"
	"github.com/xwjdsh/freeproxy/storage"
	"github.com/xwjdsh/freeproxy/validator"
)

type Handler struct {
	cfg       *config.Config
	parser    *parser.Parser
	validator *validator.Validator
	storage   *storage.Handler
	exporter  *exporter.Exporter
}

func Init(cfg *config.Config) (*Handler, error) {
	h, err := storage.Init(cfg.Storage)
	if err != nil {
		return nil, err
	}
	return &Handler{
		cfg:       cfg,
		parser:    parser.New(cfg.Parser),
		validator: validator.New(cfg.Validator),
		storage:   h,
		exporter:  exporter.New(cfg.Exporter),
	}, nil
}

func (h *Handler) Start(ctx context.Context) {
	parserResultChan := make(chan *parser.Result)
	validatorResultChan := make(chan proxy.Proxy)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		wg.Done()
		h.parser.Parse(ctx, parserResultChan)
		close(parserResultChan)
	}()

	go func() {
		defer wg.Done()

		validatorWG := sync.WaitGroup{}
		validatorWG.Add(h.cfg.ValidatorCount)
		for i := 0; i < h.cfg.ValidatorCount; i++ {
			go func() {
				defer validatorWG.Done()
				for {
					select {
					case r, ok := <-parserResultChan:
						if !ok {
							return
						}
						if r.Err != nil {
							continue
						}

						vr, err := h.validator.Validate(r.Proxy)
						if err != nil {
							continue
						}
						b := vr.Proxy.GetBase()
						b.CountryCode, b.Country, b.CountryEmoji = vr.CountryCode, vr.Country, vr.CountryEmoji
						b.Delay = vr.Delay
						validatorResultChan <- vr.Proxy
					}
				}
			}()
		}
		validatorWG.Wait()
		close(validatorResultChan)
	}()

	go func() {
		defer wg.Done()

		storageWG := sync.WaitGroup{}
		storageWG.Add(h.cfg.StorageCount)
		for i := 0; i < h.cfg.StorageCount; i++ {
			go func() {
				defer storageWG.Done()
				for {
					select {
					case p, ok := <-validatorResultChan:
						if !ok {
							return
						}
						pp, err := h.storage.Store(context.Background(), p)
						if err != nil {
							fmt.Println(err)
						} else {
							fmt.Printf("save %s\n", pp.Server)
						}
					}
				}
			}()
		}
		storageWG.Wait()
	}()

	wg.Wait()
}

func (h *Handler) Export(ctx context.Context, tfp string) (string, error) {
	ps, err := h.storage.GetProxies(ctx)
	if err != nil {
		return "", nil
	}

	return h.exporter.Export(ps, tfp)
}
