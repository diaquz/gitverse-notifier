package gitverse

import "time"

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Login    string `json:"login"`
	Type     string `json:"type"`
	Bio      string `json:"bio"`
	Email    string `json:"email"`
	URL      string `json:"url"`
	Location string `json:"location"`
}

type RepoOwner struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

type Repository struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Owner         RepoOwner `json:"owner"`
	Private       bool      `json:"private"`
	Description   string    `json:"description"`
	DefaultBranch string    `json:"default_branch"`
	URL           string    `json:"url"`
	SSHURL        string    `json:"ssh_url"`
}

type PRBranch struct {
	Label  string     `json:"label"`
	Ref    string     `json:"ref"`
	SHA    string     `json:"sha"`
	RepoID int64      `json:"repo_id"`
	Repo   Repository `json:"repo"`
}

type PullRequest struct {
	ID                  int64      `json:"id"`
	URL                 string     `json:"url"`
	Number              int        `json:"number"`
	User                User       `json:"user"`
	Title               string     `json:"title"`
	Body                string     `json:"body"`
	RequestedReviewers  []User     `json:"requested_reviewers"`
	State               string     `json:"state"`
	Locked              bool       `json:"locked"`
	IsDraft             bool       `json:"is_draft"`
	Comments            int        `json:"comments"`
	Mergeable           bool       `json:"mergeable"`
	Merged              bool       `json:"merged"`
	MergedAt            *time.Time `json:"merged_at"`
	MergeCommitSHA      string     `json:"merge_commit_sha"`
	MergedBy            *User      `json:"merged_by"`
	MaintainerCanModify bool       `json:"maintainer_can_modify"`
	Base                PRBranch   `json:"base"`
	Head                PRBranch   `json:"head"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	ClosedAt            *time.Time `json:"closed_at"`
}
