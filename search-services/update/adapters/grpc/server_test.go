package grpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	updatepb "yadro.com/course/proto/update"
	updategrpc "yadro.com/course/update/adapters/grpc"
	"yadro.com/course/update/core"
	"yadro.com/course/update/mocks"
)

func newServer(t *testing.T) (*updategrpc.Server, *mocks.MockUpdater) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockUpdater(ctrl)
	return updategrpc.NewServer(mock), mock
}

func TestPing(t *testing.T) {
	server, _ := newServer(t)

	reply, err := server.Ping(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Nil(t, reply)
}

func TestStatus(t *testing.T) {
	server, mock := newServer(t)

	cases := []struct {
		name           string
		coreStatus     core.ServiceStatus
		expectedStatus updatepb.Status
	}{
		{
			name:           "idle",
			coreStatus:     core.StatusIdle,
			expectedStatus: updatepb.Status_STATUS_IDLE,
		},
		{
			name:           "running",
			coreStatus:     core.StatusRunning,
			expectedStatus: updatepb.Status_STATUS_RUNNING,
		},
		{
			name:           "default/unknown",
			coreStatus:     core.ServiceStatus("unknown"),
			expectedStatus: updatepb.Status_STATUS_UNSPECIFIED,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock.EXPECT().
				Status(gomock.Any()).
				Return(tc.coreStatus)

			reply, err := server.Status(context.Background(), &emptypb.Empty{})
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, reply.Status)
		})
	}
}

func TestUpdate(t *testing.T) {
	server, mock := newServer(t)

	t.Run("success", func(t *testing.T) {
		mock.EXPECT().Update(gomock.Any()).Return(nil)

		_, err := server.Update(context.Background(), &emptypb.Empty{})
		assert.NoError(t, err)
	})

	t.Run("already exists maps to codes.AlreadyExists", func(t *testing.T) {
		mock.EXPECT().Update(gomock.Any()).Return(core.ErrAlreadyExists)

		_, err := server.Update(context.Background(), &emptypb.Empty{})
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	t.Run("other error", func(t *testing.T) {
		mock.EXPECT().Update(gomock.Any()).Return(errors.New("unexpected"))

		_, err := server.Update(context.Background(), &emptypb.Empty{})
		assert.Error(t, err)
		assert.NotEqual(t, codes.AlreadyExists, status.Code(err))
	})
}

func TestStats(t *testing.T) {
	server, mock := newServer(t)

	t.Run("success with mapping", func(t *testing.T) {
		mock.EXPECT().Stats(gomock.Any()).Return(core.ServiceStats{
			DBStats: core.DBStats{
				WordsTotal:    100,
				WordsUnique:   50,
				ComicsFetched: 10,
			}, ComicsTotal: 20,
		}, nil)

		reply, err := server.Stats(context.Background(), &emptypb.Empty{})
		assert.NoError(t, err)
		assert.Equal(t, int64(100), reply.WordsTotal)
		assert.Equal(t, int64(50), reply.WordsUnique)
		assert.Equal(t, int64(20), reply.ComicsTotal)
		assert.Equal(t, int64(10), reply.ComicsFetched)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().Stats(gomock.Any()).Return(core.ServiceStats{}, errors.New("stats error"))

		_, err := server.Stats(context.Background(), &emptypb.Empty{})
		assert.Error(t, err)
	})
}

func TestDrop(t *testing.T) {
	server, mock := newServer(t)

	t.Run("success", func(t *testing.T) {
		mock.EXPECT().Drop(gomock.Any()).Return(nil)

		_, err := server.Drop(context.Background(), &emptypb.Empty{})
		assert.NoError(t, err)
	})

	t.Run("already exists maps to codes.FailedPrecondition", func(t *testing.T) {
		mock.EXPECT().Drop(gomock.Any()).Return(core.ErrAlreadyExists)

		_, err := server.Drop(context.Background(), &emptypb.Empty{})
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))
	})

	t.Run("other error", func(t *testing.T) {
		mock.EXPECT().Drop(gomock.Any()).Return(errors.New("unexpected"))

		_, err := server.Drop(context.Background(), &emptypb.Empty{})
		assert.Error(t, err)
		assert.NotEqual(t, codes.FailedPrecondition, status.Code(err))
	})
}
