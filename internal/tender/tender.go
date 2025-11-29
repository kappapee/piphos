package tender

import (
	"context"
	"fmt"
	"os"

	"github.com/kappapee/piphos/internal/validate"
)

type Tender interface {
	Pull(ctx context.Context) (map[string]string, error)
	Push(ctx context.Context, hostname, ip string) error
}

func New(tender string) (Tender, error) {
	token := os.Getenv("PIPHOS_TOKEN")
	err := validate.Token(token)
	if err != nil {
		return nil, err
	}
	switch tender {
	case "gh":
		return newGithub(token), nil
	default:
		return nil, fmt.Errorf("unknown tender: %s", tender)
	}
}
