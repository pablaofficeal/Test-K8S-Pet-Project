package mq

import (
	"testing"
	"errors"
)

func TestParseVerificationRequest(t *testing.T) {
	good := `{"event":"verification.requested","user_uuid":"7450e1ea-47a5-4a8b-9c2d-0123456789ab","trace_id":"t1"}`
	req, err := ParseVerificationRequest([]byte(good))
	if err != nil || req.UserUUID == "" || req.TraceID != "t1" {
		t.Fatalf("valid message rejected: %v", err)
	}

	bad := map[string]string{
		"not json":       `{`,
		"wrong event":    `{"event":"other","user_uuid":"7450e1ea-47a5-4a8b-9c2d-0123456789ab"}`,
		"bad uuid":       `{"event":"verification.requested","user_uuid":"nope"}`,
		"missing fields": `{}`,
	}
	for name, body := range bad {
		_, err := ParseVerificationRequest([]byte(body))
		if !errors.Is(err, ErrPoison) {
			t.Errorf("%s: expected a poison error, got %v", name, err)
		}
	}
}