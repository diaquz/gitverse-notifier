package jira

import (
	"fmt"
	"strings"
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
	author := escapeWiki(pr.Author)
	if author == "" {
		author = escapeWiki(pr.AuthorLogin)
	}
	title := escapeWiki(pr.Title)
	repo := escapeWiki(pr.Repository)

	var b strings.Builder
	fmt.Fprintf(&b, "*%s* mentioned this issue in pull request *%s*", author, title)
	if repo != "" {
		fmt.Fprintf(&b, " (%s)", repo)
	}
	if pr.URL != "" {
		fmt.Fprintf(&b, "\n%s", pr.URL)
	}
	if pr.Branch != "" {
		fmt.Fprintf(&b, "\nBranch: *%s*", escapeWiki(pr.Branch))
	}
	return b.String()
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
