package service

import "context"

type Beacon interface {
	Ping(ctx context.Context) (string, error)
}
