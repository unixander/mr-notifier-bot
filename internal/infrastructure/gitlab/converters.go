package gitlab

import (
	domainRequests "review_reminder_bot/internal/domain/requests"
	"slices"

	gitlabExt "github.com/xanzy/go-gitlab"
)

func FromBasicUserToDomainUser(baseUser *gitlabExt.BasicUser) *domainRequests.User {
	if baseUser == nil {
		return nil
	}
	return &domainRequests.User{
		ID:       baseUser.ID,
		Username: baseUser.Username,
		Name:     baseUser.Name,
		State:    baseUser.State,
	}
}

func FromUsersSliceToDomain(extUsers []*gitlabExt.BasicUser) []*domainRequests.User {
	if len(extUsers) == 0 {
		return []*domainRequests.User{}
	}
	users := make([]*domainRequests.User, 0, len(extUsers))
	for _, user := range extUsers {
		users = append(users, FromBasicUserToDomainUser(user))
	}
	return users
}

func FromPipelineToDomain(pipeline *gitlabExt.Pipeline) *domainRequests.Pipeline {
	if pipeline == nil {
		return nil
	}
	return &domainRequests.Pipeline{
		ID:         pipeline.ID,
		IID:        pipeline.IID,
		ProjectID:  pipeline.ProjectID,
		Status:     domainRequests.PipelineStatus(pipeline.Status),
		UpdatedAt:  pipeline.UpdatedAt,
		CreatedAt:  pipeline.CreatedAt,
		StartedAt:  pipeline.StartedAt,
		FinishedAt: pipeline.FinishedAt,
	}
}

func FromMergeRequestToDomain(request *gitlabExt.MergeRequest) *domainRequests.MergeRequest {
	if request == nil {
		return nil
	}
	return &domainRequests.MergeRequest{
		ID:                          request.ID,
		IID:                         request.IID,
		ProjectID:                   request.ProjectID,
		Title:                       request.Title,
		State:                       request.State,
		CreatedAt:                   request.CreatedAt,
		UpdatedAt:                   request.UpdatedAt,
		Author:                      FromBasicUserToDomainUser(request.Author),
		Assignee:                    FromBasicUserToDomainUser(request.Assignee),
		Assignees:                   FromUsersSliceToDomain(request.Assignees),
		Reviewers:                   FromUsersSliceToDomain(request.Reviewers),
		Draft:                       request.Draft,
		WIP:                         request.WorkInProgress,
		UserNotesCount:              request.UserNotesCount,
		Pipeline:                    FromPipelineToDomain(request.HeadPipeline),
		BlockingDiscussionsResolved: request.BlockingDiscussionsResolved,
		WebURL:                      request.WebURL,
	}
}

func FromDiscussionToDomain(extDiscussion *gitlabExt.Discussion) *domainRequests.Discussion {
	if extDiscussion == nil {
		return nil
	}

	discussion := &domainRequests.Discussion{
		ID:             extDiscussion.ID,
		IndividualNote: extDiscussion.IndividualNote,
		Notes:          make([]*domainRequests.Note, 0, len(extDiscussion.Notes)),
	}

	if len(extDiscussion.Notes) > 0 {
		for _, note := range extDiscussion.Notes {
			if note.System {
				continue
			}
			discussion.Notes = append(discussion.Notes, &domainRequests.Note{
				ID: note.ID,
				Author: domainRequests.User{
					ID:       note.Author.ID,
					Username: note.Author.Username,
					Name:     note.Author.Name,
					State:    note.Author.State,
				},
				Resolvable: note.Resolvable,
				Resolved:   note.Resolved,
				CreatedAt:  note.CreatedAt,
			})
		}

		slices.SortStableFunc(discussion.Notes, func(first, second *domainRequests.Note) int {
			if first.CreatedAt == nil && second.CreatedAt == nil {
				return 0
			}
			if first.CreatedAt == nil && second.CreatedAt != nil {
				return -1
			}
			if first.CreatedAt != nil && second.CreatedAt == nil {
				return 1
			}

			if first.CreatedAt.Before(*second.CreatedAt) {
				return -1
			} else if first.CreatedAt.After(*second.CreatedAt) {
				return 1
			}
			return 0
		})
	}
	return discussion
}
