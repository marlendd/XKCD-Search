package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"
)

func main() {
	// Подключаемся к брокеру
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Получение сообщений в канал с буфером на 10
	ch := make(chan *nats.Msg, 10)
	sub, err := nc.ChanSubscribe("xkcd.db.updated", ch)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			if err = sub.Unsubscribe(); err != nil {
				panic(err)
			}
			return
		case msg := <-ch:
			slog.Info("message received", "data", msg.Data)
		}
	}
}
