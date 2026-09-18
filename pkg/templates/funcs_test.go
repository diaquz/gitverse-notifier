package templates

import (
	"testing"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events"

	"github.com/stretchr/testify/require"
)

func TestLastReviewerLinks(t *testing.T) {
	config.GlobalConfig = &config.Config{TelegramParseMode: "MarkdownV2"}

	reviewers := []events.Actor{
		{Name: "first", URL: "https://git.example/u/first"},
		{Name: "last", URL: "https://git.example/u/last"},
	}

	require.Equal(t, "[last](https://git.example/u/last)", TelegramLastReviewer(reviewers))
	require.Equal(t, "[last|https://git.example/u/last]", JiraLastReviewer(reviewers))
	require.Equal(t, "", TelegramLastReviewer(nil))
	require.Equal(t, "", JiraLastReviewer(nil))
}

func TestTelegramMention(t *testing.T) {
	config.GlobalConfig = &config.Config{TelegramParseMode: "MarkdownV2"}

	require.Equal(t, "@test", TelegramMention(events.Actor{TgTag: "test"}))
	require.Equal(t, "@test", TelegramMention(events.Actor{TgTag: "@test"}))
	require.Equal(t, "@user\\_name", TelegramMention(events.Actor{TgTag: "user_name"}))
	require.Equal(t, "[alice](https://git.example/u/alice)", TelegramMention(events.Actor{
		Name: "alice",
		URL:  "https://git.example/u/alice",
	}))
	require.Equal(t, "", TelegramMention(events.Actor{}))

	reviewers := []events.Actor{
		{Name: "first", URL: "https://git.example/u/first"},
		{Name: "last", URL: "https://git.example/u/last", TgTag: "last_tg"},
	}
	require.Equal(t, "@last\\_tg", TelegramLastReviewer(reviewers))
}

func TestTelegramMentions(t *testing.T) {
	config.GlobalConfig = &config.Config{TelegramParseMode: "MarkdownV2"}

	actors := []events.Actor{
		{TgTag: "one"},
		{Name: "two", URL: "https://git.example/u/two"},
		{TgTag: "three_tag"},
	}
	require.Equal(t, "@one, [two](https://git.example/u/two), @three\\_tag", TelegramMentions(actors))
	require.Equal(t, "", TelegramMentions(nil))
}
