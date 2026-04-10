package update

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	updatepb "yadro.com/course/proto/update"
)

type Client struct {
	log    *slog.Logger
	client updatepb.UpdateClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to grpc", "error", err)
		return nil, err
	}
	return &Client{
		client: updatepb.NewUpdateClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	return err
}

func (c Client) Status(ctx context.Context) (core.UpdateStatus, error) {
	reply, err := c.client.Status(ctx, &emptypb.Empty{})
	if err != nil {
		return core.UpdateStatus(core.StatusUpdateUnknown), err
	}
	switch reply.Status {
	case updatepb.Status_STATUS_IDLE:
		return core.StatusUpdateIdle, nil
	case updatepb.Status_STATUS_RUNNING:
		return core.StatusUpdateRunning, nil
	default:
		return core.StatusUpdateUnknown, nil
	}
}

func (c Client) Stats(ctx context.Context) (core.UpdateStats, error) {
	reply, err := c.client.Stats(ctx, &emptypb.Empty{})
	if err != nil {
		return core.UpdateStats{}, err
	}

	return core.UpdateStats{
		WordsTotal:    int(reply.WordsTotal),
		WordsUnique:   int(reply.WordsUnique),
		ComicsFetched: int(reply.ComicsFetched),
		ComicsTotal:   int(reply.ComicsTotal),
	}, nil
}

func (c Client) Update(ctx context.Context) error {
	_, err := c.client.Update(ctx, &emptypb.Empty{})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return core.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (c Client) Drop(ctx context.Context) error {
	if _, err := c.client.Drop(ctx, &emptypb.Empty{}); err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return core.ErrAlreadyExists
		}
		return err
	}
	return nil
}