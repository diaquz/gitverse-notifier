package jira

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gitverse-notifier/pkg/config"

	gojira "github.com/andygrunwald/go-jira"
)

var (
	ErrTaskNotFound = errors.New("jira task not found")
	ErrNotConfigured = errors.New("jira is not configured")
)

type JiraClient struct {
	client *gojira.Client
	baseURL string
}

func NewJiraClient() (*JiraClient, error) {
	url := config.GlobalConfig.JiraURL
	if url == "" {
		return nil, ErrNotConfigured
	}

	httpClient, err := authHTTPClient()
	if err != nil {
		return nil, err
	}

	client, err := gojira.NewClient(httpClient, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create jira client: %w", err)
	}

	return &JiraClient{
		client:  client,
		baseURL: url,
	}, nil
}

func authHTTPClient() (*http.Client, error) {
	cfg := config.GlobalConfig

	switch {
	case cfg.JiraUsername != "" && cfg.JiraPassword != "":
		tp := gojira.BasicAuthTransport{
			Username: cfg.JiraUsername,
			Password: cfg.JiraPassword,
		}
		return tp.Client(), nil
	case cfg.JiraToken != "":
		tp := gojira.PATAuthTransport{Token: cfg.JiraToken}
		return tp.Client(), nil
	default:
		return nil, fmt.Errorf("%w: set JIRA_USERNAME + JIRA_PASSWORD or JIRA_TOKEN", ErrNotConfigured)
	}
}

func (j *JiraClient) FindTask(taskCode string) (*gojira.Issue, error) {
	taskCode = strings.TrimSpace(taskCode)
	if taskCode == "" {
		return nil, fmt.Errorf("empty task code")
	}

	issue, resp, err := j.client.Issue.Get(taskCode, nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w: %s", ErrTaskNotFound, taskCode)
		}

		return nil, fmt.Errorf("failed to find jira issue %s: %w", taskCode, err)
	}

	return issue, nil
}

func (j *JiraClient) IssueURL(taskCode string) string {
	return fmt.Sprintf("%s/browse/%s", j.baseURL, strings.TrimSpace(taskCode))
}

func (j *JiraClient) AddPRComment(taskCode string, pr IssueComment) (*gojira.Comment, error) {
	if _, err := j.FindTask(taskCode); err != nil {
		return nil, err
	}
	return j.AddComment(taskCode, pr.Format())
}

func (j *JiraClient) AddComment(taskCode, body string) (*gojira.Comment, error) {
	comment, resp, err := j.client.Issue.AddComment(taskCode, &gojira.Comment{Body: body})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w: %s", ErrTaskNotFound, taskCode)
		}
		return nil, fmt.Errorf("failed to add comment to issue %s: %w", taskCode, err)
	}

	return comment, nil
}
