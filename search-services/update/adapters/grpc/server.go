package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
)

func NewServer(service core.Updater) *Server {
	return &Server{service: service}
}

type Server struct {
	updatepb.UnimplementedUpdateServer
	service core.Updater
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *Server) Status(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatusReply, error) {
	var st updatepb.Status

	switch s.service.Status(ctx) {
	case core.StatusIdle:
		st = updatepb.Status_STATUS_IDLE
	case core.StatusRunning:
		st = updatepb.Status_STATUS_RUNNING
	default:
		st = updatepb.Status_STATUS_UNSPECIFIED
	}

	return &updatepb.StatusReply{Status: st}, nil
}

func (s *Server) Update(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.service.Update(ctx)
	if err != nil {
		if errors.Is(err, core.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Stats(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatsReply, error) {
	st, err := s.service.Stats(ctx)
	if err != nil {
		return nil, err
	}

	return &updatepb.StatsReply{
		WordsTotal:    int64(st.WordsTotal),
		WordsUnique:   int64(st.WordsUnique),
		ComicsTotal:   int64(st.ComicsTotal),
		ComicsFetched: int64(st.ComicsFetched),
	}, nil
}

func (s *Server) Drop(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.service.Drop(ctx); err != nil {
		if errors.Is(err, core.ErrAlreadyExists) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
