package service

import (
	"context"
	"net/http"
)

type Beacon struct {
	Name string
	URL  string
}

type BeaconService interface {
	Ping(ctx context.Context, client *http.Client) (string, error)
}
