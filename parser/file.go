package parser

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
)

type generalFileExecutor struct {
	address string
	name    string
}

func (c *generalFileExecutor) Name() string {
	return c.name
}

func (c *generalFileExecutor) Execute(ctx context.Context, linkChan chan<- *linkResp) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.address, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	respData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(respData))
	for scanner.Scan() {
		linkChan <- &linkResp{
			Source: c.name,
			Link:   scanner.Text(),
		}
	}

	return scanner.Err()
}
