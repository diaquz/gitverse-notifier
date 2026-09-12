package parsers

import (
	"testing"

	"gitverse-notifier/pkg/events"
)

func TestParseEventPush(t *testing.T) {
	body := []byte(`{
		"ref":"refs/heads/develop",
		"before":"aaa",
		"after":"bbb",
		"totalCommits":2,
		"repository":{"fullName":"bks-ruby/control-objects"},
		"sender":{"id":1,"name":"alice","email":"a@example.com"}
	}`)

	ev, err := ParseEvent("push", "push", body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != events.BranchPush {
		t.Fatalf("Type=%q", ev.Type)
	}
	if ev.Repository != "bks-ruby/control-objects" {
		t.Fatalf("Repository=%q", ev.Repository)
	}
	if ev.Branch != "develop" {
		t.Fatalf("Branch=%q", ev.Branch)
	}
	if ev.Sender.Name != "alice" || ev.Sender.ID != 1 {
		t.Fatalf("Sender=%+v", ev.Sender)
	}
	if ev.Push.Before != "aaa" || ev.Push.After != "bbb" || ev.Push.TotalCommits != 2 {
		t.Fatalf("Push=%+v", ev.Push)
	}
	if ev.PullRequest != (events.PullRequestInfo{}) {
		t.Fatalf("PullRequest should be empty, got %+v", ev.PullRequest)
	}
}

func TestParseEventPullRequestOpened(t *testing.T) {
	body := []byte(`{
		"action":"opened",
		"number":41,
		"pullRequest":{
			"title":"STKPVT-686 fix",
			"body":"from STKPVT-686",
			"state":"open",
			"user":{"id":45,"name":"nkim","email":"n@ex.com"}
		},
		"repository":{"fullName":"org/repo"},
		"sender":{"id":45,"name":"nkim","email":"n@ex.com"}
	}`)

	ev, err := ParseEvent("pull_request", "pull_request", body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != events.PullRequestOpened {
		t.Fatalf("Type=%q", ev.Type)
	}
	if ev.PullRequest.Number != 41 || ev.PullRequest.Title != "STKPVT-686 fix" {
		t.Fatalf("PullRequest=%+v", ev.PullRequest)
	}
	if ev.PullRequest.Author.Name != "nkim" {
		t.Fatalf("Author=%+v", ev.PullRequest.Author)
	}
	if ev.Sender.Name != "nkim" {
		t.Fatalf("Sender=%+v", ev.Sender)
	}
	if ev.Comment != (events.CommentInfo{}) {
		t.Fatalf("Comment should be empty, got %+v", ev.Comment)
	}
}

func TestParseEventIssueComment(t *testing.T) {
	body := []byte(`{
		"action":"created",
		"isPull":true,
		"issue":{
			"title":"STKPVT-686 title",
			"body":"desc",
			"state":"open",
			"user":{"id":45,"name":"nkim"}
		},
		"comment":{
			"body":"hello",
			"user":{"id":40,"name":"reviewer"}
		},
		"repository":{"fullName":"org/repo"},
		"sender":{"id":40,"name":"reviewer"}
	}`)

	ev, err := ParseEvent("issue_comment", "pull_request_comment", body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != events.PullRequestComment {
		t.Fatalf("Type=%q", ev.Type)
	}
	if ev.Comment.Body != "hello" || ev.Comment.Author.Name != "reviewer" {
		t.Fatalf("Comment=%+v", ev.Comment)
	}
	if ev.PullRequest.Title != "STKPVT-686 title" {
		t.Fatalf("PullRequest=%+v", ev.PullRequest)
	}
}

func TestParseEventMissingSender(t *testing.T) {
	body := []byte(`{
		"ref":"refs/heads/main",
		"repository":{"full_name":"org/repo"}
	}`)

	ev, err := ParseEvent("push", "", body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != events.BranchPush {
		t.Fatalf("Type=%q", ev.Type)
	}
	if ev.Sender != (events.Actor{}) {
		t.Fatalf("Sender should be empty, got %+v", ev.Sender)
	}
	if ev.Repository != "org/repo" {
		t.Fatalf("Repository=%q", ev.Repository)
	}
	if ev.Branch != "main" {
		t.Fatalf("Branch=%q", ev.Branch)
	}
}
