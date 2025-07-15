package chain

import "time"

const (
	// Each of these variables controls API request rate.
	// Higher means fresher data but risk of rate limiting.

	// Max concurrent blocks processed for ethereum
	EthBlockWorkers = 2
	// Delay between block updates for ethereum
	EthBlockTicker = 2 * time.Second

	// Max concurrent slots processed for solana
	SolSlotWorkers = 1
	// Delay between slot updates for solana
	UpdateSlotTicker = 500 * time.Millisecond
)
