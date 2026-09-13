package parsers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"gitverse-notifier/pkg/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubEnricher struct {
	name string
	err  error
	seen []*events.Event
}

func (s *stubEnricher) Name() string { return s.name }

func (s *stubEnricher) Enrich(_ context.Context, event *events.Event) error {
	cloned := *event
	s.seen = append(s.seen, &cloned)
	return s.err
}

func resetGlobalParser(t *testing.T) {
	t.Helper()
	prev := globalEventParser
	t.Cleanup(func() { globalEventParser = prev })
	globalEventParser = eventParser{}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

func TestSetupEventParser(t *testing.T) {
	resetGlobalParser(t)

	enricher := &stubEnricher{name: "stub"}
	require.NoError(t, SetupEventParser(enricher))
	require.Len(t, globalEventParser.enrichers, 1)
	assert.Equal(t, "stub", globalEventParser.enrichers[0].Name())
}

func TestParseEvent_PullRequestOpened(t *testing.T) {
	resetGlobalParser(t)

	body := mustJSON(t, map[string]any{
		"action": "opened",
		"number": 42,
		"repository": map[string]any{
			"fullName": "org/app",
		},
		"sender": map[string]any{
			"id":    7,
			"name":  "alice",
			"email": "alice@example.com",
		},
		"pullRequest": map[string]any{
			"title": "Add feature",
			"body":  "Implements APP-1",
			"state": "open",
			"user": map[string]any{
				"id":   8,
				"name": "bob",
			},
		},
	})

	event, err := ParseEvent(context.Background(), "pull_request", "pull_request", body)
	require.NoError(t, err)

	assert.Equal(t, events.PullRequestOpened, event.Type)
	assert.Equal(t, "opened", event.Action)
	assert.Equal(t, "org/app", event.Repository)
	assert.Equal(t, events.Actor{ID: 7, Name: "alice", Email: "alice@example.com"}, event.Sender)
	assert.Equal(t, 42, event.PullRequest.Number)
	assert.Equal(t, "Add feature", event.PullRequest.Title)
	assert.Equal(t, "Implements APP-1", event.PullRequest.Body)
	assert.Equal(t, "open", event.PullRequest.State)
	assert.Equal(t, events.Actor{ID: 8, Name: "bob"}, event.PullRequest.Author)
	assert.Empty(t, event.Comment.Body)
}

func TestParseEvent_PullRequestReviewComment(t *testing.T) {
	resetGlobalParser(t)

	body := mustJSON(t, map[string]any{
		"repository": map[string]any{
			"full_name": "org/app",
		},
		"sender": map[string]any{
			"name": "reviewer",
		},
		"number": 10,
		"pullRequest": map[string]any{
			"title": "Review me",
			"state": "open",
		},
		"review": map[string]any{
			"type":    "comment",
			"content": "Looks good",
		},
	})

	event, err := ParseEvent(context.Background(), "pull_request_comment", "pull_request_review_comment", body)
	require.NoError(t, err)

	assert.Equal(t, events.PullRequestReviewComment, event.Type)
	assert.Equal(t, "org/app", event.Repository)
	assert.Equal(t, "Looks good", event.Comment.Body)
	assert.Equal(t, "reviewer", event.Comment.Author.Name)
	assert.Equal(t, "Review me", event.PullRequest.Title)
}

func TestParseEvent_PullRequestReviewComment_EmptyContentStillSetsComment(t *testing.T) {
	resetGlobalParser(t)

	body := mustJSON(t, map[string]any{
		"repository": map[string]any{"fullName": "org/app"},
		"sender":     map[string]any{"name": "reviewer"},
		"number":     11,
		"pullRequest": map[string]any{
			"title": "Empty review",
		},
		"review": map[string]any{
			"content": "",
		},
	})

	event, err := ParseEvent(context.Background(), "pull_request_comment", "pull_request_review_comment", body)
	require.NoError(t, err)
	assert.Equal(t, events.PullRequestReviewComment, event.Type)
	assert.Equal(t, "", event.Comment.Body)
	assert.Equal(t, "reviewer", event.Comment.Author.Name)
}

func TestParseEvent_IssueComment(t *testing.T) {
	resetGlobalParser(t)

	body := mustJSON(t, map[string]any{
		"repository": map[string]any{"fullName": "org/app"},
		"sender":     map[string]any{"name": "carol"},
		"isPull":     true,
		"issue": map[string]any{
			"title": "PR title",
			"body":  "PR body",
			"state": "open",
			"user":  map[string]any{"name": "author"},
		},
		"comment": map[string]any{
			"body": "Please fix",
			"user": map[string]any{"name": "commenter"},
		},
	})

	event, err := ParseEvent(context.Background(), "issue_comment", "pull_request_comment", body)
	require.NoError(t, err)

	assert.Equal(t, events.PullRequestComment, event.Type)
	assert.Equal(t, "Please fix", event.Comment.Body)
	assert.Equal(t, "commenter", event.Comment.Author.Name)
	assert.Equal(t, "PR title", event.PullRequest.Title)
	assert.Equal(t, "author", event.PullRequest.Author.Name)
}

func TestParseEvent_IssueComment_FallsBackToSenderAuthor(t *testing.T) {
	resetGlobalParser(t)

	body := mustJSON(t, map[string]any{
		"repository": map[string]any{"fullName": "org/app"},
		"sender":     map[string]any{"name": "carol"},
		"comment": map[string]any{
			"body": "no user",
		},
	})

	event, err := ParseEvent(context.Background(), "issue_comment", "pull_request_comment", body)
	require.NoError(t, err)
	assert.Equal(t, "carol", event.Comment.Author.Name)
	assert.Empty(t, event.PullRequest.Title)
}

func TestParseEvent_BranchPush(t *testing.T) {
	resetGlobalParser(t)

	body := mustJSON(t, map[string]any{
		"ref":          "refs/heads/main",
		"before":       "aaa",
		"after":        "bbb",
		"totalCommits": 3,
		"repository": map[string]any{
			"fullName": "org/app",
		},
		"pusher": map[string]any{
			"name": "dave",
		},
	})

	event, err := ParseEvent(context.Background(), "push", "push", body)
	require.NoError(t, err)

	assert.Equal(t, events.BranchPush, event.Type)
	assert.Equal(t, "refs/heads/main", event.Ref)
	assert.Equal(t, "main", event.Branch)
	assert.Equal(t, "dave", event.Sender.Name)
	assert.Equal(t, events.PushInfo{Before: "aaa", After: "bbb", TotalCommits: 3}, event.Push)
}

func TestParseEvent_BranchCreatedAndDeleted(t *testing.T) {
	resetGlobalParser(t)

	createdBody := mustJSON(t, map[string]any{
		"ref": "refs/heads/feature",
		"repository": map[string]any{
			"fullName": "org/app",
		},
		"sender": map[string]any{"name": "erin"},
	})
	created, err := ParseEvent(context.Background(), "create", "create", createdBody)
	require.NoError(t, err)
	assert.Equal(t, events.BranchCreated, created.Type)
	assert.Equal(t, "feature", created.Branch)

	deletedBody := mustJSON(t, map[string]any{
		"ref": "refs/heads/old",
		"repository": map[string]any{
			"fullName": "org/app",
		},
		"sender": map[string]any{"name": "frank"},
	})
	deleted, err := ParseEvent(context.Background(), "delete", "delete", deletedBody)
	require.NoError(t, err)
	assert.Equal(t, events.BranchDeleted, deleted.Type)
	assert.Equal(t, "old", deleted.Branch)
}

func TestParseEvent_CICDStatus(t *testing.T) {
	resetGlobalParser(t)

	body := mustJSON(t, map[string]any{
		"state":       "success",
		"context":     "ci/tests",
		"description": "All good",
		"sha":         "deadbeef",
		"repository": map[string]any{
			"fullName": "org/app",
		},
		"sender": map[string]any{"name": "bot"},
	})

	event, err := ParseEvent(context.Background(), "status", "status", body)
	require.NoError(t, err)

	assert.Equal(t, events.CICDStatus, event.Type)
	assert.Equal(t, events.StatusInfo{
		State:       "success",
		Context:     "ci/tests",
		Description: "All good",
		SHA:         "deadbeef",
	}, event.Status)
}

func TestParseEvent_PullRequestAuthorFallsBackToSender(t *testing.T) {
	resetGlobalParser(t)

	body := mustJSON(t, map[string]any{
		"action": "closed",
		"number": 5,
		"repository": map[string]any{
			"fullName": "org/app",
		},
		"sender": map[string]any{
			"name": "alice",
		},
		"pullRequest": map[string]any{
			"title": "Close me",
			"state": "closed",
		},
	})

	event, err := ParseEvent(context.Background(), "pull_request", "pull_request", body)
	require.NoError(t, err)
	assert.Equal(t, events.PullRequestClosed, event.Type)
	assert.Equal(t, "alice", event.PullRequest.Author.Name)
}

func TestParseEvent_UsesActionFromBodyForTypeResolution(t *testing.T) {
	resetGlobalParser(t)

	cases := []struct {
		action string
		want   events.EventType
	}{
		{action: "opened", want: events.PullRequestOpened},
		{action: "edited", want: events.PullRequestEdited},
		{action: "synchronized", want: events.PullRequestSynchronized},
		{action: "review_requested", want: events.PullRequestReviewRequested},
	}

	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			body := mustJSON(t, map[string]any{
				"action": tc.action,
				"repository": map[string]any{
					"fullName": "org/app",
				},
				"sender": map[string]any{"name": "alice"},
				"pullRequest": map[string]any{
					"title": "x",
				},
			})

			event, err := ParseEvent(context.Background(), "pull_request", "pull_request", body)
			require.NoError(t, err)
			assert.Equal(t, tc.want, event.Type)
		})
	}
}

func TestParseEvent_Errors(t *testing.T) {
	resetGlobalParser(t)

	t.Run("invalid common json", func(t *testing.T) {
		_, err := ParseEvent(context.Background(), "push", "push", []byte("{"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse event json")
	})

	t.Run("unknown event type", func(t *testing.T) {
		body := mustJSON(t, map[string]any{
			"repository": map[string]any{"fullName": "org/app"},
		})
		_, err := ParseEvent(context.Background(), "weird", "event", body)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to detect event type")
	})

	t.Run("invalid pull request json after type resolve", func(t *testing.T) {
		// Valid JSON for common fields, but nested types that break PR unmarshal
		// are hard because common and PR both use json.Unmarshal on same body.
		// Use a body that parses as common but fails as PR via wrong types.
		body := []byte(`{"action":"opened","repository":{"fullName":"org/app"},"number":"not-int"}`)
		_, err := ParseEvent(context.Background(), "pull_request", "pull_request", body)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse pull request json")
	})

	t.Run("invalid issue comment json", func(t *testing.T) {
		body := []byte(`{"repository":{"fullName":"org/app"},"isPull":"nope"}`)
		_, err := ParseEvent(context.Background(), "issue_comment", "pull_request_comment", body)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse issue comment json")
	})

	t.Run("invalid push json", func(t *testing.T) {
		body := []byte(`{"repository":{"fullName":"org/app"},"totalCommits":"x"}`)
		_, err := ParseEvent(context.Background(), "push", "push", body)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse push/ref json")
	})

	t.Run("invalid status json", func(t *testing.T) {
		body := []byte(`{"repository":{"fullName":"org/app"},"state":{"nested":true}}`)
		_, err := ParseEvent(context.Background(), "status", "status", body)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse status json")
	})
}

func TestParseEvent_RunsEnrichersAndIgnoresEnricherErrors(t *testing.T) {
	resetGlobalParser(t)

	ok := &stubEnricher{name: "ok"}
	failing := &stubEnricher{name: "failing", err: errors.New("boom")}
	require.NoError(t, SetupEventParser(ok, failing))

	body := mustJSON(t, map[string]any{
		"ref": "refs/heads/main",
		"repository": map[string]any{
			"fullName": "org/app",
		},
		"sender": map[string]any{"name": "alice"},
	})

	event, err := ParseEvent(context.Background(), "push", "push", body)
	require.NoError(t, err)
	assert.Equal(t, events.BranchPush, event.Type)

	require.Len(t, ok.seen, 1)
	require.Len(t, failing.seen, 1)
	assert.Equal(t, "org/app", ok.seen[0].Repository)
	assert.Equal(t, events.BranchPush, failing.seen[0].Type)
}

func TestParseEvent_EmptyBodyFailsOnTypedPayload(t *testing.T) {
	resetGlobalParser(t)

	// fillCommon accepts an empty body, but type-specific fillers still require JSON.
	_, err := ParseEvent(context.Background(), "push", "push", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse push/ref json")
}

func TestParseEvent_ReviewApprovedAndRejected(t *testing.T) {
	resetGlobalParser(t)

	approvedBody := mustJSON(t, map[string]any{
		"repository": map[string]any{"fullName": "org/app"},
		"sender":     map[string]any{"name": "approver"},
		"number":     3,
		"pullRequest": map[string]any{
			"title": "Ready",
			"state": "open",
		},
	})
	approved, err := ParseEvent(context.Background(), "pull_request_approved", "pull_request_review_approved", approvedBody)
	require.NoError(t, err)
	assert.Equal(t, events.PullRequestReviewApproved, approved.Type)

	rejectedBody := mustJSON(t, map[string]any{
		"repository": map[string]any{"fullName": "org/app"},
		"sender":     map[string]any{"name": "reviewer"},
		"number":     4,
		"pullRequest": map[string]any{
			"title": "Needs work",
			"state": "open",
		},
	})
	rejected, err := ParseEvent(context.Background(), "pull_request_rejected", "pull_request_review_rejected", rejectedBody)
	require.NoError(t, err)
	assert.Equal(t, events.PullRequestReviewRejected, rejected.Type)
}
