package mq

import (
	"Test-K8s-Pet-Project-Verify/providers"
	"Test-K8s-Pet-Project-Verify/configs"
	"Test-K8s-Pet-Project-Verify/core"
	"gorm.io/gorm"
	"context"
	"errors"
	"log"
)

func NewVerificationHandler(gdb *gorm.DB, provider providers.Provider, cfg configs.VerifyConfig) func(context.Context, []byte) error {
	return func(ctx context.Context, body []byte) error {
		req, err := ParseVerificationRequest(body)
		if err != nil {
			return err
		}

		err = core.Issue(ctx, gdb, provider, cfg, req.UserUUID)
		var cooldown *core.CooldownError
		switch {
		case err == nil:
			log.Printf("trace_id=%s verification code issued", req.TraceID)
			return nil
		case errors.Is(err, core.ErrAlreadyVerified), errors.Is(err, core.ErrRateLimited), errors.As(err, &cooldown):
			log.Printf("trace_id=%s verification request skipped: %v", req.TraceID, err)
			return nil
		}
		return err
	}
}