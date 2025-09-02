package service

import (
	"context"
)

type Tender struct {
	Name string
	URL  string
}

type TenderService interface {
	Push(ctx context.Context, hostname, ip string) error
	Pull(ctx context.Context) (map[string]string, error)
}
