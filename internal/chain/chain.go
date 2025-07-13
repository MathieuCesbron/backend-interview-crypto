package chain

import "math/big"

type Chain string

const (
	SolanaName   Chain = "solana"
	EthereumName Chain = "ethereum"
)

type Transaction struct {
	Chain       Chain    `json:"chain"`
	ID          string   `json:"id"`
	User        string   `json:"user"`
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Amount      *big.Int `json:"amount"`
	Fee         *big.Int `json:"fee"`
}

type Watcher interface {
	Name() Chain
	Addresses() []string
	Watch()
}
