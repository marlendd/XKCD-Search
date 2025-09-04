package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	// Подключаемся к брокеру
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		panic(err)
	}
	slog.Info("connected to broker")
	defer nc.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			// Отправляем сообщение в топик "xkcd.updates"
			slog.Info("sending message to subscribers")
			err = nc.Publish("xkcd.db.updated", []byte("XKCD DB has been updated"))
			if err != nil {
				slog.Error("could not publish message", "error", err)
			}
			nc.Flush()
		case <-ctx.Done():
			return
		}
	}
}
