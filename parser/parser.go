package parser

import (
	"bufio"
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
	ch  chan *Result
	cfg *config.ParserConfig
}

func New(cfg *config.ParserConfig) *Parser {
	return &Parser{
		ch:  make(chan *Result),
		cfg: cfg,
	}
}

func (p *Parser) Parse() {
	list, err := ioutil.ReadDir(p.cfg.Dir)
	if err != nil {
		log.Fatalf("parser: ioutil.ReadDir error: %v", err)
	}

	wg := sync.WaitGroup{}
	for _, item := range list {
		if item.IsDir() {
			continue
		}
		fp := filepath.Join(p.cfg.Dir, item.Name())
		wg.Add(1)
		go func() {
			defer wg.Done()

			file, err := os.Open(fp)
			if err != nil {
				log.Fatalf("parser: os.Open error: %v, file: %s", err, fp)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "ss://") {
					r := new(Result)
					r.Proxy, r.Err = proxy.NewShadowsocksByLink(line)
					p.ch <- r
				}
			}
			if err := scanner.Err(); err != nil {
				log.Fatalf("parser: scanner.Err: %v", err)
			}
		}()
	}
	wg.Wait()
	p.ch <- nil
}

func (p *Parser) Chan() <-chan *Result {
	return p.ch
}
