package requests

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/gitverse"
	"gitverse-notifier/pkg/logger"
)

type GetPullRequestInfo struct {
	Repo      string
	Title     string
	Number    int
	EventType events.EventType
}

func (d *RequestsDispatcher) HandleGetPullRequestInfo(ctx context.Context, request GetPullRequestInfo) (*gitverse.PullRequest, error) {
	if shouldUpdatePRCache(request.EventType) {
		return d.fetchAndCache(ctx, request)
	}

	if pr, ok := d.prs.GetByTitle(request.Repo, request.Title, request.Number); ok {
		logger.Debug(ctx, "fetched PR info from cache",
			"action", "request.get-pull-request-info",
			"repository", request.Repo, "pr-number", pr.Number, "pr-title", pr.Title)

		return pr, nil
	}

	return d.fetchAndCache(ctx, request)
}

func (d *RequestsDispatcher) fetchAndCache(ctx context.Context, request GetPullRequestInfo) (*gitverse.PullRequest, error) {
	number := request.Number
	if number <= 0 {
		number = d.prs.GetPRNumber(request.Repo, request.Title)
		if number <= 0 {
			return nil, nil
		}
	}

	pr, err := d.client.GetPullRequest(ctx, request.Repo, number)
	if err != nil {
		return nil, err
	}

	d.prs.Set(request.Repo, pr.Title, number, pr)

	return pr, nil
}

func shouldUpdatePRCache(t events.EventType) bool {
	switch t {
	case events.PullRequestOpened, events.PullRequestClosed, events.PullRequestEdited,
		events.PullRequestSynchronized, events.PullRequestReviewRequested:
		return true
	default:
		return false
	}
}
