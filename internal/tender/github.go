package tender

import "net/http"

const (
	GithubGistDescription = "_piphos_"
)

type GithubTender struct {
	baseURL string
	client  *http.Client
	headers map[string]string
	name    string
	token   string
}

type GithubGist struct {
	ID          string                `json:"id"`
	Description string                `json:"description"`
	Public      bool                  `json:"public"`
	Files       map[string]GithubFile `json:"files"`
}

type GithubFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}
