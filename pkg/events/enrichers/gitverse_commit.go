package enrichers

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/gitverse"
	gvqueries "gitverse-notifier/pkg/queries/gitverse"
)

type GitverseCommit struct {
	queries *gvqueries.Queries
}

func NewGitverseCommit(queries *gvqueries.Queries) *GitverseCommit {
	return &GitverseCommit{queries: queries}
}

func (e *GitverseCommit) Name() string {
	return "gitverse.commi"
}

func (e *GitverseCommit) Skip(event *events.Event) bool {
	return event.Type != events.BranchPush || event.Push.After == ""
}

func (e *GitverseCommit) Enrich(ctx context.Context, event *events.Event) error {

	commit, err := e.queries.GetCommitInfo(ctx, gvqueries.GetCommitInfo{
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
