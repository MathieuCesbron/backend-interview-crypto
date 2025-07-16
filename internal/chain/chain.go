package chain

import "math/big"

type Chain string

const (
	SolanaName   Chain = "solana"
	EthereumName Chain = "ethereum"
)

type Transaction struct {
	// The blockchain network.
	Chain Chain `json:"chain"`

	// Unique identifier of the transaction.
	ID string `json:"id"`

	// User ID watched.
	User string `json:"user"`

	// Sender address of the transaction.
	Source string `json:"source"`

	// Receiver address of the transaction.
	Destination string `json:"destination"`

	// Amount transferred in the transaction, denominated in the smallest unit of the blockchain
	// (e.g., lamports for Solana, wei for Ethereum).
	Amount *big.Int `json:"amount"`

	// Transaction fee.
	Fee *big.Int `json:"fee"`
}

type Watcher interface {
	Name() Chain

	// Addresses returns the list of adress to watch.
	Addresses() []string

	// Watch monitors new blocks for transactions.
	Watch()
}
