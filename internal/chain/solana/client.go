package solana

import (
	"net/http"
	"os"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
)

const (
	rpcURL = "https://svc.blockdaemon.com/solana/mainnet/native"
)

type bearerTokenRoundTripper struct {
	next  http.RoundTripper
	token string
}

func (t *bearerTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)

	return t.next.RoundTrip(req)
}

func CreateSolanaClient() *client.Client {
	tokenTransport := &bearerTokenRoundTripper{
		next:  http.DefaultTransport,
		token: os.Getenv("BLOCKDAEMON_API_KEY"),
	}

	customHTTPClient := &http.Client{
		Transport: tokenTransport,
		Timeout:   10 * time.Second,
	}

	return client.New(rpc.WithEndpoint(rpcURL), rpc.WithHTTPClient(customHTTPClient))
}
