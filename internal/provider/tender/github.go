package tender

import (
	"context"

	"github.com/kappapee/piphos/internal/service"
)

type TenderGH service.Tender

func NewTenderGH() *TenderGH {
	return &TenderGH{
		Name: "github",
		URL:  "https://api.github.com/gists",
	}
}

func (t *TenderGH) Push(ctx context.Context, hostname, ip string) error {
	return nil
}

func (t *TenderGH) Pull(ctx context.Context) (map[string]string, error) {
	results := make(map[string]string)
	return results, nil
}

type GitHubFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type GitHubGist struct {
	ID          string                `json:"id"`
	Description string                `json:"description"`
	Public      bool                  `json:"public"`
	Files       map[string]GitHubFile `json:"files"`
}
