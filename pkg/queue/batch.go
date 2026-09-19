package queue

import (
	"fmt"
	"sync"
	"time"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/queue/batches"
	"gitverse-notifier/pkg/settings"
)

type BatchesManager interface {
	Add(event *events.Event) (string, bool)
	TryProcessBucket(key string) ([]*events.Event, bool)
	ProcessBucket(key string) ([]*events.Event, bool)
	PopExpired(now time.Time) []*events.Event
}

type InMemoryBatchesManager struct {
	mu              sync.Mutex
	buckets         map[string]*bucket
	groupTypes      map[string]batches.GroupType
	strategies      map[string]batches.Strategy
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
	prGroup := batches.PullRequestBatch{}
	useLast := batches.UseLastEvent{}

	return &InMemoryBatchesManager{
		settingsManager: settingsManager,
		buckets:         make(map[string]*bucket),
		groupTypes: map[string]batches.GroupType{
			prGroup.Name(): prGroup,
		},
		strategies: map[string]batches.Strategy{
			useLast.Name(): useLast,
		},
	}
}

func (m *InMemoryBatchesManager) Add(event *events.Event) (key string, ok bool) {
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
	return b.key, true
}

func (m *InMemoryBatchesManager) TryProcessBucket(key string) ([]*events.Event, bool) {
	defer m.mu.Unlock()
	m.mu.Lock()

	b, ok := m.buckets[key]
	if !ok || len(b.events) < b.size {
		return nil, false
	}

	delete(m.buckets, key)
	return m.applyStrategy(b), true
}

func (m *InMemoryBatchesManager) ProcessBucket(key string) ([]*events.Event, bool) {
	if b, ok := m.buckets[key]; ok {
		return m.applyStrategy(b), true
	}

	return nil, false
}

func (m *InMemoryBatchesManager) PopExpired(now time.Time) []*events.Event {
	defer m.mu.Unlock()
	m.mu.Lock()

	out := make([]*events.Event, 0)

	for key, b := range m.buckets {
		if b.expiresAt.IsZero() || now.Before(b.expiresAt) {
			continue
		}

		out = append(out, m.applyStrategy(b)...)
		delete(m.buckets, key)
	}

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
