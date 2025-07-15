package chain

import (
	"net/http"
	"os"
	"time"
)

const timeout = 3 * time.Second

type BearerTokenRoundTripper struct {
	Next  http.RoundTripper
	Token string
}

func (t *BearerTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.Token)

	return t.Next.RoundTrip(req)
}

func NewCustomClient() *http.Client {
	tokenTransport := &BearerTokenRoundTripper{
		Next:  http.DefaultTransport,
		Token: os.Getenv("BLOCKDAEMON_API_KEY"),
	}

	return &http.Client{
		Transport: tokenTransport,
		Timeout:   timeout,
	}
}
