package parser

import (
	"bufio"
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/proxy"
)

type Result struct {
	Proxy proxy.Proxy
	Err   error
}

type Parser struct {
	cfg *config.ParserConfig
}

func New(cfg *config.ParserConfig) *Parser {
	return &Parser{
		cfg: cfg,
	}
}

func (p *Parser) Parse(ctx context.Context, ch chan<- *Result) {
	list, err := ioutil.ReadDir(p.cfg.Dir)
	if err != nil {
		log.Fatalf("parser: ioutil.ReadDir error: %v", err)
	}

	wg := sync.WaitGroup{}
	for _, item := range list {
		if item.IsDir() {
			continue
		}
		wg.Add(1)
		fp := filepath.Join(p.cfg.Dir, item.Name())

		go func() {
			defer wg.Done()
			file, err := os.Open(fp)
			if err != nil {
				log.Fatalf("parser: os.Open error: %v, file: %s", err, fp)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				if ctx.Err() != nil {
					return
				}
				line := scanner.Text()
				r := new(Result)
				switch {
				case strings.HasPrefix(line, "ss://"):
					r.Proxy, r.Err = proxy.NewShadowsocksByLink(line)
				default:
					r = nil
				}
				if r != nil {
					ch <- r
				}
			}

			if err := scanner.Err(); err != nil {
				log.Fatalf("parser: scanner.Err: %v", err)
			}
		}()
	}

	wg.Wait()
}
