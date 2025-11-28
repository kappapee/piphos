package beacon

import (
	"context"
	"fmt"
)

type Beacon interface {
	Ping(ctx context.Context) (string, error)
}

func New(beacon string) (Beacon, error) {
	switch beacon {
	case "haz":
		return newWeb("https://ipv4.icanhazip.com", "haz"), nil
	case "aws":
		return newWeb("https://checkip.amazonaws.com", "aws"), nil
	default:
		return nil, fmt.Errorf("unknown beacon: %s", beacon)
	}
}
