package rabbitmq

// import (
// 	"context"
// 	"testing"

// 	"github.com/golang/mock/gomock"
// 	"github.com/sgmcloud/api-mono/internal/adapters/secondary/queue"
// 	"github.com/sgmcloud/api-mono/mocks"
// 	"gotest.tools/assert"
// )

// func TestRabbitMQ_SimplePublish(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	logger := mocks.NewMockLogger(ctrl)

// 	rmq := &RabbitMQ{
// 		dsn:    "amqp://guest:guest@localhost:5672/",
// 		logger: logger,
// 	}

// 	assert.NilError(t, rmq.Publish(context.Background(), "test", queue.Message{
// 		Domain:   "test-domain",
// 		Body:     []byte(`{"field": "value"}`),
// 		Metadata: map[string]string{},
// 	}))
// }
