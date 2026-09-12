package parsers

import (
	"encoding/json"
	"fmt"

	"gitverse-notifier/pkg/events"
)

type issueCommentPayload struct {
	IsPull bool `json:"isPull"`
	Issue  struct {
		Title string       `json:"title"`
		Body  string       `json:"body"`
		State string       `json:"state"`
		User  actorPayload `json:"user"`
	} `json:"issue"`
	Comment struct {
		Body string       `json:"body"`
		User actorPayload `json:"user"`
	} `json:"comment"`
	Sender actorPayload `json:"sender"`
}

func fillIssueComment(event *events.Event, body []byte) error {
	var payload issueCommentPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to parse issue comment json: %w", err)
	}

	event.Comment = events.CommentInfo{
		Body:   payload.Comment.Body,
		Author: payload.Comment.User.toActor(),
	}
	if event.Comment.Author == (events.Actor{}) {
		event.Comment.Author = payload.Sender.toActor()
	}

	if payload.IsPull || payload.Issue.Title != "" {
		event.PullRequest = events.PullRequestInfo{
			Title:  payload.Issue.Title,
			Body:   payload.Issue.Body,
			State:  payload.Issue.State,
			Author: payload.Issue.User.toActor(),
		}
	}

	return nil
}
