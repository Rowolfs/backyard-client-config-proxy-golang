package proxy

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"backyard-client-config-proxy-golang/config"
	"backyard-client-config-proxy-golang/internal/transform"
)

func New(
	upstreamURL string,
	transformer *transform.Transformer,
) (*httputil.ReverseProxy, error) {

	upstream, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	p := httputil.NewSingleHostReverseProxy(upstream)

	originalDirector := p.Director

	p.Director = func(req *http.Request) {
		originalDirector(req)

		req.Header.Set("Accept-Encoding", "identity")

		log.Printf(
			"request: %s %s hwid=%s ua=%s",
			req.Method,
			req.URL.Path,
			req.Header.Get("X-Hwid"),
			req.Header.Get("User-Agent"),
		)
	}

	p.ModifyResponse = func(resp *http.Response) error {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		log.Printf(
			"response: status=%s type=%s size=%d",
			resp.Status,
			resp.Header.Get("Content-Type"),
			len(body),
		)

		subscriptionID := extractSubscriptionID(
			resp.Request.URL.Path,
			config.RemnawaveSecretSubsPath(),
		)

		if subscriptionID == "" {
			log.Printf(
				"[proxy] subscription id not found in path=%q",
				resp.Request.URL.Path,
			)
		} else {
			log.Printf(
				"[proxy] subscription id=%s",
				subscriptionID,
			)
		}

		newBody, err := transformer.Transform(
			resp.Request.Context(),
			subscriptionID,
			body,
		)
		if err != nil {
			return err
		}

		resp.Body = io.NopCloser(
			bytes.NewReader(newBody),
		)

		resp.ContentLength = int64(len(newBody))

		resp.Header.Del("Content-Length")
		resp.Header.Del("ETag")

		return nil
	}

	p.ErrorHandler = func(
		w http.ResponseWriter,
		req *http.Request,
		err error,
	) {
		log.Printf("proxy error: %v", err)

		http.Error(
			w,
			"Bad Gateway",
			http.StatusBadGateway,
		)
	}

	return p, nil
}

func extractSubscriptionID(
	path string,
	secretPath string,
) string {

	secretPath = strings.Trim(secretPath, "/")
	path = strings.Trim(path, "/")

	if secretPath == "" || path == "" {
		return ""
	}

	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		return ""
	}

	if parts[0] != secretPath {
		return ""
	}

	return parts[1]
}
