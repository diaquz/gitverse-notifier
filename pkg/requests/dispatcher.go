package requests

import (
	"gitverse-notifier/pkg/cache"
	"gitverse-notifier/pkg/integrations/gitverse"
)

type RequestsDispatcher struct {
	client *gitverse.Client
	prs    cache.PullRequestCache
}

func NewRequestsDispather(client *gitverse.Client, prs cache.PullRequestCache) *RequestsDispatcher {
	return &RequestsDispatcher{client: client, prs: prs}
}
