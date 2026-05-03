package search_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/adapters/search"
	"yadro.com/course/api/core"
	"yadro.com/course/api/mocks"
	searchpb "yadro.com/course/proto/search"
)

func newClient(t *testing.T) (*search.Client, *mocks.MockSearchClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockSearchClient(ctrl)
	client := search.NewClientFromMock(mock, slog.Default())
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

func TestSearch(t *testing.T) {
	client, mock := newClient(t)

	t.Run("success with mapping", func(t *testing.T) {
		mock.EXPECT().
			Search(gomock.Any(), &searchpb.SearchRequest{Phrase: "golang", Limit: 5}, gomock.Any()).
			Return(&searchpb.SearchReply{
				Comics: []*searchpb.SearchResult{
					{Id: 1, Url: "http://example.com"},
				},
			}, nil)

		result, err := client.Search(context.Background(), "golang", 5)
		assert.NoError(t, err)
		assert.Equal(t, core.SearchResult{
			Comics: []core.Comic{{ID: 1, URL: "http://example.com"}},
		}, result)
	})

	t.Run("grpc error", func(t *testing.T) {
		mock.EXPECT().
			Search(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, status.Error(codes.Internal, "internal error"))

		_, err := client.Search(context.Background(), "golang", 5)
		assert.Error(t, err)
	})
}

func TestISearch(t *testing.T) {
	client, mock := newClient(t)

	t.Run("success with mapping", func(t *testing.T) {
		mock.EXPECT().
			ISearch(gomock.Any(), &searchpb.SearchRequest{Phrase: "golang", Limit: 5}, gomock.Any()).
			Return(&searchpb.SearchReply{
				Comics: []*searchpb.SearchResult{
					{Id: 2, Url: "http://example2.com"},
				},
			}, nil)

		result, err := client.ISearch(context.Background(), "golang", 5)
		assert.NoError(t, err)
		assert.Equal(t, core.SearchResult{
			Comics: []core.Comic{{ID: 2, URL: "http://example2.com"}},
		}, result)
	})

	t.Run("grpc error", func(t *testing.T) {
		mock.EXPECT().
			ISearch(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, status.Error(codes.Internal, "internal error"))

		_, err := client.ISearch(context.Background(), "golang", 5)
		assert.Error(t, err)
	})
}