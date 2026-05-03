package initiator_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"yadro.com/course/search/adapters/initiator"
	"yadro.com/course/search/mocks"
)

func TestRun(t *testing.T) {
	t.Run("calls BuildIndex immediately", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mock := mocks.NewMockIndexer(ctrl)

		mock.EXPECT().BuildIndex(gomock.Any()).Return(nil).MinTimes(1)

		ctx, cancel := context.WithCancel(context.Background())
		i := initiator.New(mock, time.Hour, slog.Default())

		go i.Run(ctx)
		time.Sleep(50 * time.Millisecond)
		cancel()
	})

	t.Run("stops on ctx.Done", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mock := mocks.NewMockIndexer(ctrl)

		mock.EXPECT().BuildIndex(gomock.Any()).Return(nil).MinTimes(1)

		ctx, cancel := context.WithCancel(context.Background())
		i := initiator.New(mock, time.Hour, slog.Default())

		done := make(chan struct{})
		go func() {
			i.Run(ctx)
			close(done)
		}()

		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Run did not stop after ctx.Done")
		}
	})

	t.Run("does not panic on BuildIndex error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mock := mocks.NewMockIndexer(ctrl)

		mock.EXPECT().BuildIndex(gomock.Any()).Return(errors.New("index error")).MinTimes(1)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		i := initiator.New(mock, time.Hour, slog.Default())

		go i.Run(ctx)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("calls BuildIndex on ticker", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mock := mocks.NewMockIndexer(ctrl)

		// минимум 2 вызова: сразу + по тикеру
		mock.EXPECT().BuildIndex(gomock.Any()).Return(nil).MinTimes(2)

		ctx, cancel := context.WithCancel(context.Background())
		i := initiator.New(mock, 50*time.Millisecond, slog.Default())

		go i.Run(ctx)
		time.Sleep(200 * time.Millisecond)
		cancel()
	})
}