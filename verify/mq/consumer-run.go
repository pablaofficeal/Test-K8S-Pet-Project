package mq

import (
	"context"
	"time"
	"log"
)

func RunConsumer(ctx context.Context, url string, interval time.Duration, handle func(context.Context, []byte) error) {
	for ctx.Err() == nil {
		if err := ConsumeOnce(ctx, url, handle); err != nil {
			log.Printf("rabbitmq consumer stopped, retry in %s: %v", interval.String(), err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}