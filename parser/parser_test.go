package parser

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCfmem(t *testing.T) {
	linkChan := make(chan *linkResp)
	go func() {
		for {
			<-linkChan
		}
	}()
	err := cfmemInstance.Execute(context.Background(), linkChan)
	assert.Nil(t, err)
}

func TestFreefq(t *testing.T) {
	linkChan := make(chan *linkResp)
	go func() {
		for {
			fmt.Println(<-linkChan)
		}
	}()
	err := freefqInstance.Execute(context.Background(), linkChan)
	assert.Nil(t, err)
}
