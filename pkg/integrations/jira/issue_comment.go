package jira

import (
	"fmt"
	"strings"

	"gitverse-notifier/pkg/templates"
)

type IssueComment struct {
	Repository  string
	Title       string
	URL         string
	Author      string
	AuthorLogin string
	Branch      string
	BranchURL   string
	Action      string
}

func (pr *IssueComment) Format() string {
	author := templates.JiraEscape(pr.Author)
	if author == "" {
		author = templates.JiraEscape(pr.AuthorLogin)
	}
	title := templates.JiraEscape(pr.Title)
	repo := templates.JiraEscape(pr.Repository)

	var b strings.Builder
	fmt.Fprintf(&b, "*%s* mentioned this issue in pull request *%s*", author, title)
	if repo != "" {
		fmt.Fprintf(&b, " (%s)", repo)
	}
	if pr.URL != "" {
		fmt.Fprintf(&b, "\n%s", pr.URL)
	}
	if pr.Branch != "" {
		fmt.Fprintf(&b, "\nBranch: *%s*", templates.JiraEscape(pr.Branch))
	}
	return b.String()
}
