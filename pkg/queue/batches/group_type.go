package batches

import (
	"fmt"
	"gitverse-notifier/pkg/events"
)

type GroupType interface {
	Name() string
	Key(event *events.Event) string
}

type PullRequestBatch struct{}

func (PullRequestBatch) Name() string { return "pull_request_batch" }

func (PullRequestBatch) Key(event *events.Event) string {
	return fmt.Sprintf("%s/", event.Repository, event.PullRequest.Number)
}
