package ethereum

import (
	"context"
	"encoding/json"
	"log"
	"math/big"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/segmentio/kafka-go"
)

type EthereumWatcher struct {
	Client EthClient

	CurrentBlock uint64
	MaxBlock     uint64

	KafkaChan chan<- kafka.Message
}

type EthClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
}

func NewEthereumWatcher(client EthClient, kafkaChan chan<- kafka.Message) *EthereumWatcher {
	e := &EthereumWatcher{
		Client:    client,
		KafkaChan: kafkaChan,
	}

	maxBlock, err := e.Client.BlockNumber(context.Background())
	for err != nil {
		log.Printf("error getting ethereum max block: %v. Retrying...\n", err)
		time.Sleep(time.Second)
		maxBlock, err = e.Client.BlockNumber(context.Background())
	}

	atomic.StoreUint64(&e.MaxBlock, maxBlock)
	atomic.StoreUint64(&e.CurrentBlock, maxBlock)

	return e
}

func (e *EthereumWatcher) Name() chain.Chain {
	return chain.EthereumName
}

func (e *EthereumWatcher) Addresses() []string {
	return strings.Split(strings.ToLower(os.Getenv("ETHEREUM_ADDRESSES")), ",")
}

func (e *EthereumWatcher) UpdateMaxBlock() {
	ticker := time.NewTicker(chain.EthBlockTicker)
	defer ticker.Stop()

	for range ticker.C {
		maxBlock, err := e.Client.BlockNumber(context.Background())
		if err != nil {
			log.Printf("error getting ethereum current block: %v", err)
			continue
		}

		current := atomic.LoadUint64(&e.CurrentBlock)

		if maxBlock >= current {
			atomic.StoreUint64(&e.MaxBlock, maxBlock)
			log.Printf("Ethereum block lag: %d blocks", maxBlock-current)
		}
	}
}

func (e *EthereumWatcher) startWorkerPool(blocks <-chan uint64, workers int) {
	for range workers {
		go func() {
			for block := range blocks {
				e.handleBlock(block)
			}
		}()
	}
}

func (e *EthereumWatcher) FilterTxs(data *types.Block) []chain.Transaction {
	filtered := []chain.Transaction{}

	for _, tx := range data.Transactions() {
		if tx.To() == nil || len(tx.Data()) != 0 {
			continue
		}

		signer := types.MakeSigner(params.MainnetChainConfig, data.Number(), data.Time())
		wallet, err := types.Sender(
			signer, tx,
		)
		if err != nil {
			log.Printf("error deriving signature for ethereum transaction %s: %v", tx.Hash().Hex(), err)
			continue
		}

		amount := tx.Value()
		fee := new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(tx.Gas()))

		source := strings.ToLower(wallet.Hex())
		destination := strings.ToLower(tx.To().Hex())

		for _, addr := range e.Addresses() {
			if source == addr || destination == addr {
				filtered = append(filtered, chain.Transaction{
					Chain:       chain.EthereumName,
					ID:          tx.Hash().Hex(),
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

	return filtered
}

func (e *EthereumWatcher) handleBlock(block uint64) {
	data, err := e.Client.BlockByNumber(context.Background(), big.NewInt(int64(block)))
	if err != nil {
		log.Printf("error getting ethereum transactions for block %d: %v", block, err)
		return
	}

	filteredTxs := e.FilterTxs(data)
	for _, filteredTx := range filteredTxs {
		payload, err := json.Marshal(filteredTx)
		if err != nil {
			log.Printf("error marshalling ethereum transaction: %+v\n", filteredTx)
			continue
		}
		e.KafkaChan <- kafka.Message{Value: payload}
	}
}

func (e *EthereumWatcher) scheduleBlocks(blocks chan<- uint64) {
	for {
		currentBlock := atomic.LoadUint64(&e.CurrentBlock)
		maxBlock := atomic.LoadUint64(&e.MaxBlock)

		if currentBlock < maxBlock {
			blocks <- currentBlock
			atomic.AddUint64(&e.CurrentBlock, 1)
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (e *EthereumWatcher) Watch() {
	go e.UpdateMaxBlock()

	blocks := make(chan uint64, chain.EthBlockWorkers)

	e.startWorkerPool(blocks, chain.EthBlockWorkers)
	e.scheduleBlocks(blocks)
}
