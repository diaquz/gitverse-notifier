package queries

import "gitverse-notifier/pkg/events"

type cacheMode int
const (
	skip cacheMode = iota
	preferCache
	forceRefresh
)

func pullRequestCacheMode(t events.EventType) cacheMode {
	switch t {
	case events.PullRequestOpened, events.PullRequestClosed, events.PullRequestEdited,
		events.PullRequestSynchronized, events.PullRequestReviewRequested:
		return forceRefresh
	case events.PullRequestReviewApproved, events.PullRequestReviewRejected,
		events.PullRequestReviewComment, events.PullRequestComment:
		return preferCache
	default:
		return skip
	}
}
