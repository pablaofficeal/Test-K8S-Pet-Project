package mq

import "errors"

var ErrPoison = errors.New("poison message")