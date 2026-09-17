package queries

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/gitverse"
)

type GetPullRequest struct {
	Repo      string
	Title     string
	Number    int
	EventType events.EventType
}

func (q *Queries) GetPullRequest(ctx context.Context, query GetPullRequest) (*gitverse.PullRequest, error) {
	switch pullRequestCacheMode(query.EventType) {
	case skip:
		return nil, nil
	case preferCache:
		if pr, ok := q.prs.GetByTitle(query.Repo, query.Title, query.Number); ok {
			return pr, nil
		}
		return q.fetchAndCache(ctx, query)
	case forceRefresh:
		return q.fetchAndCache(ctx, query)
	default:
		return nil, nil
	}
}

func (q *Queries) fetchAndCache(ctx context.Context, query GetPullRequest) (*gitverse.PullRequest, error) {
	number := query.Number
	if number <= 0 {
		number = q.prs.GetPRNumber(query.Repo, query.Title)
		if number <= 0 {
			return nil, nil
		}
	}

	pr, err := q.client.GetPullRequest(ctx, query.Repo, number)
	if err != nil {
		return nil, err
	}

	q.prs.Set(query.Repo, pr.Title, number, pr)
	return pr, nil
}
