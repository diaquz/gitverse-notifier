package jira

import (
	"fmt"
	"strings"
)

type IssueComment struct {
	Repository string
	Title      string
	URL        string
	Author     string
	AuthorLogin string
	Branch     string
	BranchURL string
	Action     string
}

func (pr *IssueComment) Format() string {
	text := fmt.Sprintf(
		"*%s*[%s] %s mentioned this issue in *pull request*[%s] on branch *%s*[%s]",
		pr.Author, pr.AuthorLogin, pr.URL, pr.Branch, pr.BranchURL,
	)

	return text
}

func escapeWiki(s string) string {
	replacer := strings.NewReplacer(
		"{", "\\{",
		"}", "\\}",
		"[", "\\[",
		"]", "\\]",
		"|", "\\|",
	)
	return replacer.Replace(s)
}
