package beacon

import "context"

type Beacon interface {
	Name() string
	Ping(ctx context.Context) (string, error)
	URL() string
}
