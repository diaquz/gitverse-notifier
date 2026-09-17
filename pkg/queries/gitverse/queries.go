package queries

import (
	"gitverse-notifier/pkg/cache"
	"gitverse-notifier/pkg/integrations/gitverse"
)

type Queries struct {
	client *gitverse.Client
	prs    cache.PullRequestCache
}

func New(client *gitverse.Client, prs cache.PullRequestCache) *Queries {
	return &Queries{client: client, prs: prs}
}
