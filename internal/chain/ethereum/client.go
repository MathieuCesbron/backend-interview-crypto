package ethereum

import (
	"context"

	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const rpcURL = "https://svc.blockdaemon.com/ethereum/mainnet/native"

func CreateClient() *ethclient.Client {
	HTTPClient := chain.NewCustomClient()
	rpcClient, _ := rpc.DialOptions(context.Background(), rpcURL, rpc.WithHTTPClient(HTTPClient))

	return ethclient.NewClient(rpcClient)
}
