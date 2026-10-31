package providers

import (
	"context"
	"log"
)

type TestProvider struct{}

func (TestProvider) Name() string {
	return "test"
}

func (TestProvider) Send(ctx context.Context, msg Message) error {
	log.Printf("test provider: nothing was delivered to %s, the fixed test code is in use", MaskEmail(msg.To))
	return nil
}