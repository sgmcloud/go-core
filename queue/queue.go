package queue

import (
	"context"
	"encoding/json"
)

type Queue interface {
	Publish(ctx context.Context, exchange string, msg Message) error
	Subscribe(ctx context.Context, queue, retry string, f func(Message) error)
}

type Message struct {
	Body     []byte            `json:"body"`
	Metadata map[string]string `json:"metadata"`
}

func (m *Message) Marshal() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		return []byte{}
	}

	return b
}
