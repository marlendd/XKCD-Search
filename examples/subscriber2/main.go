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

	// Асинхронное получение сообщений
	sub, err := nc.Subscribe("xkcd.db.updated", func(msg *nats.Msg) {
		slog.Info("received message", "data", msg.Data)
	})

	if err != nil {
		panic(err)
	}

	<-ctx.Done()

	if err = sub.Unsubscribe(); err != nil {
		panic(err)
	}
}
