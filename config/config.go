package config

import "os"

const (
	ListenAddress = ":8080"
	UpstreamURL   = "http://127.0.0.1:3010"
	debug         = true
)

func RemnawaveAPIToken() string {
	return os.Getenv("REMNAWAVE_API_TOKEN")
}

func RemnawaveAPIBaseUrl() string {
	return os.Getenv("REMNAWAVE_API_BASE_URL")
}
