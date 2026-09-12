package actions

import (
	"context"
	"strings"

	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
	"gitverse-notifier/pkg/templates"
)

const defaultTelegramTemplate = "telegram/default.tmpl"

type TelegramNotify struct {
	Templates *templates.Engine
}

func NewTelegramNotify(engine *templates.Engine) *TelegramNotify {
	return &TelegramNotify{Templates: engine}
}

func (a *TelegramNotify) Name() string {
	return "telegram.notify"
}

func (a *TelegramNotify) Run(_ context.Context, ev events.Event, rule events.ActionRule) error {
	name := strings.TrimSpace(rule.Template)
	if name == "" {
		name = defaultTelegramTemplate
	}

	body, err := a.Templates.Render(name, templates.DataFromEvent(ev))
	if err != nil {
		return err
	}
	logger.Debugf("[Action=%s] rendered telegram message body: %s", a.Name(), body)
	return nil
}
