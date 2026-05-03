package nats

import (
	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn *nats.Conn
}

func New(address string) (*Publisher, error) {
	nc, err := nats.Connect(address)
	if err != nil {
		return nil, err
	}

	return &Publisher{conn: nc}, nil
}

func (p *Publisher) Publish(subject string, data []byte) error {
    if err := p.conn.Publish(subject, data); err != nil {
        return err
    }
    return p.conn.Flush()
}

func (p *Publisher) Close() {
	p.conn.Close()
}
