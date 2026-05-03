package words_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/search/adapters/words"
	"yadro.com/course/search/core"
	"yadro.com/course/search/mocks"
	wordspb "yadro.com/course/proto/words"
)

func newClient(t *testing.T) (*words.Client, *mocks.MockWordsClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockWordsClient(ctrl)
	return words.NewClientFromMock(mock, slog.Default()), mock
}

func TestPing(t *testing.T) {
	client, mock := newClient(t)

	t.Run("success", func(t *testing.T) {
		mock.EXPECT().
			Ping(gomock.Any(), nil, gomock.Any()).
			Return(&emptypb.Empty{}, nil)

		assert.NoError(t, client.Ping(context.Background()))
	})

	t.Run("error", func(t *testing.T) {
		mock.EXPECT().
			Ping(gomock.Any(), nil, gomock.Any()).
			Return(nil, errors.New("unavailable"))

		assert.Error(t, client.Ping(context.Background()))
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
			phrase:      "too long",
			err:         status.Error(codes.ResourceExhausted, "too long"),
			expectedErr: core.ErrBadArguments,
		},
		{
			name:        "grpc error",
			phrase:      "golang",
			err:         status.Error(codes.Internal, "internal"),
			expectedErr: status.Error(codes.Internal, "internal"),
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