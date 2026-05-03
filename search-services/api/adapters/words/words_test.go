package words_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/adapters/words"
	"yadro.com/course/api/core"
	"yadro.com/course/api/mocks"
	wordspb "yadro.com/course/proto/words"
)

func newClient(t *testing.T) (*words.Client, *mocks.MockWordsClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockWordsClient(ctrl)
	client := words.NewClientFromMock(mock, slog.Default())
	return client, mock
}

func TestPing(t *testing.T) {
	client, mock := newClient(t)

	t.Run("success", func(t *testing.T) {
		mock.EXPECT().
			Ping(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
			Return(&emptypb.Empty{}, nil)

		err := client.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().
			Ping(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
			Return(nil, status.Error(codes.Unavailable, "unavailable"))

		err := client.Ping(context.Background())
		assert.Error(t, err)
	})
}

func TestNorm(t *testing.T) {
	client, mock := newClient(t)

	cases := []struct {
		name        string
		phrase      string
		reply       *wordspb.WordsReply
		err         error
		expected    []string
		expectedErr error
	}{
		{
			name:     "success",
			phrase:   "running cats",
			reply:    &wordspb.WordsReply{Words: []string{"run", "cat"}},
			expected: []string{"run", "cat"},
		},
		{
			name:        "resource exhausted maps to ErrBadArguments",
			phrase:      "too long phrase",
			err:         status.Error(codes.ResourceExhausted, "too long"),
			expectedErr: core.ErrBadArguments,
		},
		{
			name:        "grpc error",
			phrase:      "golang",
			err:         status.Error(codes.Internal, "internal error"),
			expectedErr: status.Error(codes.Internal, "internal error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock.EXPECT().
				Norm(gomock.Any(), &wordspb.WordsRequest{Phrase: tc.phrase}, gomock.Any()).
				Return(tc.reply, tc.err)

			result, err := client.Norm(context.Background(), tc.phrase)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}