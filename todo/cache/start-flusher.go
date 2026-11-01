package cache

import (
	"context"
	"time"
	"log"
)

func StartFlusher(ctx context.Context, interval time.Duration, done chan<- struct{}) {
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if _, err := FlushOrders(ctx); err != nil {
					log.Printf("order flush error: %v", err)
				}
			case <-ctx.Done():
				final, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if _, err := FlushOrders(final); err != nil {
					log.Printf("final order flush error: %v", err)
				}
				cancel()
				return
			}
		}
	}()
}