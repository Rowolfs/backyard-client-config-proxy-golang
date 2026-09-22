package config

import "os"

const (
	Debug = true
)

func ListenAddress() string {
	value := os.Getenv("LISTEN_ADDRESS")
	if value == "" {
		return ":8080"
	}

	return value
}

func UpstreamURL() string {
	value := os.Getenv("UPSTREAM_URL")
	if value == "" {
		return "http://127.0.0.1:3010"
	}

	return value
}

func RemnawaveAPIToken() string {
	return os.Getenv("REMNAWAVE_API_TOKEN")
}

func RemnawaveAPIBaseURL() string {
	return os.Getenv("REMNAWAVE_API_BASE_URL")
}