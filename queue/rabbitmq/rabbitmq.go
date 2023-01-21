package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sgmcloud/go-core/log"
	"github.com/sgmcloud/go-core/queue"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

type RabbitMQ struct {
	logger log.Logger
	dsn    string
}

type LoggerAdapter struct{}

func (LoggerAdapter) Fatalf(string, ...interface{}) {}
func (LoggerAdapter) Errorf(string, ...interface{}) {}
func (LoggerAdapter) Warnf(string, ...interface{})  {}
func (LoggerAdapter) Infof(string, ...interface{})  {}
func (LoggerAdapter) Debugf(string, ...interface{}) {}
func (LoggerAdapter) Tracef(string, ...interface{}) {}

func NewRabbit(logger log.Logger, dsn string) (*RabbitMQ, error) {
	conn, err := rabbitmq.NewConn(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "init rabbitmq connection")
	}
	defer conn.Close()

	return &RabbitMQ{
		logger: logger,
		dsn:    dsn,
	}, nil
}

// func (rmq *RabbitMQ) InitTopic(topic string) error {
// 	s := jazz.Settings{
// 		Exchanges: map[string]jazz.Exchange{
// 			topic: {
// 				Durable: bool,
// 				Type: "topic",
// 			},
// 		},
// 		Queues: map[string]jazz.QueueSpec{
// 			topic: {
// 				Durable: bool,
// 				Bindings: []jazz.Binding{
// 					{
// 						Exchange: topic,
// 						Key: "#",
// 					},
// 				},
// 			},
// 		},
// 	}

// 	if err := rmq.jazz.CreateScheme(s); err != nil {
// 		return err
// 	}

// 	return nil
// }

// Publish: exchangeDsn: exchange-name@key
func (rmq *RabbitMQ) Publish(ctx context.Context, exchange string, msg queue.Message) error {
	logger := &LoggerAdapter{}

	conn, err := rabbitmq.NewConn(
		rmq.dsn,
		rabbitmq.WithConnectionOptionsLogger(logger),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogger(logger),
		rabbitmq.WithPublisherOptionsExchangeName(exchange),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeDurable,
		rabbitmq.WithPublisherOptionsExchangeKind("fanout"),
	)
	if err != nil {
		return err
	}
	defer publisher.Close()

	publisher.NotifyReturn(func(r rabbitmq.Return) {
		rmq.logger.Debug("message returned from server: " + string(r.Body))
	})

	publisher.NotifyPublish(func(c rabbitmq.Confirmation) {
		rmq.logger.Debug("message confirmed from server", log.Any("tag", c.DeliveryTag), log.Any("ack", c.Ack))
	})

	buff := msg.Marshal()
	err = publisher.Publish(
		buff,
		[]string{""}, // routing keys
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsExchange(exchange),
	)
	if err != nil {
		return err
	}

	return nil
}

// retry isn't used on rabbitMQ, just ignore
func (rmq *RabbitMQ) Subscribe(ctx context.Context, topic, retry string, f func(queue.Message) error) {
	conn, err := rabbitmq.NewConn(
		rmq.dsn,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		rmq.logger.Error("initializing rabbitmq connection", err, log.Any("dsn", rmq.dsn))

		return
	}

	consumer, err := rabbitmq.NewConsumer(
		conn,
		func(d rabbitmq.Delivery) rabbitmq.Action {
			var msg queue.Message

			if err := json.Unmarshal(d.Body, &msg); err != nil {
				rmq.logger.Error("unmarshal message", err, log.Any("topic", topic), log.Any("body", string(d.Body)))

				return rabbitmq.NackDiscard
			}

			if err := f(msg); err != nil {
				rmq.logger.Error("nacked message", err, log.Any("topic", topic), log.Error(err))

				return rabbitmq.NackDiscard
			}
			rmq.logger.Debug("consumed", log.Any("topic", topic))

			// rabbitmq.Ack, rabbitmq.NackDiscard, rabbitmq.NackRequeue
			return rabbitmq.Ack
		},
		topic,
		rabbitmq.WithConsumerOptionsExchangeName(topic),
		rabbitmq.WithConsumerOptionsQueueDurable,
		rabbitmq.WithConsumerOptionsQueueArgs(rabbitmq.Table{"x-dead-letter-exchange": topic + ".dlx"}),
	)
	if err != nil {
		rmq.logger.Error("initializing rabbitmq consumer", err, log.Any("queue", topic))

		return
	}
	defer consumer.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Millisecond):
		}
	}
}
