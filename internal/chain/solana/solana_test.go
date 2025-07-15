package solana

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/google/go-cmp/cmp"
	"github.com/mr-tron/base58"
	"github.com/segmentio/kafka-go"
)

const (
	publicKey1 = "bUz1BcqoGdWC32C5EfA5UVoCqTibxBSTtztjwwcQBQR"

	// we watch this public key
	publicKey2 = "Es1cHBCrQKnQ8EHHBXMquCr2msuwV7PZ1oHz4D9WLoC5"

	fee    = 5000
	amount = 100_000
)

var (
	txID = make([]byte, 64)
)

type mockClient struct {
	slot uint64
	from string
	to   string
}

func (m *mockClient) GetSlot(ctx context.Context) (uint64, error) {
	m.slot++
	return m.slot, nil
}

func (m *mockClient) GetBlockWithConfig(ctx context.Context, slot uint64, cfg client.GetBlockConfig) (*client.Block, error) {
	amount := uint64(amount)
	data := make([]byte, 12)
	data[0] = 2
	binary.LittleEndian.PutUint32(data[0:4], 2)
	binary.LittleEndian.PutUint64(data[4:12], amount)

	return &client.Block{
		Transactions: []client.BlockTransaction{
			{
				Transaction: types.Transaction{
					Message: types.Message{
						Accounts: []common.PublicKey{
							common.PublicKeyFromString(m.from),
							common.PublicKeyFromString(m.to),
						},
						Instructions: []types.CompiledInstruction{
							{
								ProgramIDIndex: 2,
								Accounts:       []int{0, 1},
								Data:           data,
							},
						},
					},
					Signatures: []types.Signature{
						txID,
					},
				},
				Meta: &client.TransactionMeta{
					Fee: fee,
				},
				AccountKeys: []common.PublicKey{
					common.PublicKeyFromString(m.from),
					common.PublicKeyFromString(m.to),
					common.SystemProgramID,
				},
			},
		},
	}, nil
}

func TestSolanaWatch(t *testing.T) {
	tests := []struct {
		name       string
		from       string
		to         string
		expectedTx chain.Transaction
	}{
		{
			name: "watched one transaction with user as source",
			from: publicKey2,
			to:   publicKey1,
			expectedTx: chain.Transaction{
				Chain:       chain.SolanaName,
				ID:          base58.Encode(txID),
				User:        publicKey2,
				Source:      publicKey2,
				Destination: publicKey1,
				Amount:      big.NewInt(amount),
				Fee:         big.NewInt(fee),
			},
		},
		{
			name: "watched one transaction with user as destination",
			from: publicKey1,
			to:   publicKey2,
			expectedTx: chain.Transaction{
				Chain:       chain.SolanaName,
				ID:          base58.Encode(txID),
				User:        publicKey2,
				Source:      publicKey1,
				Destination: publicKey2,
				Amount:      big.NewInt(amount),
				Fee:         big.NewInt(fee),
			},
		},
		{
			name:       "watched zero transaction",
			from:       publicKey1,
			to:         publicKey1,
			expectedTx: chain.Transaction{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &mockClient{
				from: test.from,
				to:   test.to,
			}

			kafkaChan := make(chan kafka.Message, 1)
			s := NewSolanaWatcher(client, kafkaChan)

			os.Setenv("SOLANA_ADDRESSES", publicKey2)

			go s.Watch()

			select {
			case msg := <-kafkaChan:
				var got chain.Transaction
				if err := json.Unmarshal(msg.Value, &got); err != nil {
					t.Errorf("failed to decode kafka message: %v", err)
				}

				if diff := cmp.Diff(test.expectedTx, got, cmp.AllowUnexported(big.Int{})); diff != "" {
					t.Errorf("transaction mismatch. (-want +got):\n%s", diff)
				}

			case <-time.After(UpdateMaxSlotTicker + time.Second):
				if test.expectedTx != (chain.Transaction{}) {
					t.Errorf("got nothing, expected a transaction: %+v", test.expectedTx)
				}
			}
		})
	}
}
