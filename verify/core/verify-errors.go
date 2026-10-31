package core

import (
	"errors"
	"time"
	"fmt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrAlreadyVerified = errors.New("email already verified")
	ErrNoActiveCode    = errors.New("no active verification code")
	ErrTooManyAttempts = errors.New("too many attempts")
	ErrRateLimited     = errors.New("too many verification requests")
	ErrSendFailed      = errors.New("could not send the verification code")
)

type CooldownError struct {
	RetryAfter time.Duration
}

func (e *CooldownError) Error() string {
	return fmt.Sprintf("wait %s before requesting a new code", e.RetryAfter.Round(time.Second))
}

type InvalidCodeError struct {
	AttemptsLeft int
}

func (e *InvalidCodeError) Error() string {
	return "invalid code"
}