package parser

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/xwjdsh/freeproxy/proxy"
)

type Result struct {
	Proxy proxy.Proxy
	Err   error
}

type Parser struct {
	ch  chan *Result
	dir string
}

func New(dir string) *Parser {
	return &Parser{
		ch:  make(chan *Result),
		dir: dir,
	}
}

func (p *Parser) Parse() {
	list, err := ioutil.ReadDir(p.dir)
	if err != nil {
		log.Fatalf("parser: ioutil.ReadDir error: %v", err)
	}

	wg := sync.WaitGroup{}
	for _, item := range list {
		if item.IsDir() {
			continue
		}
		fp := filepath.Join(p.dir, item.Name())
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
					r.Proxy, r.Err = parseSSLink(line)
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

func parseSSLink(link string) (*proxy.Shadowsocks, error) {
	uri, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	cipher := ""
	password := ""
	if uri.User.String() == "" {
		infos, err := base64Decode(uri.Hostname())
		if err != nil {
			return nil, err
		}
		uri, err = url.Parse("ss://" + infos)
		if err != nil {
			return nil, err
		}
		cipher = uri.User.Username()
		password, _ = uri.User.Password()
	} else {
		cipherInfoString, err := base64Decode(uri.User.Username())
		if err != nil {
			return nil, fmt.Errorf("base64Decode error: %w", err)
		}
		cipherInfo := strings.SplitN(cipherInfoString, ":", 2)
		if len(cipherInfo) < 2 {
			return nil, fmt.Errorf("password parse error")
		}
		cipher = strings.ToLower(cipherInfo[0])
		password = cipherInfo[1]
	}
	server := uri.Hostname()
	port, _ := strconv.Atoi(uri.Port())

	moreInfos := uri.Query()
	pluginString := moreInfos.Get("plugin")
	plugin := ""
	pluginOpts := make(map[string]interface{})
	if strings.Contains(pluginString, ";") {
		pluginInfos, err := url.ParseQuery(pluginString)
		if err == nil {
			if strings.Contains(pluginString, "obfs") {
				plugin = "obfs"
				pluginOpts["mode"] = pluginInfos.Get("obfs")
				pluginOpts["host"] = pluginInfos.Get("obfs-host")
			} else if strings.Contains(pluginString, "v2ray") {
				plugin = "v2ray-plugin"
				pluginOpts["mode"] = pluginInfos.Get("mode")
				pluginOpts["host"] = pluginInfos.Get("host")
				pluginOpts["tls"] = strings.Contains(pluginString, "tls")
			}
		}
	}
	if port == 0 || cipher == "" {
		return nil, fmt.Errorf("invalid link")
	}

	return &proxy.Shadowsocks{
		Base: proxy.Base{
			Server: server,
			Port:   port,
		},
		Password:   password,
		Cipher:     cipher,
		Plugin:     plugin,
		PluginOpts: pluginOpts,
	}, nil
}

func base64Decode(src string) (string, error) {
	if src == "" {
		return "", nil
	}

	for _, encoder := range []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	} {
		data, err := encoder.DecodeString(src)
		if err == nil {
			return string(data), err
		}
	}

	return "", fmt.Errorf("parser: base64 decode error")
}
