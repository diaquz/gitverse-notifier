package groups

import (
	"context"
	"fmt"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/settings"
	"sync"
	"time"
)

type InMemoryEventGroupsManager struct {
	mu              sync.Mutex
	groups          map[string]*eventsGroup
	groupTypes      map[string]GroupType
	strategies      map[string]Strategy
	settingsManager *settings.SettingsManager
}

type eventsGroup struct {
	key       string
	events    []*events.Event
	expiresAt time.Time
	size      int
	strategy  string
}

func NewInMemoryEventGroupsManager(settingsManager *settings.SettingsManager) *InMemoryEventGroupsManager {
	prGroup := PullRequestGroup{}
	useLast := UseLastEvent{}

	return &InMemoryEventGroupsManager{
		settingsManager: settingsManager,
		groups:          make(map[string]*eventsGroup),
		groupTypes: map[string]GroupType{
			prGroup.Name(): prGroup,
		},
		strategies: map[string]Strategy{
			useLast.Name(): useLast,
		},
	}
}

func (m *InMemoryEventGroupsManager) Add(ctx context.Context, event *events.Event) (key string, ok bool) {
	repoSettings := m.settingsManager.SettingsByRepository(event.Repository)
	groupSettings := repoSettings.FindEventGroupSettings(event)
	if groupSettings == nil {
		return
	}

	groupType, ok := m.groupTypes[groupSettings.GroupBy]
	if !ok {
		return
	}

	b := m.addToGroup(groupType.Key(event), groupSettings, event)
	logger.Debug(ctx, "added event to group", "key", b.key, "event", event.Type, "repository", event.Repository, "expires-at", b.expiresAt)

	return b.key, true
}

func (m *InMemoryEventGroupsManager) TryProcessGroup(ctx context.Context, key string) ([]*events.Event, bool) {
	defer m.mu.Unlock()
	m.mu.Lock()

	b, ok := m.groups[key]
	if !ok || len(b.events) < b.size {
		return nil, false
	}

	logger.Debug(ctx, "processing full group", "key", key, "strategy", b.strategy)

	delete(m.groups, key)
	return m.applyStrategy(b), true
}

func (m *InMemoryEventGroupsManager) PopExpiredGroups(ctx context.Context, now time.Time) []*events.Event {
	defer m.mu.Unlock()
	m.mu.Lock()

	out := make([]*events.Event, 0)

	for key, b := range m.groups {
		if b.expiresAt.IsZero() || now.Before(b.expiresAt) {
			continue
		}

		logger.Debug(ctx, "processing expired group", "key", key, "strategy", b.strategy)

		delete(m.groups, key)
		out = append(out, m.applyStrategy(b)...)
	}

	return out
}

func (m *InMemoryEventGroupsManager) PopAllGroups(ctx context.Context) []*events.Event {
	defer m.mu.Unlock()
	m.mu.Lock()

	out := make([]*events.Event, 0, len(m.groups))
	for key, b := range m.groups {
		logger.Debug(ctx, "flushing group on shutdown", "key", key, "strategy", b.strategy, "size", len(b.events))
		out = append(out, m.applyStrategy(b)...)
	}
	m.groups = make(map[string]*eventsGroup)

	return out
}

func (m *InMemoryEventGroupsManager) addToGroup(key string, groupSettings *settings.EventGroupSettings, event *events.Event) *eventsGroup {
	defer m.mu.Unlock()
	m.mu.Lock()

	key = fmt.Sprintf("%s|%s", groupSettings.Key, key)
	b, ok := m.groups[key]
	if !ok {
		b = &eventsGroup{
			key:       key,
			expiresAt: time.Now().Add(groupSettings.TTLTime()),
			size:      groupSettings.Size,
			strategy:  groupSettings.Strategy,
		}
		m.groups[key] = b
	}

	b.events = append(b.events, event)
	return b
}

func (m *InMemoryEventGroupsManager) applyStrategy(b *eventsGroup) []*events.Event {
	if strategy, ok := m.strategies[b.strategy]; ok {
		return strategy.Apply(b.events)
	}

	return b.events
}
