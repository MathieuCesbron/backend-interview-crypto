package solana

import (
	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
)

const (
	rpcURL = "https://svc.blockdaemon.com/solana/mainnet/native"
)

func CreateClient() *client.Client {
	HTTPClient := chain.NewCustomClient()

	return client.New(rpc.WithEndpoint(rpcURL), rpc.WithHTTPClient(HTTPClient))
}
