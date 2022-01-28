package main

import (
	"log"

	"github.com/xwjdsh/freeproxy"
	"github.com/xwjdsh/freeproxy/config"
)

func main() {
	cfg, err := config.Init("./config.yml")
	if err != nil {
		log.Fatal(err)
	}
	h := freeproxy.New(cfg)
	h.Start()
}
