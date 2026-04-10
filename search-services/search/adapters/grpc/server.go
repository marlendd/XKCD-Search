package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

func NewServer(searcher core.Searcher, isearcher core.ISearcher) *Server {
	return &Server{
		searcher: searcher,
		isearcher: isearcher,
	}
}

type Server struct {
	searchpb.UnimplementedSearchServer
	searcher core.Searcher
	isearcher core.ISearcher
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *Server) Search(ctx context.Context, req *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	res, err := s.searcher.Search(ctx, req.Phrase, int(req.Limit))
	if err != nil {
		return nil, err
	}
	comics := make([]*searchpb.SearchResult, len(res.Comics))
	for i, c := range res.Comics {
		comics[i] = &searchpb.SearchResult{
			Id:  int64(c.ID),
			Url: c.URL,
		}
	}

	return &searchpb.SearchReply{
		Comics: comics,
	}, nil
}

func (s *Server) ISearch(ctx context.Context, req *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	res, err := s.isearcher.ISearch(ctx, req.Phrase, int(req.Limit))
	if err != nil {
		return nil, err
	}
	comics := make([]*searchpb.SearchResult, len(res.Comics))
	for i, c := range res.Comics {
		comics[i] = &searchpb.SearchResult{
			Id:  int64(c.ID),
			Url: c.URL,
		}
	}

	return &searchpb.SearchReply{
		Comics: comics,
	}, nil
}
