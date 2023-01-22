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

func NewMessage(obj interface{}, metadata map[string]string) (*Message, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	msg := &Message{
		Metadata: metadata,
		Body:     b,
	}

	return msg, nil
}

func (m *Message) Marshal() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		return []byte{}
	}

	return b
}
