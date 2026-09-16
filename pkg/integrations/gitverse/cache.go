package gitverse

import (
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

const defaultPRTTL = 15 * time.Minute

type PullRequestStore interface {
	Get(repo string, number int) (*PullRequest, bool)
	Set(repo string, number int, pr *PullRequest)
}

type MemoryPullRequestStore struct {
	cache *gocache.Cache
}

func NewMemoryPullRequestStore() *MemoryPullRequestStore {
	return &MemoryPullRequestStore{
		cache: gocache.New(defaultPRTTL, 30*time.Minute),
	}
}

func (s *MemoryPullRequestStore) Get(repo string, number int) (pr *PullRequest, ok bool) {
	value, ok := s.cache.Get(prCacheKey(repo, number))
	if !ok {
		return nil, false
	}

	pr, ok = value.(*PullRequest)
	return
}

func (s *MemoryPullRequestStore) Set(repo string, number int, pr *PullRequest) {
	if pr == nil {
		return
	}
	s.cache.Set(prCacheKey(repo, number), pr, gocache.DefaultExpiration)
}

func prCacheKey(repo string, number int) string {
	return fmt.Sprintf("%s#%d", repo, number)
}
