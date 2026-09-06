package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/jira"
	"gitverse-notifier/pkg/logger"
)

var issueKeyRe = regexp.MustCompile(`\b([A-Z][A-Z0-9]+-\d+)\b`)

type JiraCommentIssue struct {
	Client *jira.JiraClient
}

func NewJiraCommentIssue(client *jira.JiraClient) *JiraCommentIssue {
	return &JiraCommentIssue{Client: client}
}

func (a *JiraCommentIssue) Name() string {
	return "jira.comment_issue"
}

func (a *JiraCommentIssue) Run(ctx context.Context, ev events.Event, rule events.ActionRule) error {
	_ = ctx
	_ = rule

	if a.Client == nil {
		return jira.ErrNotConfigured
	}

	pr, err := decodePRPayload(ev.Raw)
	if err != nil {
		return err
	}

	keys := uniqueIssueKeys(pr.PullRequest.Title, pr.PullRequest.Body)
	if len(keys) == 0 {
		logger.Infof("jira.comment_issue: no issue keys in PR title/body (repo=%s)", ev.Repository)
		return nil
	}

	author := pr.PullRequest.User.Name
	if author == "" {
		author = pr.Sender.Name
	}

	comment := jira.IssueComment{
		Repository:  ev.Repository,
		Title:       pr.PullRequest.Title,
		Author:      author,
		AuthorLogin: author,
		Action:      ev.Action,
	}

	for _, key := range keys {
		if _, err := a.Client.AddPRComment(key, comment); err != nil {
			return fmt.Errorf("comment on %s: %w", key, err)
		}
		logger.Infof("jira.comment_issue: commented on %s for repo=%s", key, ev.Repository)
	}
	return nil
}

type prPayload struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		User  struct {
			Name string `json:"name"`
		} `json:"user"`
	} `json:"pullRequest"`
	Sender struct {
		Name string `json:"name"`
	} `json:"sender"`
}

func decodePRPayload(raw json.RawMessage) (prPayload, error) {
	var pr prPayload
	if err := json.Unmarshal(raw, &pr); err != nil {
		return pr, fmt.Errorf("decode pull request payload: %w", err)
	}
	return pr, nil
}

func uniqueIssueKeys(texts ...string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, text := range texts {
		for _, match := range issueKeyRe.FindAllString(text, -1) {
			key := strings.ToUpper(match)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	return out
}
