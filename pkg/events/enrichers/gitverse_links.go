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
	base := strings.TrimRight(strings.TrimSpace(config.GlobalConfig.GitverseBaseURL), "/")
	return &GitverseLinks{baseURL: base}
}

func (e *GitverseLinks) Name() string {
	return "gitverse.links"
}

func (e *GitverseLinks) Enrich(event *events.Event) error {
	if e.baseURL == "" {
		return nil
	}

	event.Sender = e.withUserURL(event.Sender)
	event.PullRequest.Author = e.withUserURL(event.PullRequest.Author)
	event.Comment.Author = e.withUserURL(event.Comment.Author)

	if event.Repository == "" {
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

func (e *GitverseLinks) withUserURL(actor events.Actor) events.Actor {
	name := strings.TrimSpace(actor.Name)
	if name == "" {
		return actor
	}
	actor.URL = fmt.Sprintf("%s/%s", e.baseURL, url.PathEscape(name))
	return actor
}
