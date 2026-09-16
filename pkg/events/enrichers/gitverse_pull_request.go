package enrichers

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/gitverse"
)

type prFetcher interface {
	GetPullRequest(ctx context.Context, repository string, number int) (*gitverse.PullRequest, error)
}

type GitversePullRequest struct {
	client prFetcher
}

func NewGitversePullRequest(client prFetcher) *GitversePullRequest {
	return &GitversePullRequest{client: client}
}

func (e *GitversePullRequest) Name() string {
	return "gitverse.pull_request"
}

func (e *GitversePullRequest) Enrich(ctx context.Context, event *events.Event) error {
	if event.PullRequest.Number == 0 || event.Repository == "" {
		return nil
	}

var (
		pr  *gitverse.PullRequest
		err error
	)

	if shouldAlwaysFetchPR(event.Type) {
	} else if isPREvent(event.Type) {
	} else {
		return nil
	}
	if err != nil {
		return err
	}
	if pr == nil {
		return nil
	}

	applyPullRequest(event, pr)
	return nil
}

func shouldAlwaysFetchPR(t events.EventType) bool {
	switch t {
	case events.PullRequestOpened, events.PullRequestClosed, events.PullRequestEdited,
		events.PullRequestSynchronized, events.PullRequestReviewRequested:
		return true
	default:
		return false
	}
}

func isPREvent(t events.EventType) bool {
	switch t {
	case events.PullRequestOpened, events.PullRequestClosed, events.PullRequestEdited,
		events.PullRequestSynchronized, events.PullRequestReviewRequested,
		events.PullRequestReviewApproved, events.PullRequestReviewRejected,
		events.PullRequestReviewComment, events.PullRequestComment:
		return true
	default:
		return false
	}
}

func applyPullRequest(event *events.Event, pr *gitverse.PullRequest) {
	event.PullRequest.Number = pr.Number
	event.PullRequest.Title = pr.Title
	event.PullRequest.Body = pr.Body
	event.PullRequest.State = pr.State
	event.PullRequest.Merged = pr.Merged
	event.PullRequest.URL = pr.URL

	event.PullRequest.Author = events.Actor{
		ID:    pr.User.ID,
		Name:  pr.User.Login,
		Email: pr.User.Email,
		URL:   pr.User.URL,
	}
}
