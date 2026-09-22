package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"backyard-client-config-proxy-golang/config"
	"backyard-client-config-proxy-golang/internal/proxy"
	"backyard-client-config-proxy-golang/internal/remnawave"
	"backyard-client-config-proxy-golang/internal/transform"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf(
			".env not loaded: %v",
			err,
		)
	}

	token := config.RemnawaveAPIToken()
	baseURL := config.RemnawaveAPIBaseURL()
	if token == "" || baseURL == "" {
		log.Fatal(
			"REMNAWAVE_API_TOKEN or REMNAWAVE_API_BASE_URL is empty",
		)
	}

	log.Printf(
	"Remnawave API configured: %s, token loaded=%t",
	baseURL,
	token != "",
	)

	remnawaveClient := remnawave.NewClient(
		baseURL,
		token,
	)

	transformer := transform.New(
		remnawaveClient,
	)

	p, err := proxy.New(
		config.UpstreamURL(),
		transformer,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"Starting reverse proxy on %s -> %s",
		config.ListenAddress(),
		config.UpstreamURL(),
	)

	if err := http.ListenAndServe(
		config.ListenAddress(),
		p,
	); err != nil {
		log.Fatal(err)
	}
}