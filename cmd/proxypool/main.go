package main

import (
	"log"

	"github.com/xwjdsh/proxypool"
	"github.com/xwjdsh/proxypool/config"
)

func main() {
	cfg, err := config.Init("./config.yml")
	if err != nil {
		log.Fatal(err)
	}
	h := proxypool.New(cfg)
	h.Start()
}
