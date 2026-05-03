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

		publisher, err := New(srv.ClientURL())
		require.NoError(t, err)
		require.NotNil(t, publisher)
		t.Cleanup(publisher.Close)
	})

	t.Run("invalid address", func(t *testing.T) {
		publisher, err := New("://bad address")
		require.Error(t, err)
		assert.Nil(t, publisher)
	})
}

func TestPublish(t *testing.T) {
	srv := runTestServer(t)
	publisher, err := New(srv.ClientURL())
	require.NoError(t, err)
	t.Cleanup(publisher.Close)

	observerConn, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	defer observerConn.Close()

	sub, err := observerConn.SubscribeSync("updates")
	require.NoError(t, err)
	require.NoError(t, observerConn.Flush())

	payload := []byte("comic updated")
	require.NoError(t, publisher.Publish("updates", payload))

	msg, err := sub.NextMsg(time.Second)
	require.NoError(t, err)
	assert.Equal(t, payload, msg.Data)
}

func TestPublishAfterClose(t *testing.T) {
	srv := runTestServer(t)
	publisher, err := New(srv.ClientURL())
	require.NoError(t, err)

	publisher.Close()
	err = publisher.Publish("updates", []byte("data"))
	require.Error(t, err)
}
