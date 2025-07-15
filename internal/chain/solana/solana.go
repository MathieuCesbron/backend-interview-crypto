package solana

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"log"
	"math/big"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/mr-tron/base58"
	"github.com/segmentio/kafka-go"
)

const (
	// For now, we will use 1 worker to avoid rate limiting issues.
	maxSlotWorkers = 1

	// The time to wait before updating the max slot.
	// The lower the better, but we don't want to get rate limited.
	UpdateMaxSlotTicker = 500 * time.Millisecond
)

type SolanaWatcher struct {
	Client SolClient

	CurrentSlot uint64
	MaxSlot     uint64

	KafkaChan chan<- kafka.Message
}

type SolClient interface {
	GetSlot(ctx context.Context) (uint64, error)
	GetBlockWithConfig(ctx context.Context, slot uint64, cfg client.GetBlockConfig) (*client.Block, error)
}

func NewSolanaWatcher(client SolClient, kafkaChan chan<- kafka.Message) *SolanaWatcher {
	s := &SolanaWatcher{
		Client:    client,
		KafkaChan: kafkaChan,
	}

	maxSlot, err := s.GetMaxSlot()
	for err != nil {
		log.Printf("error getting solana max slot: %v. Retrying...\n", err)
		time.Sleep(time.Second)
		maxSlot, err = s.GetMaxSlot()
	}

	atomic.StoreUint64(&s.CurrentSlot, maxSlot)
	atomic.StoreUint64(&s.MaxSlot, maxSlot)

	return s
}

func (s *SolanaWatcher) Name() chain.Chain {
	return chain.SolanaName
}

func (s *SolanaWatcher) Addresses() []string {
	return strings.Split(os.Getenv("SOLANA_ADDRESSES"), ",")
}

func (s *SolanaWatcher) GetMaxSlot() (uint64, error) {
	slot, err := s.Client.GetSlot(context.Background())
	if err != nil {
		return 0, err
	}
	return slot, nil
}

func (s *SolanaWatcher) UpdateMaxSlot() {
	ticker := time.NewTicker(UpdateMaxSlotTicker)
	defer ticker.Stop()

	for range ticker.C {
		maxSlot, err := s.GetMaxSlot()
		if err != nil {
			log.Printf("error getting current solana slot: %v", err)
			continue
		}
		atomic.StoreUint64(&s.MaxSlot, maxSlot)

		current := atomic.LoadUint64(&s.CurrentSlot)
		log.Printf("Solana slot lag: %d", maxSlot-current)
	}
}

func (s *SolanaWatcher) GetTxs(slot uint64) ([]client.BlockTransaction, error) {
	block, err := s.Client.GetBlockWithConfig(context.Background(), slot, client.GetBlockConfig{
		TransactionDetails: "full",
	})
	if err != nil {
		return nil, err
	}

	return block.Transactions, nil
}

func (s *SolanaWatcher) FilterTxs(txs []client.BlockTransaction) []chain.Transaction {
	filtered := []chain.Transaction{}

	for _, tx := range txs {
		if tx.Meta == nil || tx.Meta.Err != nil {
			continue
		}

		for _, inst := range tx.Transaction.Message.Instructions {
			if tx.AccountKeys[inst.ProgramIDIndex] != common.SystemProgramID {
				continue
			}
			if len(inst.Data) < 12 {
				continue
			}
			if binary.LittleEndian.Uint32(inst.Data[:4]) != 2 {
				continue
			}

			amount := new(big.Int).SetUint64(binary.LittleEndian.Uint64(inst.Data[4:12]))
			fee := new(big.Int).SetUint64(tx.Meta.Fee)

			source := tx.AccountKeys[inst.Accounts[0]].String()
			destination := tx.AccountKeys[inst.Accounts[1]].String()

			for _, addr := range s.Addresses() {
				if source == addr || destination == addr {
					filtered = append(filtered, chain.Transaction{
						Chain:       chain.SolanaName,
						ID:          base58.Encode(tx.Transaction.Signatures[0]),
						User:        addr,
						Source:      source,
						Destination: destination,
						Amount:      amount,
						Fee:         fee,
					})
					break
				}
			}
		}
	}

	return filtered
}

func (s *SolanaWatcher) startWorkerPool(slots chan uint64, workers int) {
	for range workers {
		go func() {
			for slot := range slots {
				s.handleSlot(slot)
			}
		}()
	}
}

func (s *SolanaWatcher) handleSlot(slot uint64) {
	txs, err := s.GetTxs(slot)
	if err != nil {
		log.Printf("error getting solana transactions for slot %d: %v\n", slot, err)
		return
	}

	filteredTxs := s.FilterTxs(txs)
	for _, filteredTx := range filteredTxs {
		payload, err := json.Marshal(filteredTx)
		if err != nil {
			log.Printf("error marshalling solana transaction: %+v\n", filteredTx)
			continue
		}
		s.KafkaChan <- kafka.Message{Value: payload}
	}
}

func (s *SolanaWatcher) scheduleSlots(slots chan uint64) {
	for {
		currentSlot := atomic.LoadUint64(&s.CurrentSlot)
		maxSlot := atomic.LoadUint64(&s.MaxSlot)

		if currentSlot < maxSlot {
			slots <- currentSlot
			atomic.AddUint64(&s.CurrentSlot, 1)
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (s *SolanaWatcher) Watch() {
	go s.UpdateMaxSlot()

	slots := make(chan uint64, maxSlotWorkers)

	s.startWorkerPool(slots, maxSlotWorkers)
	s.scheduleSlots(slots)
}
