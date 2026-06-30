package bootstrap

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type HTTPClient struct {
	client *http.Client
}

type HTTPError struct {
	StatusCode int
	Method     string
	URL        string
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s %s failed with status %d: %s", e.Method, e.URL, e.StatusCode, e.Body)
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func NewKubeHTTPClient() (*HTTPClient, string, string, error) {
	tokenBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return nil, "", "", err
	}
	caBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return nil, "", "", err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caBytes)

	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
					RootCAs:    pool,
				},
			},
		},
	}, string(bytes.TrimSpace(tokenBytes)), string(caBytes), nil
}

func (c *HTTPClient) JSONRequest(method, rawURL string, headers map[string]string, body []byte) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Method:     method,
			URL:        rawURL,
			Body:       string(bytes.TrimSpace(payload)),
		}
	}
	return payload, nil
}

func BearerHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
		"Accept":        "application/json",
	}
}
