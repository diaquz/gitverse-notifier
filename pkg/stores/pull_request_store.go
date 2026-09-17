package stores

import (
	"fmt"
	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/integrations/gitverse"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

const cacheCleanupPeriod = 30 * time.Minute

type PullRequestStore interface {
	// Во многих событиях есть только заголовок PR, поэтому дополнительно кэширует номер
	GetByTitle(repo, title string, number int) (*gitverse.PullRequest, bool)
	Set(repo, title string, number int, pr *gitverse.PullRequest)
}

type MemoryPullRequestStore struct {
	cache *gocache.Cache
}

func NewMemoryPullRequestStore() *MemoryPullRequestStore {
	ttl := time.Duration(config.GlobalConfig.CacheTTL)
	return &MemoryPullRequestStore{
		cache: gocache.New(ttl*time.Minute, cacheCleanupPeriod),
	}
}

func (s *MemoryPullRequestStore) GetByTitle(repo, title string, number int) (pr *gitverse.PullRequest, ok bool) {
	if number <= 0 {
		number = s.getPRNumber(repo, title)
	}

	value, ok := s.cache.Get(prCacheKey(repo, number))
	if !ok {
		return nil, false
	}

	pr, ok = value.(*gitverse.PullRequest)
	return
}

func (s *MemoryPullRequestStore) Set(repo, title string, number int, pr *gitverse.PullRequest) {
	if pr == nil || number <= 0 {
		return
	}
	s.cache.Set(prCacheKey(repo, number), pr, gocache.DefaultExpiration)
	s.cache.Set(prNumberCacheKey(repo, title), number, gocache.DefaultExpiration)
}

func (s *MemoryPullRequestStore) getPRNumber(repo, title string) int {
	value, ok := s.cache.Get(prNumberCacheKey(repo, title))
	if !ok {
		return 0
	}

	if number, ok := value.(int); ok {
		return number
	}

	return 0
}

func prCacheKey(repo string, number int) string {
	return fmt.Sprintf("%s#%d", repo, number)
}

func prNumberCacheKey(repo, title string) string {
	return fmt.Sprintf("%s|%s", repo, title)
}
