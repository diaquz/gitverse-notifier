package requests

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

func (d *RequestsDispatcher) HandleGetCommitInfo(ctx context.Context, request GetCommitInfo) (*gitverse.CommitInfo, error) {
	return d.client.GetCommit(ctx, request.Repo, request.After)
}
