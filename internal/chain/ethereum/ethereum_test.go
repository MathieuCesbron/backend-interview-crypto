package ethereum

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/google/go-cmp/cmp"
	"github.com/segmentio/kafka-go"
)

const (
	publicKey1  = "0x8c4e4B063b682dd78c2E79896361C7B13a9d011F"
	privateKey1 = "bfee092a0cdb7bb15d99b56fe20a938a7cf72616ffb690ca85cf0c0c1fd972ed"

	// we watch this public key
	publicKey2  = "0x1554027f94c6F0FD54B066cB0B0Ef3cBB0aB8eCB"
	privateKey2 = "fa65838a0b111dc2e44a37c6c824b3acd6e15ae45df95e4696779c35ec125d68"

	txID1 = "0xa5b19a9260df27151fdc86fad7881d0b9a1935eb643cf8a12e160b548b484428"
	txID2 = "0x9d4a900791bb1060cba6b0e05ae2f61114ceac289c4c822b9682066a0ad58653"

	amount   = 100_000_000
	gasLimit = 21_000
	gasPrice = 10_0000_000
)

type mockClient struct {
	block       uint64
	fromPrivate string
	to          string
}

func (m *mockClient) BlockNumber(ctx context.Context) (uint64, error) {
	m.block++
	return m.block, nil
}

func (m *mockClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	tx := types.NewTransaction(
		0, common.HexToAddress(m.to), big.NewInt(amount), gasLimit, big.NewInt(gasPrice), nil,
	)

	blockNumber := big.NewInt(int64(m.block))
	blockTime := uint64(time.Now().Unix())

	signer := types.MakeSigner(params.MainnetChainConfig, blockNumber, blockTime)

	pk, _ := crypto.HexToECDSA(m.fromPrivate)
	signedTx, _ := types.SignTx(tx, signer, pk)

	block := types.NewBlockWithHeader(
		&types.Header{
			Number: blockNumber,
			Time:   blockTime,
		},
	)

	block = block.WithBody(types.Body{
		Transactions: []*types.Transaction{signedTx}},
	)

	return block, nil
}

func TestEthereumWatch(t *testing.T) {
	tests := []struct {
		name        string
		fromPrivate string
		to          string
		expectedTx  chain.Transaction
	}{
		{
			name:        "watched one transaction with user as source",
			fromPrivate: privateKey2,
			to:          publicKey1,
			expectedTx: chain.Transaction{
				ID:          txID2,
				Chain:       chain.EthereumName,
				User:        strings.ToLower(publicKey2),
				Source:      strings.ToLower(publicKey2),
				Destination: strings.ToLower(publicKey1),
				Amount:      big.NewInt(amount),
				Fee:         big.NewInt(gasLimit * gasPrice),
			},
		},
		{
			name:        "watched one transaction with user as destination",
			fromPrivate: privateKey1,
			to:          publicKey2,
			expectedTx: chain.Transaction{
				ID:          txID1,
				Chain:       chain.EthereumName,
				User:        strings.ToLower(publicKey2),
				Source:      strings.ToLower(publicKey1),
				Destination: strings.ToLower(publicKey2),
				Amount:      big.NewInt(amount),
				Fee:         big.NewInt(gasLimit * gasPrice),
			},
		},
		{
			name:        "watched zero transaction",
			fromPrivate: privateKey1,
			to:          publicKey1,
			expectedTx:  chain.Transaction{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &mockClient{
				fromPrivate: test.fromPrivate,
				to:          test.to,
			}
			kafkaChan := make(chan kafka.Message, 1)
			e := NewEthereumWatcher(client, kafkaChan)

			os.Setenv("ETHEREUM_ADDRESSES", publicKey2)

			go e.Watch()

			select {
			case msg := <-kafkaChan:
				var got chain.Transaction
				if err := json.Unmarshal(msg.Value, &got); err != nil {
					t.Errorf("failed to decode Kafka message: %v", err)
				}

				if diff := cmp.Diff(test.expectedTx, got, cmp.AllowUnexported(big.Int{})); diff != "" {
					t.Errorf("transaction mismatch. (-want +got):\n%s", diff)
				}

			case <-time.After(chain.EthBlockTicker + time.Second):
				if test.expectedTx != (chain.Transaction{}) {
					t.Errorf("got nothing, expected a transaction: %+v", test.expectedTx)
				}
			}
		})
	}
}
