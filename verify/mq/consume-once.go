package mq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"context"
	"errors"
	"log"
	"fmt"
)

func ConsumeOnce(ctx context.Context, url string, handle func(context.Context, []byte) error) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("connect to rabbitmq: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := ch.Qos(10, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}
	deliveries, err := ch.Consume(VerifyQueue, "verify-service", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %s: %w", VerifyQueue, err)
	}
	log.Printf("consuming %s", VerifyQueue)

	closed := conn.NotifyClose(make(chan *amqp.Error, 1))
	for {
		select {
		case <-ctx.Done():
			return nil
		case amqpErr := <-closed:
			return fmt.Errorf("connection closed: %v", amqpErr)
		case d, ok := <-deliveries:
			if !ok {
				return errors.New("deliveries channel closed")
			}
			err := handle(ctx, d.Body)
			switch {
			case err == nil:
				d.Ack(false)
			case errors.Is(err, ErrPoison):
				log.Printf("dropping message to the dead letter queue: %v", err)
				d.Reject(false)
			default:
				log.Printf("message will be retried: %v", err)
				d.Nack(false, true)
			}
		}
	}
}