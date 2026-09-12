package parsers

import (
	"encoding/json"
	"fmt"

	"gitverse-notifier/pkg/events"
)

type pullRequestPayload struct {
	Number      int `json:"number"`
	PullRequest struct {
		Title string       `json:"title"`
		Body  string       `json:"body"`
		State string       `json:"state"`
		User  actorPayload `json:"user"`
	} `json:"pullRequest"`
	Review struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	} `json:"review"`
	Sender actorPayload `json:"sender"`
}

func fillPullRequest(event *events.Event, body []byte) error {
	var payload pullRequestPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to parse pull request json: %w", err)
	}

	event.PullRequest = events.PullRequestInfo{
		Number: payload.Number,
		Title:  payload.PullRequest.Title,
		Body:   payload.PullRequest.Body,
		State:  payload.PullRequest.State,
		Author: payload.PullRequest.User.toActor(),
	}

	if event.PullRequest.Author == (events.Actor{}) {
		event.PullRequest.Author = event.Sender
	}

	// Для события PullRequestReviewComment приходят пустые сообщения, учитываем их
	if payload.Review.Content != "" || event.Type == events.PullRequestReviewComment {
		event.Comment = events.CommentInfo{
			Body:   payload.Review.Content,
			Author: event.Sender,
		}
	}

	return nil
}
