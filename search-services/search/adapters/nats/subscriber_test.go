package nats

import (
	"testing"
	"time"

	natssrv "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runTestServer(t *testing.T) *natssrv.Server {
	t.Helper()

	srv, err := natssrv.NewServer(&natssrv.Options{
		Host:   "127.0.0.1",
		Port:   -1,
		NoLog:  true,
		NoSigs: true,
	})
	require.NoError(t, err)

	go srv.Start()
	require.True(t, srv.ReadyForConnections(5*time.Second))

	t.Cleanup(func() {
		srv.Shutdown()
		srv.WaitForShutdown()
	})

	return srv
}

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := runTestServer(t)

		subscriber, err := New(srv.ClientURL())
		require.NoError(t, err)
		require.NotNil(t, subscriber)
		t.Cleanup(subscriber.Close)
	})

	t.Run("invalid address", func(t *testing.T) {
		subscriber, err := New("://bad address")
		require.Error(t, err)
		assert.Nil(t, subscriber)
	})
}

func TestSubscribe(t *testing.T) {
	srv := runTestServer(t)
	subscriber, err := New(srv.ClientURL())
	require.NoError(t, err)
	t.Cleanup(subscriber.Close)

	done := make(chan struct{}, 1)
	err = subscriber.Subscribe("search.rebuild", func() {
		done <- struct{}{}
	})
	require.NoError(t, err)

	publisherConn, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	defer publisherConn.Close()

	require.NoError(t, publisherConn.Publish("search.rebuild", []byte("trigger")))
	require.NoError(t, publisherConn.Flush())

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handler was not called")
	}
}

func TestSubscribeAfterClose(t *testing.T) {
	srv := runTestServer(t)
	subscriber, err := New(srv.ClientURL())
	require.NoError(t, err)

	subscriber.Close()
	err = subscriber.Subscribe("search.rebuild", func() {})
	require.Error(t, err)
}
