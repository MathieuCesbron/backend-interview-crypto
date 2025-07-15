package main

import (
	"log"

	"github.com/MathieuCesbron/backend-interview-crypto/internal"
	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain"
	"github.com/MathieuCesbron/backend-interview-crypto/internal/chain/solana"
	"github.com/MathieuCesbron/backend-interview-crypto/internal/kafka"

	kafkago "github.com/segmentio/kafka-go"
)

func main() {
	// check env variables
	err := internal.CheckEnvVars()
	if err != nil {
		log.Fatal(err)
	}

	err = kafka.CreateKafkaTopic()
	if err != nil {
		log.Fatal("failed to create Kafka topic:", err)
	}
	kafkaWriter := kafka.InitKafkaWriter()
	kafkaChan := make(chan kafkago.Message, 1000)

	// Start kafka writer
	go kafka.StartKafka(kafkaChan, kafkaWriter)

	// Start blockchains watchers
	watchers := []chain.Watcher{
		solana.NewSolanaWatcher(
			solana.CreateClient(), kafkaChan),
		// ethereum.NewEthereumWatcher(
		// 	ethereum.CreateClient(), kafkaChan),
	}
	for _, watcher := range watchers {
		go watcher.Watch()
		log.Printf("Started watching chain: %s\n", watcher.Name())
	}

	select {}
}
