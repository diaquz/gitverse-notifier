package events

type EventType int8

const (
	Any EventType = iota
	PullRequestOpened
	PullRequestReviewRequested
	PullRequestReviewComment
	PullRequestComment
	PullRequestApproved
	PullRequestClosed
	BranchPush
	BranchCreate
	CICDStatusUpdated
)

type Event struct {
	Type EventType

}

func ParseEvent(payload string) (Event, error) {
	return Event{}, nil
}
