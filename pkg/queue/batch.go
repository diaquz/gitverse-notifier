package queue

import (
	"fmt"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/queue/batches"
	"gitverse-notifier/pkg/settings"
	"sync"
	"time"
)

type BatchGroupType string
type BatchStrategyType string

const (
	GroupByPR    BatchGroupType    = "batch.pull_request"
	UseLastEvent BatchStrategyType = "batch.use_last_event"
)

type BatchesManager interface {
	Add(event *events.Event) (string, bool)
	TryProcessBucket(key string) ([]*events.Event, bool)
	ProcessBucket(key string) ([]*events.Event, bool)
}

type InMemoryBatchesManager struct {
	mu              sync.Mutex
	buckets         map[string]*bucket
	groupTypes      map[string]batches.GroupType
	strategies      map[string]batches.Strategy
	settingsManager *settings.SettingsManager
}

type bucket struct {
	key string
	events   []*events.Event
	timer    *time.Timer
	flushing bool
	timeout  int
	size     int
	strategy string
}

func NewInMemoryBatchesManager(settingsManager *settings.SettingsManager) *InMemoryBatchesManager {
	manager := &InMemoryBatchesManager{
		settingsManager: settingsManager,
		buckets:         make(map[string]*bucket, 0),
		groupTypes: map[string]batches.GroupType{
			string(GroupByPR): batches.PullRequestBatch{},
		},
		strategies: map[string]batches.Strategy{
			string(UseLastEvent): batches.UseLastEvent{},
		},
	}

	return manager
}

func (m *InMemoryBatchesManager) Add(event *events.Event) (key string, ok bool) {
	settings := m.settingsManager.SettingsByRepository(event.Repository)
	batchSettings := settings.FindBatchSettings(event)
	if batchSettings == nil {
		return
	}

	groupType, ok := m.groupTypes[batchSettings.GroupBy]
	if !ok {
		return
	}

	bucket := m.addToBucket(groupType.Key(event), batchSettings, event)
	return bucket.key, true
}

func (m *InMemoryBatchesManager) TryProcessBucket(key string) ([]*events.Event, bool) {
	if _, ok := m.buckets[key]; ok {
		return m.ProcessBucket(key)
	}

	return nil, false
}

func (m *InMemoryBatchesManager) ProcessBucket(key string) ([]*events.Event, bool) {
	bucket := m.popBucket(key)
	if bucket == nil {
		return nil, false
	}

	if strategy, ok := m.strategies[bucket.strategy]; ok {
		return strategy.Apply(bucket.events), true
	}

	return nil, false
}

func (m *InMemoryBatchesManager) addToBucket(key string, batchSettings *settings.BatchSetting, event *events.Event) *bucket {
	defer m.mu.Unlock()
	m.mu.Lock()

	key = fmt.Sprintf("%s|%s", batchSettings.Key, key)
	b, ok := m.buckets[key]
	if !ok {
		b = &bucket{
			key: key,
			timeout:  batchSettings.Timeout,
			size:     batchSettings.Size,
			strategy: batchSettings.Strategy,
		}
		m.buckets[key] = b
	}

	b.events = append(b.events, event)

	return b
}

func (m *InMemoryBatchesManager) popBucket(key string) *bucket {
	defer m.mu.Unlock()
	m.mu.Lock()

	bucket, ok := m.buckets[key]
	if !ok {
		return nil
	}

	delete(m.buckets, key)
	return bucket
}
