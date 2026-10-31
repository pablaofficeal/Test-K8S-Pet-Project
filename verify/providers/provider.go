package providers

import (
	"context"
	"time"
)

type Message struct {
	To        string
	Username  string
	Code      string
	ExpiresIn time.Duration
}

type Provider interface {
	Name() string
	Send(ctx context.Context, msg Message) error
}