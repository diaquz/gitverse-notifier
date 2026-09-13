package actions

import (
	"context"
	"fmt"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/integrations/telegram"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/templates"
)

const defaultTelegramTemplate = "telegram/default"

type TelegramNotify struct {
	Client    *telegram.Client
	Templates *templates.Engine
}

func NewTelegramNotify(client *telegram.Client, engine *templates.Engine) *TelegramNotify {
	return &TelegramNotify{Client: client, Templates: engine}
}

func (a *TelegramNotify) Name() string {
	return "telegram.notify"
}

func (a *TelegramNotify) Run(ctx context.Context, ev events.Event, rule events.ActionRule) error {
	name := strings.TrimSpace(rule.Template)
	if name == "" {
		name = defaultTelegramTemplate
	}

	body, err := a.Templates.Render(name, templates.DataFromEvent(ev))
	if err != nil {
		return err
	}
	logger.Debugf("[Action=%s] rendered telegram message body:\n%s", a.Name(), body)

	if err := a.Client.SendMessage(ctx, body); err != nil {
		return fmt.Errorf("failed to send telegram notification: %w", err)
	}

	logger.Infof("[Action=%s] successfully sent telegram notification (repository=%s)", a.Name(), ev.Repository)
	return nil
}
