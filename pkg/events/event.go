package events

import (
	"encoding/json"
	"fmt"
	"strings"
)

type EventType string

const (
	Unknown EventType = ""

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

type Event struct {
	Type       EventType
	Repository string
	Action     string
	Ref        string

	PullRequestInfo struct {
	}
	PushInfo struct {
	}
	Sender struct {
	}

	raw json.RawMessage
}

type payloadMeta struct {
	Action     string `json:"action"`
	Ref        string `json:"ref"`
	Repository struct {
		FullName      string `json:"fullName"`
		FullNameSnake string `json:"full_name"`
	} `json:"repository"`
}

func ParseEvent(event, eventType string, body []byte) (Event, error) {
	ev := Event{
		raw: append(json.RawMessage(nil), body...),
	}

	var meta payloadMeta
	if len(body) > 0 {
		if err := json.Unmarshal(body, &meta); err != nil {
			return ev, fmt.Errorf("invalid event json: %w", err)
		}
	}

	ev.Action = meta.Action
	ev.Ref = meta.Ref
	ev.Repository = meta.Repository.FullName
	if ev.Repository == "" {
		ev.Repository = meta.Repository.FullNameSnake
	}

	ev.Type = mapEventType(event, eventType, meta.Action)

	return ev, nil
}

func mapEventType(event, eventType, action string) EventType {
	event = strings.ToLower(strings.TrimSpace(event))
	eventType = strings.ToLower(strings.TrimSpace(eventType))
	action = strings.ToLower(strings.TrimSpace(action))

	event_mapping := map[string]EventType{
		// event + event_type
		"push.push":     BranchPush,
		"create.create": BranchCreated,
		"delete.delete": BranchDeleted,
		"status.status": CICDStatus,
		"pull_request_approved.pull_request_review_approved": PullRequestReviewApproved,
		"pull_request_rejected.pull_request_review_rejected": PullRequestReviewRejected,
		"pull_request_comment.pull_request_review_comment":   PullRequestReviewComment,
		"issue_comment.pull_request_review_comment":          PullRequestComment,
		"pull_request.pull_request_review_request":           PullRequestReviewRequested,
		"pull_request.pull_request_sync":                     PullRequestSynchronized,
		// event + event_type + action
		"pull_request.pull_request.pull_request_review_request": PullRequestReviewRequested,
		"pull_request.pull_request.pull_request_sync":           PullRequestSynchronized,
		"pull_request.pull_request.opened":                      PullRequestOpened,
		"pull_request.pull_request.closed":                      PullRequestClosed,
		"pull_request.pull_request.edited":                      PullRequestEdited,
		"pull_request.pull_request.synchronized":                PullRequestSynchronized,
		"pull_request.pull_request.review_requested":            PullRequestReviewRequested,
	}

	if value, ok := event_mapping[fmt.Sprintf("%s.%s.%s", event, eventType, action)]; ok {
		return value
	}

	if value, ok := event_mapping[fmt.Sprintf("%s.%s", event, eventType)]; ok {
		return value
	}

	return Unknown
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
