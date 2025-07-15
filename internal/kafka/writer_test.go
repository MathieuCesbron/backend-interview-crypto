// kafka_test.go
package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWriter struct {
	mock.Mock
	messages []kafkago.Message
}

func (m *mockWriter) WriteMessages(ctx context.Context, msgs ...kafkago.Message) error {
	m.messages = append(m.messages, msgs...)
	m.Called(ctx, msgs)
	return nil
}

func TestStartKafka(t *testing.T) {
	writer := new(mockWriter)
	writer.On("WriteMessages", context.Background(), mock.Anything).Return(nil)

	c := make(chan kafkago.Message, 10)

	go StartKafka(c, writer)

	c <- kafka.Message{Key: []byte("key"), Value: []byte("value")}

	time.Sleep(flushInterval + time.Second)

	assert.NotEmpty(t, writer.messages, "expected captured messages")
	assert.Equal(t, "value", string(writer.messages[0].Value))

	writer.AssertExpectations(t)
}
