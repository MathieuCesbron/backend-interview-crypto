package ethereum

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/segmentio/kafka-go"
)

type mockEthClient struct{}

func (m *mockEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	return 100, nil
}

func (m *mockEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	to := common.HexToAddress("")

	tx := types.NewTransaction(
		0, to, big.NewInt(1000000000000000000), 21000, big.NewInt(1000000000), nil,
	)

	block := types.NewBlock(
		&types.Header{},
		&types.Body{
			Transactions: []*types.Transaction{tx}}, nil, nil)

	return block, nil
}

func TestEthereumWatch(t *testing.T) {
	mockClient := &mockEthClient{}
	kafkaChan := make(chan kafka.Message, 1)
	e := NewEthereumWatcher(mockClient, kafkaChan)

	go e.Watch()

	select {
	case msg := <-kafkaChan:
		var tx chain.Transaction
		err := json.Unmarshal(msg.Value, &tx)
		if err != nil {
			t.Fatalf("Failed to decode Kafka message: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("No Kafka message received")
	}
}
