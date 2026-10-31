package mq

import (
	"github.com/google/uuid"
	"encoding/json"
	"fmt"
)

type VerificationRequest struct {
	Event    string `json:"event"`
	UserUUID string `json:"user_uuid"`
	TraceID  string `json:"trace_id"`
}

func ParseVerificationRequest(body []byte) (VerificationRequest, error) {
	var req VerificationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return req, fmt.Errorf("%w: %v", ErrPoison, err)
	}
	if req.Event != EventVerificationRequested {
		return req, fmt.Errorf("%w: unexpected event %q", ErrPoison, req.Event)
	}
	if _, err := uuid.Parse(req.UserUUID); err != nil {
		return req, fmt.Errorf("%w: invalid user uuid", ErrPoison)
	}
	return req, nil
}