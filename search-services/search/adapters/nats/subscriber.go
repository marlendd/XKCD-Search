package nats

import "github.com/nats-io/nats.go"

type Subscriber struct {
	conn *nats.Conn
}

func New(address string) (*Subscriber, error) {
	nc, err := nats.Connect(address)
	if err != nil {
		return nil, err
	}

	return &Subscriber{conn: nc}, nil
}

func (s *Subscriber) Subscribe(subject string, handler func()) error {
	_, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler()
	})

	return err
}

func (s *Subscriber) Close() {
    s.conn.Close()
}
