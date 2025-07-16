package main

import (
	"log"
	"os"

	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain"
	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain/ethereum"
	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain/solana"
	"github.com/MathieuCesbron/backend-interview-crypto/internal/kafka"
	"github.com/joho/godotenv"

	kafkago "github.com/segmentio/kafka-go"
)

const (
	kafkaBuffer = 1000

	EnvBlockdaemonAPIKey = "BLOCKDAEMON_API_KEY"
)

func main() {
	// check env var
	_ = godotenv.Load()
	if os.Getenv(EnvBlockdaemonAPIKey) == "" {
		log.Fatalf("missing required environment variable: %s", EnvBlockdaemonAPIKey)
	}

	err := kafka.CreateKafkaTopic()
	if err != nil {
		log.Fatal("failed to create Kafka topic:", err)
	}
	kafkaWriter := kafka.InitKafkaWriter()
	kafkaChan := make(chan kafkago.Message, kafkaBuffer)

	// start kafka writer
	go kafka.StartKafka(kafkaChan, kafkaWriter)

	// watch each supported blockchain
	watchers := []chain.Watcher{
		solana.NewSolanaWatcher(
			solana.CreateClient(), kafkaChan),
		ethereum.NewEthereumWatcher(
			ethereum.CreateClient(), kafkaChan),
	}
	for _, watcher := range watchers {
		if len(watcher.Addresses()) != 0 {
			go watcher.Watch()
			log.Printf("Started watching chain: %s\n", watcher.Name())
		}
	}

	select {}
}
