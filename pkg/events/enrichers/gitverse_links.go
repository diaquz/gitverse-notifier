package enrichers

import (
	"fmt"
	"net/url"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/repositories"
)

type GitverseLinks struct {
	manager *repositories.RepositoriesManager
}

func NewGitverseLinks(manager *repositories.RepositoriesManager) *GitverseLinks {
	return &GitverseLinks{manager: manager}
}

func (e *GitverseLinks) Name() string {
	return "gitverse.links"
}

func (e *GitverseLinks) Enrich(event *events.Event) error {
	baseURL := e.baseURLFor(event.Repository)
	if baseURL == "" {
		return nil
	}

	event.Sender = e.withUserURL(baseURL, event.Sender)
	event.PullRequest.Author = e.withUserURL(baseURL, event.PullRequest.Author)
	event.Comment.Author = e.withUserURL(baseURL, event.Comment.Author)

	if event.Repository == "" {
		return nil
	}

	repoPath := strings.Trim(event.Repository, "/")
	event.RepositoryURL = fmt.Sprintf("%s/%s", baseURL, repoPath)

	if event.Branch != "" {
		event.BranchURL = fmt.Sprintf("%s/%s/src/branch/%s", baseURL, repoPath, url.PathEscape(event.Branch))
	}

	if event.PullRequest.Number > 0 {
		event.PullRequest.URL = fmt.Sprintf("%s/%s/pulls/%d", baseURL, repoPath, event.PullRequest.Number)
	}

	return nil
}

func (e *GitverseLinks) baseURLFor(repository string) string {
	settings := e.manager.SettingsByRepository(repository)

	return strings.TrimRight(strings.TrimSpace(settings.Url), "/")
}

func (e *GitverseLinks) withUserURL(baseURL string, actor events.Actor) events.Actor {
	name := strings.TrimSpace(actor.Name)
	if name == "" {
		return actor
	}

	actor.URL = fmt.Sprintf("%s/%s", baseURL, url.PathEscape(name))
	return actor
}
