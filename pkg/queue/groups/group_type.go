package groups

import (
	"fmt"
	"gitverse-notifier/pkg/events"
)

type GroupType interface {
	Name() string
	Key(event *events.Event) string
}

type PullRequestGroup struct{}

func (PullRequestGroup) Name() string { return "group.pull_request" }

func (PullRequestGroup) Key(event *events.Event) string {
	return fmt.Sprintf("%s|%d", event.Repository, event.PullRequest.Number)
}
