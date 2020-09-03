package service

import "time"

type Message struct {
	UID       uint32     `json:"uid"`
	Raw       RawMessage `json:"raw"`
	Timestamp time.Time  `json:"timestamp"`
}

type RawMessage struct {
	Message string `json:"message"`
}
