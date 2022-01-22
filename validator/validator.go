package validator

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/xwjdsh/proxypool/crawler"
)

type Result struct {
	Proxy                *crawler.Proxy
	Available            bool
	Location             string
	Total                time.Duration
	Dns                  time.Duration
	TlsHandshak          time.Duration
	Connect              time.Duration
	GotFirstResponseByte time.Duration
}

type Validator struct{}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(ch chan *crawler.Proxy, resultChan chan *Result) {
	wg := sync.WaitGroup{}
	for {
		select {
		case proxy, ok := <-ch:
			if !ok {
				wg.Wait()
				return
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				resultChan <- v.ValidateOne(proxy)
			}()
		}
	}
}

func (v *Validator) ValidateOne(proxy *crawler.Proxy) *Result {
	addr := fmt.Sprintf("%s:%d", proxy.Host, proxy.Port)
	client := http.Client{
		Transport: &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(fmt.Sprintf("http://%s", addr))
			},
		},
	}
	defer client.CloseIdleConnections()

	var start, connect, dns, tlsHandshake time.Time
	result := &Result{Proxy: proxy}
	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			result.Dns = time.Since(dns)
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			result.TlsHandshak = time.Since(tlsHandshake)
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			result.Connect = time.Since(connect)
		},

		GotFirstResponseByte: func() {
			result.GotFirstResponseByte = time.Since(start)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(
		httptrace.WithClientTrace(ctx, trace),
		http.MethodGet,
		"http://myip.ipip.net",
		nil,
	)

	start = time.Now()
	resp, err := client.Transport.RoundTrip(req)
	if err != nil {
		return result
	}
	result.Total = time.Since(start)

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result
	}

	str := string(data)
	if strings.HasPrefix(str, "当前 IP") {
		indexText := "来自于："
		if i := strings.LastIndex(str, indexText); i != -1 {
			result.Location = str[i+len(indexText):]
		}
		result.Available = true
		return result
	}

	return result
}
