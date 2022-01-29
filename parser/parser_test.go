package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCfmem(t *testing.T) {
	linkChan := make(chan string)
	go func() {
		for {
			<-linkChan
		}
	}()
	err := cfmemInstance.Execute(context.Background(), linkChan)
	assert.Nil(t, err)
}
