package core

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	cases := []struct {
		name        string
		concurrency int
		wantErr     bool
	}{
		{"invalid concurrency", -1, true},
		{"invalid concurrency", 1, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			service, err := NewService(log, nil, nil, nil, tc.concurrency, nil)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, service)
			} else {
				require.NoError(t, err)
				require.NotNil(t, service)
			}

		})
	}
}

func TestStatus(t *testing.T) {
	cases := []struct {
		name      string
		isRunning bool
		want      ServiceStatus
	}{
		{"idle", false, StatusIdle},
		{"running", true, StatusRunning},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			service, err := NewService(log, nil, nil, nil, 1, nil)
			require.NoError(t, err)
			service.running.Store(tc.isRunning)
			require.Equal(t, tc.want, service.Status(context.Background()))
		})
	}
}

func TestDrop_AlreadyRunning(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := NewService(log, nil, nil, nil, 1, nil)
	require.NoError(t, err)
	service.running.Store(true)
	err = service.Drop(context.Background())
	require.Equal(t, ErrAlreadyExists, err)
}
