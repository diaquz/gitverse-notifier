package enrichers

import (
	"context"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/gitverse"
	gvqueries "gitverse-notifier/pkg/queries/gitverse"
)

type GitversePullRequest struct {
	queries *gvqueries.Queries
}

func NewGitversePullRequest(queries *gvqueries.Queries) *GitversePullRequest {
	return &GitversePullRequest{queries: queries}
}

func (e *GitversePullRequest) Name() string {
	return "gitverse.pull_request"
}

func (e *GitversePullRequest) Skip(event *events.Event) bool {
	switch event.Type {
	case events.PullRequestOpened, events.PullRequestClosed, events.PullRequestEdited,
		events.PullRequestSynchronized, events.PullRequestReviewRequested, events.PullRequestReviewApproved, events.PullRequestReviewRejected,
		events.PullRequestReviewComment, events.PullRequestComment:
		return false
	default:
		return true
	}
}

func (e *GitversePullRequest) Enrich(ctx context.Context, event *events.Event) error {
	if event.Repository == "" || event.PullRequest.Title == "" {
		return nil
	}

	pr, err := e.queries.GetPullRequest(ctx, gvqueries.GetPullRequest{
		Repo:      event.Repository,
		Title:     event.PullRequest.Title,
		Number:    event.PullRequest.Number,
		EventType: event.Type,
	})
	if err != nil || pr == nil {
		return err
	}

	applyPullRequest(event, pr)
	return nil
}

func applyPullRequest(event *events.Event, pr *gitverse.PullRequest) {
	event.PullRequest.Number = pr.Number
	event.PullRequest.Title = pr.Title
	event.PullRequest.Body = pr.Body
	event.PullRequest.State = pr.State
	event.PullRequest.Merged = pr.Merged
	event.PullRequest.URL = pr.URL

	if len(pr.RequestedReviewers) > 0 {
		event.PullRequest.Reviewers = make([]events.Actor, 0, len(pr.RequestedReviewers))
		for _, reviewer := range pr.RequestedReviewers {
			actor := events.Actor{
				ID:    reviewer.ID,
				Name:  reviewer.Name,
				Email: reviewer.Email,
				URL:   reviewer.URL,
			}
			event.PullRequest.Reviewers = append(event.PullRequest.Reviewers, actor)
		}
	}

	event.PullRequest.Author = events.Actor{
		ID:    pr.User.ID,
		Name:  pr.User.Login,
		Email: pr.User.Email,
		URL:   pr.User.URL,
	}
}
