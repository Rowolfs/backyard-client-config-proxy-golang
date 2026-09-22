package main

import (
	"log"
	"net/http"

	"backyard-client-config-proxy-golang/config"
	"backyard-client-config-proxy-golang/internal/proxy"
)

func main() {
	p, err := proxy.New(config.UpstreamURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"Starting reverse proxy on %s -> %s",
		config.ListenAddress,
		config.UpstreamURL,
	)

	if err := http.ListenAndServe(
		config.ListenAddress,
		p,
	); err != nil {
		log.Fatal(err)
	}
}