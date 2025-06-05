package main

import (
	"os"
)

const (
	BeaconHaz    = "haz"
	BeaconAws    = "aws"
	TenderGithub = "github"
	TenderGitlab = "gitlab"
)

var TenderToken = os.Getenv("PIPHOS_TOKEN")

var BeaconConfig = map[string]Beacon{
	BeaconHaz: {Name: "icanhazip", URL: "https://ipv4.icanhazip.com"},
	BeaconAws: {Name: "aws", URL: "https://checkip.amazonaws.com"},
}

var TenderConfig = map[string]Tender{
	TenderGithub: {
		Name: TenderGithub,
		URL:  "https://api.github.com/gists",
		Headers: map[string]string{
			"X-GitHub-Api-Version": "2022-11-28",
			"Content-Type":         "application/json",
		},
	},
	TenderGitlab: {
		Name: TenderGitlab,
		URL:  "https://gitlab.com/api/v4/snippets",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	},
}
