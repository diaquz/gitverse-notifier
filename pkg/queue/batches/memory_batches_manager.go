package batches

import (
	"context"
	"fmt"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/settings"
	"sync"
	"time"
)

type InMemoryBatchesManager struct {
	mu              sync.Mutex
	buckets         map[string]*bucket
	groupTypes      map[string]GroupType
	strategies      map[string]Strategy
	settingsManager *settings.SettingsManager
}

type bucket struct {
	key       string
	events    []*events.Event
	expiresAt time.Time
	size      int
	strategy  string
}

func NewInMemoryBatchesManager(settingsManager *settings.SettingsManager) *InMemoryBatchesManager {
	prGroup := PullRequestBatch{}
	useLast := UseLastEvent{}

	return &InMemoryBatchesManager{
		settingsManager: settingsManager,
		buckets:         make(map[string]*bucket),
		groupTypes: map[string]GroupType{
			prGroup.Name(): prGroup,
		},
		strategies: map[string]Strategy{
			useLast.Name(): useLast,
		},
	}
}

func (m *InMemoryBatchesManager) Add(ctx context.Context, event *events.Event) (key string, ok bool) {
	repoSettings := m.settingsManager.SettingsByRepository(event.Repository)
	batchSettings := repoSettings.FindBatchSettings(event)
	if batchSettings == nil {
		return
	}

	groupType, ok := m.groupTypes[batchSettings.GroupBy]
	if !ok {
		return
	}

	b := m.addToBucket(groupType.Key(event), batchSettings, event)
	logger.Debug(ctx, "added event to batch", "key", b.key, "event", event.Type, "repository", event.Repository, "expires-at", b.expiresAt)

	return b.key, true
}

func (m *InMemoryBatchesManager) TryProcessBucket(ctx context.Context, key string) ([]*events.Event, bool) {
	defer m.mu.Unlock()
	m.mu.Lock()

	b, ok := m.buckets[key]
	if !ok || len(b.events) < b.size {
		return nil, false
	}

	logger.Debug(ctx, "processing full batch", "key", key, "strategy", b.strategy)

	delete(m.buckets, key)
	return m.applyStrategy(b), true
}

func (m *InMemoryBatchesManager) PopExpired(ctx context.Context, now time.Time) []*events.Event {
	defer m.mu.Unlock()
	m.mu.Lock()

	out := make([]*events.Event, 0)

	for key, b := range m.buckets {
		if b.expiresAt.IsZero() || now.Before(b.expiresAt) {
			continue
		}

		logger.Debug(ctx, "processing expired batch", "key", key, "strategy", b.strategy)

		delete(m.buckets, key)
		out = append(out, m.applyStrategy(b)...)
	}

	return out
}

func (m *InMemoryBatchesManager) PopAll(ctx context.Context) []*events.Event {
	defer m.mu.Unlock()
	m.mu.Lock()

	out := make([]*events.Event, 0, len(m.buckets))
	for key, b := range m.buckets {
		logger.Debug(ctx, "flushing batch on shutdown", "key", key, "strategy", b.strategy, "size", len(b.events))
		out = append(out, m.applyStrategy(b)...)
	}
	m.buckets = make(map[string]*bucket)

	return out
}

func (m *InMemoryBatchesManager) addToBucket(key string, batchSettings *settings.BatchSetting, event *events.Event) *bucket {
	defer m.mu.Unlock()
	m.mu.Lock()

	key = fmt.Sprintf("%s|%s", batchSettings.Key, key)
	b, ok := m.buckets[key]
	if !ok {
		b = &bucket{
			key:       key,
			expiresAt: time.Now().Add(batchSettings.TTLTime()),
			size:      batchSettings.Size,
			strategy:  batchSettings.Strategy,
		}
		m.buckets[key] = b
	}

	b.events = append(b.events, event)
	return b
}

func (m *InMemoryBatchesManager) applyStrategy(b *bucket) []*events.Event {
	if strategy, ok := m.strategies[b.strategy]; ok {
		return strategy.Apply(b.events)
	}

	return b.events
}
