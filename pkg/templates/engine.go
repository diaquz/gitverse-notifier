package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gitverse-notifier/pkg/config"
	"gitverse-notifier/pkg/events"
	"gitverse-notifier/pkg/logger"
)

type Data struct {
	Type          string
	Action        string
	Repository    string
	RepositoryURL string
	Ref           string
	Branch        string
	BranchURL     string
	Sender        events.Actor
	PullRequest   events.PullRequestInfo
	Comment       events.CommentInfo
	Push          events.PushInfo
	Status        events.StatusInfo
	IssueKeys     []string
}

func DataFromEvent(ev events.Event) Data {
	return Data{
		Type:          string(ev.Type),
		Action:        ev.Action,
		Repository:    ev.Repository,
		RepositoryURL: ev.RepositoryURL,
		Ref:           ev.Ref,
		Branch:        ev.Branch,
		BranchURL:     ev.BranchURL,
		Sender:        ev.Sender,
		PullRequest:   ev.PullRequest,
		Comment:       ev.Comment,
		Push:          ev.Push,
		Status:        ev.Status,
		IssueKeys:     ev.IssueKeys,
	}
}

type Engine struct {
	engine *template.Template
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"jiraEscape":    JiraEscape,
		"tgEscape":      TelegramEscape,
		"trim":          strings.TrimSpace,
		"join":          strings.Join,
		"jiraLink":      JiraLink,
		"tgLink":        TelegramLink,
		"issueURL":      IssueURL,
		"tgIssueURLs":   TelegramIssueURLs,
		"jiraIssueURLs": JiraIssueURLs,
	}
}

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
	s = normalizeLiteralNewlines(s)

	switch config.GlobalConfig.TelegramParseMode {
	case "HTML":
		return strings.NewReplacer(
			"&", "&amp;",
			"<", "&lt;",
			">", "&gt;",
		).Replace(s)
	case "MarkdownV2":
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
	default: // MarkdownV1
		return strings.NewReplacer(
			"_", "\\_",
			"*", "\\*",
			"`", "\\`",
			"[", "\\[",
		).Replace(s)
	}
}

func normalizeLiteralNewlines(s string) string {
	s = strings.ReplaceAll(s, "\\\\n", "\n")
	s = strings.ReplaceAll(s, "\\n", "\n")
	return s
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

func SetupTemplateEngine() (*Engine, error) {
	cfg := config.GlobalConfig
	pattern := filepath.Join(cfg.TemplatesDirPath, cfg.TemplatesPattern)

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob templates by %s: %w", pattern, err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no templates found by %s", pattern)
	}

	root := template.New("").Funcs(funcMap())
	for _, path := range matches {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read template %s: %w", path, err)
		}

		// Учитываем, что шаблоны могут быть в каталогах и иметь одинаковые имена, например
		// 	telegram/default.tmpl
		// 	jira/default.tml
		rel, err := filepath.Rel(cfg.TemplatesDirPath, path)
		name := strings.TrimSuffix(rel, filepath.Ext(rel))
		if _, err := root.New(name).Parse(string(content)); err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", path, err)
		}

		logger.Debug(nil, "template registered", "action", "templates_setup", "template", name, "path", path)
	}

	return &Engine{engine: root}, nil
}

func (e *Engine) Render(name string, data Data) (string, error) {
	tmpl := e.engine.Lookup(name)
	if tmpl == nil {
		return "", fmt.Errorf("template %q not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %q: %w", name, err)
	}

	return buf.String(), nil
}

func (e *Engine) Has(name string) bool {
	return e.engine.Lookup(name) != nil
}
