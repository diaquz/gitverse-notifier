package handlers

import (
	"context"
	"fmt"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/telegram"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/settings"
	"gitverse-notifier/pkg/templates"
)

const defaultTelegramTemplate = "telegram/default"

type TelegramNotify struct {
	client    *telegram.Client
	templates *templates.Engine
}

func NewTelegramNotify(client *telegram.Client, engine *templates.Engine) *TelegramNotify {
	return &TelegramNotify{client: client, templates: engine}
}

func (a *TelegramNotify) Name() string {
	return "telegram.notify"
}

func (a *TelegramNotify) Ready() bool {
	return a.client != nil
}

func (a *TelegramNotify) Run(ctx context.Context, event *events.Event, rule *settings.ActionRule) error {
	name := strings.TrimSpace(rule.Template)
	if name == "" {
		name = defaultTelegramTemplate
	}

	body, err := a.templates.Render(name, templates.DataFromEvent(event))
	if err != nil {
		return fmt.Errorf("failed to redener template %s: %w", name, err)
	}

	logger.Debug(ctx, "rendered telegram message body", "action", a.Name(), "body", body)

	if err := a.client.SendMessage(ctx, body); err != nil {
		return err 
	}

	logger.Info(ctx, "successfully sent telegram notification",
		"action", a.Name(),
		"event", event.Type,
		"template", name,
		"repository", event.Repository,
	)
	return nil
}
