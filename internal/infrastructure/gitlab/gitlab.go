package gitlab

import (
	"context"
	"log/slog"
	domainRequests "review_reminder_bot/internal/domain/requests"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	gitlabExt "github.com/xanzy/go-gitlab"
	"golang.org/x/time/rate"
)

type GitLabAdapter struct {
	client *gitlabExt.Client
}

func New(host, token string, requestsPerSecond int) (*GitLabAdapter, error) {
	git, err := gitlabExt.NewClient(
		token,
		gitlabExt.WithBaseURL(host),
		gitlabExt.WithCustomLimiter(rate.NewLimiter(rate.Every(time.Second), requestsPerSecond)),
	)
	if err != nil {
		return nil, err
	}
	return &GitLabAdapter{
		client: git,
	}, nil
}

func (adapter *GitLabAdapter) GetMergeRequests(ctx context.Context, groupID string, filterInterval *time.Duration) <-chan *domainRequests.MergeRequest {
	options := &gitlabExt.ListGroupMergeRequestsOptions{
		State:   gitlabExt.Ptr("opened"),
		OrderBy: gitlabExt.Ptr("created_at"),
		Draft:   gitlabExt.Ptr(false),
		WIP:     gitlabExt.Ptr("no"),
		ListOptions: gitlabExt.ListOptions{
			Pagination: "offset",
			PerPage:    100,
		},
	}
	if filterInterval != nil {
		options.UpdatedAfter = gitlabExt.Ptr(time.Now().Add(-*filterInterval))
	}

	requestsChan := make(chan *domainRequests.MergeRequest)

	go func() {
		defer close(requestsChan)

		totalPages := 1
		nextPage := 1
		for totalPages >= nextPage {
			options.ListOptions.Page = nextPage
			rawRequests, response, err := adapter.client.MergeRequests.ListGroupMergeRequests(
				groupID,
				options,
				func(r *retryablehttp.Request) error {
					*r = *r.WithContext(ctx)
					return nil
				},
			)
			if err != nil {
				slog.Error("cannot get merge requests", "error", err)
				return
			}
			slog.Info("got merge requests", "count", len(rawRequests), "total", response.TotalItems)
			for _, request := range rawRequests {
				select {
				case <-ctx.Done():
					return
				case requestsChan <- FromMergeRequestToDomain(request):
				}
			}
			nextPage = response.NextPage
			totalPages = response.TotalPages
			if nextPage == 0 {
				return
			}
		}
	}()

	return requestsChan
}

func (adapter *GitLabAdapter) GetMergeRequestApprovals(ctx context.Context, projectID, requestID int) ([]int, error) {
	approvals, _, err := adapter.client.MergeRequests.GetMergeRequestApprovals(
		projectID,
		requestID,
		func(r *retryablehttp.Request) error {
			*r = *r.WithContext(ctx)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	approvedByIDs := make([]int, 0, len(approvals.ApprovedBy))
	for _, approvedBy := range approvals.ApprovedBy {
		approvedByIDs = append(approvedByIDs, approvedBy.User.ID)
	}
	return approvedByIDs, nil
}

func (adapter *GitLabAdapter) GetMergeRequestParticipants(ctx context.Context, projectID, requestID int) ([]*domainRequests.User, error) {
	participants, _, err := adapter.client.MergeRequests.GetMergeRequestParticipants(
		projectID,
		requestID,
		func(r *retryablehttp.Request) error {
			*r = *r.WithContext(ctx)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return FromUsersSliceToDomain(participants), nil
}

func (adapter *GitLabAdapter) GetMergeRequestDiscussions(ctx context.Context, projectID, requestID int) ([]*domainRequests.Discussion, error) {
	extDiscussions, _, err := adapter.client.Discussions.ListMergeRequestDiscussions(
		projectID,
		requestID,
		nil,
		func(r *retryablehttp.Request) error {
			*r = *r.WithContext(ctx)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	discussions := make([]*domainRequests.Discussion, 0, len(extDiscussions))
	for _, discussion := range extDiscussions {
		discussions = append(discussions, FromDiscussionToDomain(discussion))
	}
	return discussions, nil
}
