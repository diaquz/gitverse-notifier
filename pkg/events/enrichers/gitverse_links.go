package enrichers

import (
	"fmt"
	"net/url"
	"strings"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events"
)

type GitverseLinks struct {
	baseURL string
}

func NewGitverseLinks() *GitverseLinks {
	base := strings.TrimRight(config.GlobalConfig.GitverseBaseURL, "/")
	return &GitverseLinks{baseURL: base}
}

func (e *GitverseLinks) Name() string {
	return "gitverse.links"
}

func (e *GitverseLinks) Enrich(event *events.Event) error {
	if event.Repository == "" || e.baseURL == "" {
		return nil
	}

	repoPath := strings.Trim(event.Repository, "/")
	event.RepositoryURL = fmt.Sprintf("%s/%s", e.baseURL, repoPath)

	if event.Branch != "" {
		event.BranchURL = fmt.Sprintf("%s/%s/src/branch/%s", e.baseURL, repoPath, url.PathEscape(event.Branch))
	}

	if event.PullRequest.Number > 0 {
		event.PullRequest.URL = fmt.Sprintf("%s/%s/pulls/%d", e.baseURL, repoPath, event.PullRequest.Number)
	}

	return nil
}
