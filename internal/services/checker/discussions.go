package checker

import (
	"context"
	"log/slog"
	"maps"
	domainNotifications "review_reminder_bot/internal/domain/notifications"
	domainRequests "review_reminder_bot/internal/domain/requests"
)

func (service *MRCheckerService) checkUnresolvedDiscussions(ctx context.Context, request *domainRequests.MergeRequest, notificationChan chan *domainNotifications.Notification) map[string]struct{} {
	// Check unresolved discussions
	discussions, err := service.RepoAdapter.GetMergeRequestDiscussions(ctx, request.ProjectID, request.IID)
	if err != nil {
		slog.Error("cannot get discussions", "error", err)
		return nil
	}

	unresolvedNotesParticipants := make(map[string]struct{})

	for _, discussion := range discussions {
		if len(discussion.Notes) == 0 {
			continue
		}

		lastNote := discussion.Notes[len(discussion.Notes)-1]
		if lastNote.Resolved {
			continue
		}

		if request.IsAssignee(discussion.Notes[0].Author.ID) {
			slog.Info("skip check comment author is assignee", "request", request.WebURL)
			continue
		}

		for _, note := range discussion.Notes {
			unresolvedNotesParticipants[note.Author.Username] = struct{}{}
		}

		if !request.IsAssignee(lastNote.Author.ID) {
			// Assignee should answer to the comment
			for _, assignee := range request.Assignees {
				service.Notify(ctx, notificationChan, &domainNotifications.Notification{
					UserName:  assignee.Username,
					RequestID: request.ID,
					ProjectID: request.ProjectID,
					Type:      domainNotifications.AwaitingThreadResponse,
					WebURL:    request.WebURL,
				})
			}
		} else {
			// Any participant of the discussion should resolve the thread
			usernamesToNotify := make(map[string]struct{})
			for _, note := range discussion.Notes[1:] {
				if request.IsAssignee(note.Author.ID) {
					continue
				}
				usernamesToNotify[note.Author.Username] = struct{}{}
			}
			for username := range maps.Keys(usernamesToNotify) {
				service.Notify(ctx, notificationChan, &domainNotifications.Notification{
					UserName:  username,
					RequestID: request.ID,
					ProjectID: request.ProjectID,
					Type:      domainNotifications.AwaitingThreadResolve,
					WebURL:    request.WebURL,
				})
			}
		}
	}
	return unresolvedNotesParticipants
}
