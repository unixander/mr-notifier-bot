package requests

import "time"

type MergeRequest struct {
	ID                          int
	IID                         int
	ProjectID                   int
	Title                       string
	State                       string
	CreatedAt                   *time.Time
	UpdatedAt                   *time.Time
	Author                      *User
	Assignee                    *User
	Assignees                   []*User
	Reviewers                   []*User
	Draft                       bool
	WIP                         bool
	UserNotesCount              int
	Pipeline                    *Pipeline
	BlockingDiscussionsResolved bool
	WebURL                      string
}

func (request *MergeRequest) IsAssignee(userID int) bool {
	for _, assignee := range request.Assignees {
		if assignee.ID == userID {
			return true
		}
	}
	return false
}
