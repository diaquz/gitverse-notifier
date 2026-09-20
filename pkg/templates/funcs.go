package templates

import (
	"fmt"
	"strings"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events"
)

func JiraEscape(s string) string {
	replacer := strings.NewReplacer(
		"{", "\\{",
		"}", "\\}",
		"[", "\\[",
		"]", "\\]",
		"|", "\\|",
	)
	return replacer.Replace(s)
}

func TelegramEscape(s string) string {
	switch config.GlobalConfig.TelegramParseMode {
	case "HTML":
		return strings.NewReplacer(
			"&", "&amp;",
			"<", "&lt;",
			">", "&gt;",
		).Replace(s)
	case "Markdown":
		return strings.NewReplacer(
			"_", "\\_",
			"*", "\\*",
			"`", "\\`",
			"[", "\\[",
		).Replace(s)
	default: // MarkdownV2
		return escapeMarkdownV2(s)
	}
}

func escapeMarkdownV2(s string) string {
	return strings.NewReplacer(
		"\\", "\\\\",
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	).Replace(s)
}

func TelegramIssueURLs(keys []string) string {
	return joinIssueURLs(keys, TelegramLink)
}

func JiraIssueURLs(keys []string) string {
	return joinIssueURLs(keys, JiraLink)
}

func joinIssueURLs(keys []string, linkFn func(text, link string) string) string {
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		parts = append(parts, linkFn(key, IssueURL(key)))
	}
	return strings.Join(parts, ", ")
}

func IssueURL(key string) string {
	base := strings.TrimRight(config.GlobalConfig.JiraURL, "/")
	if key == "" || base == "" {
		return ""
	}
	return base + "/browse/" + key
}

func JiraLink(text, link string) string {
	text = JiraEscape(text)
	if link == "" {
		return text
	}
	return fmt.Sprintf("[%s|%s]", text, link)
}

func TelegramLink(text, link string) string {
	text = TelegramEscape(text)
	if link == "" {
		return text
	}

	return fmt.Sprintf("[%s](%s)", text, link)
}

func TelegramMention(actor events.Actor) string {
	if tag := strings.TrimPrefix(actor.TgTag, "@"); tag != "" {
		return "@" + TelegramEscape(tag)
	}

	return TelegramLink(actor.Name, actor.URL)
}

func TelegramMentions(actors []events.Actor) string {
	parts := make([]string, 0, len(actors))
	for _, actor := range actors {
		if m := TelegramMention(actor); m != "" {
			parts = append(parts, m)
		}
	}
	return strings.Join(parts, ", ")
}
