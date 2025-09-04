package core

import (
	"context"
)

type Updater interface {
	Update(context.Context) error
	Stats(context.Context) (ServiceStats, error)
	Status(context.Context) ServiceStatus
	Drop(context.Context) error
}

type DB interface {
	Add(context.Context, Comics) error
	Stats(context.Context) (DBStats, error)
	Drop(context.Context) error
	IDs(context.Context) ([]int, error)
}

type XKCD interface {
	Get(context.Context, int) (XKCDInfo, error)
	LastID(context.Context) (int, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}
