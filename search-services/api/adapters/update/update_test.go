package update_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/adapters/update"
	"yadro.com/course/api/core"
	"yadro.com/course/api/mocks"
	updatepb "yadro.com/course/proto/update"
)

func newClient(t *testing.T) (*update.Client, *mocks.MockUpdateClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockUpdateClient(ctrl)
	client := update.NewClientFromMock(mock, slog.Default())
	return client, mock
}

func TestStatus(t *testing.T) {
	client, mock := newClient(t)

	cases := []struct {
		name           string
		reply          *updatepb.StatusReply
		err            error
		expectedStatus core.UpdateStatus
		expectedErr    bool
	}{
		{
			name:           "idle",
			reply:          &updatepb.StatusReply{Status: updatepb.Status_STATUS_IDLE},
			expectedStatus: core.StatusUpdateIdle,
		},
		{
			name:           "running",
			reply:          &updatepb.StatusReply{Status: updatepb.Status_STATUS_RUNNING},
			expectedStatus: core.StatusUpdateRunning,
		},
		{
			name:           "default/unknown",
			reply:          &updatepb.StatusReply{Status: updatepb.Status_STATUS_UNSPECIFIED},
			expectedStatus: core.StatusUpdateUnknown,
		},
		{
			name:        "grpc error",
			err:         status.Error(codes.Internal, "internal error"),
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock.EXPECT().
				Status(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
				Return(tc.reply, tc.err)

			s, err := client.Status(context.Background())
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedStatus, s)
			}
		})
	}
}

func TestStats(t *testing.T) {
	client, mock := newClient(t)

	t.Run("success", func(t *testing.T) {
		mock.EXPECT().
			Stats(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
			Return(&updatepb.StatsReply{
				WordsTotal:    100,
				WordsUnique:   50,
				ComicsFetched: 10,
				ComicsTotal:   20,
			}, nil)

		stats, err := client.Stats(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, core.UpdateStats{
			WordsTotal:    100,
			WordsUnique:   50,
			ComicsFetched: 10,
			ComicsTotal:   20,
		}, stats)
	})

	t.Run("grpc error", func(t *testing.T) {
		mock.EXPECT().
			Stats(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
			Return(nil, status.Error(codes.Internal, "internal error"))

		_, err := client.Stats(context.Background())
		assert.Error(t, err)
	})
}

func TestUpdate(t *testing.T) {
	client, mock := newClient(t)

	cases := []struct {
		name        string
		err         error
		expectedErr error
	}{
		{
			name: "success",
		},
		{
			name:        "already exists",
			err:         status.Error(codes.AlreadyExists, "already exists"),
			expectedErr: core.ErrAlreadyExists,
		},
		{
			name:        "grpc error",
			err:         status.Error(codes.Internal, "internal error"),
			expectedErr: status.Error(codes.Internal, "internal error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock.EXPECT().
				Update(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
				Return(&emptypb.Empty{}, tc.err)

			err := client.Update(context.Background())
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDrop(t *testing.T) {
	client, mock := newClient(t)

	cases := []struct {
		name        string
		err         error
		expectedErr error
	}{
		{
			name: "success",
		},
		{
			name:        "already exists",
			err:         status.Error(codes.FailedPrecondition, "update is running"),
			expectedErr: core.ErrAlreadyExists,
		},
		{
			name:        "grpc error",
			err:         status.Error(codes.Internal, "internal error"),
			expectedErr: status.Error(codes.Internal, "internal error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock.EXPECT().
				Drop(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
				Return(&emptypb.Empty{}, tc.err)

			err := client.Drop(context.Background())
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}