package search

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	log    *slog.Logger
	client searchpb.SearchClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to grpc", "error", err)
		return nil, err
	}
	return &Client{
		client: searchpb.NewSearchClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	return err
}

func (c Client) Search(ctx context.Context, phrase string, limit int) (core.SearchResult, error) {
	reply, err := c.client.Search(ctx, &searchpb.SearchRequest{
		Limit: int64(limit),
		Phrase: phrase,
	})
	if err != nil {
		return core.SearchResult{}, err
	}
	comics := make([]core.Comic, len(reply.Comics))
	for i, c := range reply.Comics {
		comics[i] = core.Comic{ID: int(c.Id), URL: c.Url}
	}

	return core.SearchResult{Comics: comics}, nil
}

func (c Client) ISearch(ctx context.Context, phrase string, limit int) (core.SearchResult, error) {
	reply, err := c.client.ISearch(ctx, &searchpb.SearchRequest{
		Limit: int64(limit),
		Phrase: phrase,
	})
	if err != nil {
		return core.SearchResult{}, err
	}
	comics := make([]core.Comic, len(reply.Comics))
	for i, c := range reply.Comics {
		comics[i] = core.Comic{ID: int(c.Id), URL: c.Url}
	}

	return core.SearchResult{Comics: comics}, nil
}