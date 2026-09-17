package queries

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/gitverse"
)

type GetCommitInfo struct {
	Repo      string
	After     string
	EventType events.EventType
}

func (q *Queries) GetCommitInfo(ctx context.Context, query GetCommitInfo) (*gitverse.CommitInfo, error) {
	commit, err := q.client.GetCommit(ctx, query.Repo, query.After)
	return commit, err
}
