package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func InitKafkaWriter() *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "transactions",
	})
}

func CreateKafkaTopic() error {
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             "transactions",
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		return err
	}

	return nil
}

func StartKafka(msgChan <-chan kafka.Message, writer *kafka.Writer) {
	const (
		maxBatchSize  = 100
		flushInterval = 200 * time.Millisecond
	)

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	var batch []kafka.Message

	for {
		select {
		case msg := <-msgChan:
			batch = append(batch, msg)

			if len(batch) >= maxBatchSize {
				flushKafkaBatch(writer, &batch)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				flushKafkaBatch(writer, &batch)
			}
		}
	}
}

func flushKafkaBatch(writer *kafka.Writer, batch *[]kafka.Message) {
	err := writer.WriteMessages(context.Background(), (*batch)...)
	if err != nil {
		log.Printf("Kafka write error: %v", err)
	} else {
		log.Printf("✅ Wrote %d transactions to Kafka", len(*batch))
	}
	*batch = (*batch)[:0]
}
