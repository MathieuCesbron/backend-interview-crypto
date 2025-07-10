package chain

type Chain string

const (
	SolanaName   Chain = "solana"
	EthereumName Chain = "ethereum"
)

type Transaction struct {
	Chain       Chain  `json:"chain"`
	ID          string `json:"id"`
	User        string `json:"user"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Amount      uint64 `json:"amount"`
	Fee         uint64 `json:"fee"`
}

type Watcher interface {
	Name() Chain
	Watch()
}
