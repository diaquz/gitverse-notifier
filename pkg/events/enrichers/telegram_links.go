package enrichers

import (
	"context"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/settings"
)

type TelegramLinks struct {
	manager *settings.SettingsManager
}

func NewTelegramLinks(manager *settings.SettingsManager) *TelegramLinks {
	return &TelegramLinks{manager: manager}
}

func (e *TelegramLinks) Name() string {
	return "enricher.telegram-links"
}

func (e *TelegramLinks) Skip(event *events.Event) bool {
	return event == nil
}

func (e *TelegramLinks) Enrich(_ context.Context, event *events.Event) error {
	settings := e.manager.SettingsByRepository(event.Repository)

	e.addTgTag(&event.Sender, settings)
	for i := range len(event.PullRequest.Reviewers) {
		e.addTgTag(&event.PullRequest.Reviewers[i], settings)
	}

	return nil
}

func (e *TelegramLinks) addTgTag(actor *events.Actor, settings *settings.RepositorySettings) {
	tag, ok := settings.TelegramTags[actor.Name]
	if ok {
		actor.TgTag = tag
	}
}
