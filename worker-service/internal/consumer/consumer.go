package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/Bharat1Rajput/workerService/internal/config"
	"github.com/Bharat1Rajput/workerService/internal/model"
	"github.com/Bharat1Rajput/workerService/internal/processor"
)

type Consumer struct {
	cfg       *config.Config
	processor *processor.Processor
	logger    *zap.Logger
	conn      *amqp.Connection
	ch        *amqp.Channel
	wg        sync.WaitGroup
	sem       chan struct{}
}

func New(cfg *config.Config, proc *processor.Processor, logger *zap.Logger) (*Consumer, error) {
	conn, err := amqp.Dial(cfg.RabbitURL)
	if err != nil {
		return nil, fmt.Errorf("consumer: connect rabbit: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("consumer: channel: %w", err)
	}

	if err := ch.Qos(cfg.WorkerConcurrency, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("consumer: qos: %w", err)
	}

	if err := ch.ExchangeDeclare(
		cfg.RabbitExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("consumer: declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(
		cfg.RabbitQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("consumer: declare queue: %w", err)
	}

	if err := ch.QueueBind(
		q.Name,
		cfg.RabbitRoutingKey,
		cfg.RabbitExchange,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("consumer: bind queue: %w", err)
	}

	return &Consumer{
		cfg:       cfg,
		processor: proc,
		logger:    logger,
		conn:      conn,
		ch:        ch,
		sem:       make(chan struct{}, cfg.WorkerConcurrency),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.ch.Consume(
		c.cfg.RabbitQueue,
		"worker-service",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consumer: start consume: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("consumer: context canceled, waiting for workers")
			c.wg.Wait()
			return nil
		case d, ok := <-deliveries:
			if !ok {
				c.logger.Info("consumer: deliveries channel closed")
				c.wg.Wait()
				return nil
			}

			c.sem <- struct{}{}
			c.wg.Add(1)

			go func(d amqp.Delivery) {
				defer c.wg.Done()
				defer func() { <-c.sem }()

				var job model.WebhookJob
				if err := json.Unmarshal(d.Body, &job); err != nil {
					c.logger.Error("consumer: unmarshal job", zap.Error(err))
					_ = d.Nack(false, false)
					return
				}

				jobCtx := context.Background()
				if err := c.processor.ProcessJob(jobCtx, &job); err != nil {
					// permanent failure or retries exhausted
					c.logger.Error("consumer: job processing failed", zap.Error(err), zap.String("job_id", job.ID))
					_ = d.Nack(false, false)
					return
				}

				if err := d.Ack(false); err != nil {
					c.logger.Error("consumer: ack failed", zap.Error(err))
				}
			}(d)
		}
	}
}

func (c *Consumer) Close() error {
	if err := c.ch.Close(); err != nil {
		_ = c.conn.Close()
		return err
	}
	return c.conn.Close()
}

