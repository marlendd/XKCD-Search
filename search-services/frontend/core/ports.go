package core

import "context"

type APIClient interface {
	Login(ctx context.Context, user, password string) (string, error)
	Ping(ctx context.Context) (PingResponse, error)
	Search(ctx context.Context, phrase string, limit int) (SearchResponse, error)
	ISearch(ctx context.Context, phrase string, limit int) (SearchResponse, error)
	Update(ctx context.Context, token string) error
	Stats(ctx context.Context) (UpdateStatsResponse, error)
	Status(ctx context.Context) (StatusResponse, error)
	Drop(ctx context.Context, token string) error
}
