package enrichers

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/gitverse"
	"gitverse-notifier/pkg/requests"
)

type GitverseCommit struct {
	dispatcher *requests.RequestsDispatcher
}

func NewGitverseCommit(dispatcher *requests.RequestsDispatcher) *GitverseCommit {
	return &GitverseCommit{dispatcher: dispatcher}
}

func (e *GitverseCommit) Name() string {
	return "enricher.gitverse-commit"
}

func (e *GitverseCommit) Skip(event *events.Event) bool {
	return event.Type != events.BranchPush || event.Push.After == ""
}

func (e *GitverseCommit) Enrich(ctx context.Context, event *events.Event) error {

	commit, err := e.dispatcher.HandleGetCommitInfo(ctx, requests.GetCommitInfo{
		Repo:      event.Repository,
		After:     event.Push.After,
		EventType: event.Type,
	})
	if err != nil || commit == nil {
		return err
	}

	applyCommit(event, commit)

	return nil
}

func applyCommit(event *events.Event, commit *gitverse.CommitInfo) {
	event.Push.CommitTitle = commit.Commit.Message
	event.Push.URL = commit.URL
}
