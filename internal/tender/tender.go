package tender

import "context"

type Tender interface {
	Pull(ctx context.Context) (map[string]string, error)
	Push(ctx context.Context, hostname, ip string) error
}
