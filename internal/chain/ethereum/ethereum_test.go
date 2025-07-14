package ethereum

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/segmentio/kafka-go"
)

const (
	addr1 = "0xb0CEe983642cA499A42FDf5c50882F42cDB3de15"
	addr2 = "0x92a1F35B8df8B3f98cB41b5f115ed6A114257804"
)

type mockEthClient struct {
	block uint64
}

func (m *mockEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	m.block++
	return m.block, nil
}

func (m *mockEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	// from := common.HexToAddress(addr1)
	to := common.HexToAddress(addr2)

	tx := types.NewTransaction(
		0, to, big.NewInt(1000000000000000000), 21000, big.NewInt(1000000000), nil,
	)

	// maybe because the block if fake ?
	block := types.NewBlockWithHeader(
		&types.Header{
			Number: big.NewInt(500),
		},
	)

	block = block.WithBody(types.Body{
		Transactions: []*types.Transaction{tx}},
	)

	return block, nil
}

func TestEthereumWatch(t *testing.T) {
	mockClient := &mockEthClient{}
	kafkaChan := make(chan kafka.Message, 1)
	e := NewEthereumWatcher(mockClient, kafkaChan)

	to := common.HexToAddress(addr2)
	os.Setenv("ETHEREUM_ADDRESSES", to.String())
	go e.Watch()

	select {
	case msg := <-kafkaChan:
		var tx chain.Transaction
		err := json.Unmarshal(msg.Value, &tx)
		if err != nil {
			t.Fatalf("Failed to decode Kafka message: %v", err)
		}
	case <-time.After(200 * time.Second):
		t.Fatal("No Kafka message received")
	}
}
