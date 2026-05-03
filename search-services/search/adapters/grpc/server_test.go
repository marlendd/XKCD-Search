package grpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
	searchgrpc "yadro.com/course/search/adapters/grpc"
	"yadro.com/course/search/core"
	"yadro.com/course/search/mocks"
	searchpb "yadro.com/course/proto/search"
)

func newServer(t *testing.T) (*searchgrpc.Server, *mocks.MockSearcher, *mocks.MockISearcher) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockSearcher := mocks.NewMockSearcher(ctrl)
	mockISearcher := mocks.NewMockISearcher(ctrl)
	server := searchgrpc.NewServer(mockSearcher, mockISearcher)
	return server, mockSearcher, mockISearcher
}

func TestPing(t *testing.T) {
	server, _, _ := newServer(t)

	reply, err := server.Ping(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Nil(t, reply)
}

func TestSearch(t *testing.T) {
	server, mockSearcher, _ := newServer(t)

	t.Run("success with mapping", func(t *testing.T) {
		mockSearcher.EXPECT().
			Search(gomock.Any(), "golang", 5).
			Return(core.SearchResult{
				Comics: []core.Comic{
					{ID: 1, URL: "http://example.com"},
				},
			}, nil)

		reply, err := server.Search(context.Background(), &searchpb.SearchRequest{
			Phrase: "golang",
			Limit:  5,
		})
		assert.NoError(t, err)
		assert.Len(t, reply.Comics, 1)
		assert.Equal(t, int64(1), reply.Comics[0].Id)
		assert.Equal(t, "http://example.com", reply.Comics[0].Url)
	})

	t.Run("error", func(t *testing.T) {
		mockSearcher.EXPECT().
			Search(gomock.Any(), "golang", 5).
			Return(core.SearchResult{}, errors.New("search error"))

		_, err := server.Search(context.Background(), &searchpb.SearchRequest{
			Phrase: "golang",
			Limit:  5,
		})
		assert.Error(t, err)
	})
}

func TestISearch(t *testing.T) {
	server, _, mockISearcher := newServer(t)

	t.Run("success with mapping", func(t *testing.T) {
		mockISearcher.EXPECT().
			ISearch(gomock.Any(), "golang", 5).
			Return(core.SearchResult{
				Comics: []core.Comic{
					{ID: 2, URL: "http://example2.com"},
				},
			}, nil)

		reply, err := server.ISearch(context.Background(), &searchpb.SearchRequest{
			Phrase: "golang",
			Limit:  5,
		})
		assert.NoError(t, err)
		assert.Len(t, reply.Comics, 1)
		assert.Equal(t, int64(2), reply.Comics[0].Id)
		assert.Equal(t, "http://example2.com", reply.Comics[0].Url)
	})

	t.Run("error", func(t *testing.T) {
		mockISearcher.EXPECT().
			ISearch(gomock.Any(), "golang", 5).
			Return(core.SearchResult{}, errors.New("search error"))

		_, err := server.ISearch(context.Background(), &searchpb.SearchRequest{
			Phrase: "golang",
			Limit:  5,
		})
		assert.Error(t, err)
	})
}