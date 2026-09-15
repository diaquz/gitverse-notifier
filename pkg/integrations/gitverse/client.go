package gitverse

import (
	"context"
	"fmt"
	"time"

	"gitverse-notifier/pkg/config"

	"github.com/go-resty/resty/v2"
)

const versionHeader = "application/vnd.gitverse.object+json;version=1"

type Client struct {
	resty *resty.Client
}

func NewClient() (*Client, error) {
	cfg := config.GlobalConfig
	if cfg.GitverseAPIURL == "" || cfg.GitverseToken == "" {
		return nil, fmt.Errorf("GITVERSE_API_URL and GITVERSE_TOKEN configs required")
	}

	return &Client{resty: newRestyClient(cfg.GitverseAPIURL, cfg.GitverseToken)}, nil
}

func newRestyClient(baseURL, token string) *resty.Client {
	return resty.New().
		SetBaseURL(baseURL).
		SetAuthToken(token).
		SetHeader("Accept", versionHeader).
		SetTimeout(30 * time.Second)
}

func (c *Client) GetPullRequest(ctx context.Context, repository string, number int) (pr *PullRequest, err error) {
	if repository == "" || number <= 0 {
		return nil, fmt.Errorf("invalid pull request parameters: %s/pulls/%d", repository, number)
	}

	resp, err := c.resty.R().
		SetContext(ctx).
		SetPathParams(map[string]string{
			"repo":   repository,
			"number": fmt.Sprintf("%d", number),
		}).
		SetResult(pr).
		Get("/repos/{repo}/pulls/{number}")

	if err != nil {
		return nil, fmt.Errorf("failed to make gitverse request: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("gitverse api returned %d for %s: %s",
			resp.StatusCode(), resp.Request.URL, string(resp.Body()))
	}

	return
}
