package events

import (
	"fmt"
	"strings"
)

type EventType string

const (
	Unknown EventType = "unknown"

	PullRequestOpened          EventType = "pull_request.opened"
	PullRequestClosed          EventType = "pull_request.closed"
	PullRequestEdited          EventType = "pull_request.edited"
	PullRequestSynchronized    EventType = "pull_request.synchronized"
	PullRequestReviewRequested EventType = "pull_request.review_requested"
	PullRequestReviewApproved  EventType = "pull_request.review_approved"
	PullRequestReviewRejected  EventType = "pull_request.review_rejected"
	PullRequestReviewComment   EventType = "pull_request.review_comment"
	PullRequestComment         EventType = "pull_request.comment"

	BranchPush    EventType = "branch.push"
	BranchCreated EventType = "branch.created"
	BranchDeleted EventType = "branch.deleted"

	CICDStatus EventType = "cicd.status"
)

type Actor struct {
	ID    int64
	Name  string
	Email string
	URL   string
	TgTag string
}

type PullRequestInfo struct {
	Number    int
	Title     string
	Body      string
	State     string
	Merged    bool
	Author    Actor
	URL       string
	Reviewers []Actor
}

type CommentInfo struct {
	Body   string
	Author Actor
}

type PushInfo struct {
	Before       string
	After        string
	TotalCommits int
	CommitTitle  string
	URL          string
}

type StatusInfo struct {
	State       string
	Context     string
	Description string
	SHA         string
}

type Event struct {
	Type          EventType
	Action        string
	Ref           string
	Branch        string
	BranchURL     string
	Repository    string
	RepositoryURL string
	Sender        Actor
	PullRequest   PullRequestInfo
	Comment       CommentInfo
	Push          PushInfo
	Status        StatusInfo
	IssueKeys     []string
}

func ResolveEventType(event, eventType, action string) (EventType, error) {
	event = strings.ToLower(strings.TrimSpace(event))
	eventType = strings.ToLower(strings.TrimSpace(eventType))
	action = strings.ToLower(strings.TrimSpace(action))

	eventMapping := map[string]EventType{
		"push.push":     BranchPush,
		"create.create": BranchCreated,
		"delete.delete": BranchDeleted,
		"status.status": CICDStatus,

		"pull_request_approved.pull_request_review_approved": PullRequestReviewApproved,
		"pull_request_rejected.pull_request_review_rejected": PullRequestReviewRejected,
		"pull_request_comment.pull_request_review_comment":   PullRequestReviewComment,
		"issue_comment.pull_request_comment":                 PullRequestComment,
		"pull_request.pull_request_review_request":           PullRequestReviewRequested,
		"pull_request.pull_request_sync":                     PullRequestSynchronized,

		"pull_request.pull_request.opened":           PullRequestOpened,
		"pull_request.pull_request.closed":           PullRequestClosed,
		"pull_request.pull_request.edited":           PullRequestEdited,
		"pull_request.pull_request.synchronized":     PullRequestSynchronized,
		"pull_request.pull_request.review_requested": PullRequestReviewRequested,
	}

	if value, ok := eventMapping[fmt.Sprintf("%s.%s.%s", event, eventType, action)]; ok {
		return value, nil
	}

	if value, ok := eventMapping[fmt.Sprintf("%s.%s", event, eventType)]; ok {
		return value, nil
	}

	return Unknown, fmt.Errorf("failed to detect event type for %s.%s, action = %s", event, eventType, action)
}

func ParseEventType(s string) (EventType, bool) {
	s = strings.TrimSpace(s)
	switch EventType(s) {
	case PullRequestOpened, PullRequestClosed, PullRequestEdited, PullRequestSynchronized,
		PullRequestReviewRequested, PullRequestReviewApproved, PullRequestReviewRejected,
		PullRequestReviewComment, PullRequestComment,
		BranchPush, BranchCreated, BranchDeleted, CICDStatus:
		return EventType(s), true
	default:
		return Unknown, false
	}
}

func BranchName(ref string) string {
	ref = strings.TrimSpace(ref)
	for _, prefix := range []string{"refs/heads/", "refs/tags/"} {
		if strings.HasPrefix(ref, prefix) {
			return strings.TrimPrefix(ref, prefix)
		}
	}
	return ref
}

func (p PullRequestInfo) EffectiveState() string {
	if p.Merged {
		return "merged"
	}

	return p.State
}
