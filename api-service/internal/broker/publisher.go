package broker

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Publisher interface {
	Publish(ctx context.Context, routingKey string, body []byte) error
	Close() error
}

type RabbitPublisher struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	exchange string
	logger   *zap.Logger
}

func NewRabbitPublisher(url, exchange, queue, routingKey string, logger *zap.Logger) (*RabbitPublisher, error) {
	var conn *amqp.Connection
	var ch *amqp.Channel
	var err error

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		backoff := time.Duration(1<<i) * time.Second
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
		logger.Warn("broker.rabbit: dial failed, retrying",
			zap.Error(err),
			zap.Duration("backoff", backoff),
		)
		time.Sleep(backoff)
	}
	if err != nil {
		return nil, fmt.Errorf("broker.rabbit: connect: %w", err)
	}

	ch, err = conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("broker.rabbit: channel: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("broker.rabbit: confirm mode: %w", err)
	}

	if err := ch.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("broker.rabbit: declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("broker.rabbit: declare queue: %w", err)
	}

	if err := ch.QueueBind(
		q.Name,
		routingKey,
		exchange,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("broker.rabbit: bind queue: %w", err)
	}

	return &RabbitPublisher{
		conn:     conn,
		ch:       ch,
		exchange: exchange,
		logger:   logger,
	}, nil
}

func (p *RabbitPublisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	confirms := p.ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	if err := p.ch.PublishWithContext(
		ctx,
		p.exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			DeliveryMode: amqp.Persistent,
		},
	); err != nil {
		return fmt.Errorf("broker.rabbit: publish: %w", err)
	}

	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("broker.rabbit: publish not acknowledged")
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (p *RabbitPublisher) Close() error {
	if err := p.ch.Close(); err != nil {
		_ = p.conn.Close()
		return err
	}
	return p.conn.Close()
}

