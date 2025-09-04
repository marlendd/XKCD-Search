package update

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
		return nil, err
	}
	return &Client{
		client: updatepb.NewUpdateClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	return errors.New("implement me")
}

func (c Client) Status(ctx context.Context) (core.UpdateStatus, error) {
	return core.StatusUpdateUnknown, errors.New("unknown status")
}

func (c Client) Stats(ctx context.Context) (core.UpdateStats, error) {
	return core.UpdateStats{}, nil
}

func (c Client) Update(ctx context.Context) error {
	return errors.New("implement me")
}

func (c Client) Drop(ctx context.Context) error {
	return errors.New("implement me")
}
