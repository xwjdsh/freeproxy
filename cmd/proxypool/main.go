package main

import "github.com/xwjdsh/proxypool"

func main() {
	h := proxypool.New()
	h.Start()
}
