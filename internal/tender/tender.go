package tender

import (
	"context"
	"fmt"

	"github.com/kappapee/piphos/internal/validate"
)

type Tender interface {
	Pull(ctx context.Context) (map[string]string, error)
	Push(ctx context.Context, hostname, ip string) error
}

func New(tender string) (Tender, error) {
	switch tender {
	case "gh":
		t := newGithub()
		err := validate.Token(t.token)
		if err != nil {
			return nil, err
		}
		return t, nil
	default:
		return nil, fmt.Errorf("unknown tender: %s", tender)
	}
}
