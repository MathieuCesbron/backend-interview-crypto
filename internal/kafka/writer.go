package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	batchSize     = 100
	flushInterval = 200 * time.Millisecond

	numPartitions     = 1
	replicationFactor = 1

	broker = "localhost:9092"
	topic  = "transactions"
)

type Writer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

func InitKafkaWriter() *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   topic,
	})
}

func CreateKafkaTopic() error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	})
}

func StartKafka(msgChan <-chan kafka.Message, writer Writer) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	var batch []kafka.Message

	for {
		select {
		case msg := <-msgChan:
			batch = append(batch, msg)

			if len(batch) >= batchSize {
				flushBatch(writer, &batch)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				flushBatch(writer, &batch)
			}
		}
	}
}

func flushBatch(writer Writer, batch *[]kafka.Message) {
	err := writer.WriteMessages(context.Background(), (*batch)...)
	if err != nil {
		log.Printf("Kafka write error: %v", err)
	} else {
		log.Printf("✅ Wrote %d transactions to Kafka", len(*batch))
	}
	*batch = (*batch)[:0]
}
