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

	// Синхронное получение сообщений
	sub, err := nc.SubscribeSync("xkcd.db.updated")
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	for {
		msg, err := sub.NextMsgWithContext(ctx)
		if err != nil {
			slog.Error("cannot get next message", "error", err)
			break
		}
		slog.Info("received message", "data", msg.Data)
	}

	if err = sub.Unsubscribe(); err != nil {
		panic(err)
	}
}
