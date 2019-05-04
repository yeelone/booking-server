package models

import "time"

type Message struct {
	ID uint64
	Text string
	Error bool
	ErrorText string
	CreatedBy User
	CreatedAt  time.Time
}

// TableName :
func (b *Message) TableName() string {
	return "message"
}

