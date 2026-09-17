package telegram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitverse-notifier/pkg/config"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Client struct {
	bot       *bot.Bot
	chatID    string
	threadID  int
	parseMode models.ParseMode
}

func NewTelegramClient() (*Client, error) {
	cfg := config.GlobalConfig
	token := strings.TrimSpace(cfg.TelegramBotToken)
	chatID := strings.TrimSpace(cfg.TelegramChatID)

	if token == "" || chatID == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID configs required")
	}

	opts, err := botOptions(cfg.TelegramProxyURL)
	if err != nil {
		return nil, err
	}

	tgBot, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to setup telegram integration: %w", err)
	}

	return &Client{
		bot:      tgBot,
		chatID:   chatID,
		threadID: cfg.TelegramThreadId,
		// TODO: Неправильный тип вернет ошибки при отправке, но я чет не хочу писать отдельную валидацию, будет на совести пользователя
		parseMode: models.ParseMode(cfg.TelegramParseMode),
	}, nil
}

func botOptions(proxyURL string) ([]bot.Option, error) {
	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL == "" {
		return nil, nil
	}

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid TELEGRAM_PROXY_URL %s: %w", proxyURL, err)
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsed),
		},
	}

	return []bot.Option{
		bot.WithHTTPClient(10*time.Second, httpClient),
	}, nil
}

func (c *Client) SendMessage(ctx context.Context, text string) error {
	if _, err := c.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          c.chatID,
		MessageThreadID: c.threadID,
		Text:            text,
		ParseMode:       c.parseMode,
	}); err != nil {
		return fmt.Errorf("failed to send telegram message to %s: %w", c.chatID, err)
	}

	return nil
}
