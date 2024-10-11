# Gitlab MergeRequest reminder bot

Checks if merge requests have some unfinished discussions or do not have enough approvals.
Notifications are send to Mattermost through incoming webhooks.

## Checks

- unresolved discussions with the MR assignees answer - all partipants of the discussion, except MR assignees
- unresolved discussions with comments of non assignee - assignees are notified
- merge request has enough approvals - all participants are notified

## TODO

- Add database storage
- Move Checker and Notifier services into different jobs with their own schedule
- Add integration tests
