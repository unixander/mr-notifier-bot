package accesscontrol

import (
	"regexp"
	"review_reminder_bot/internal/infrastructure/config"
	"strings"
)

type AccessManager struct {
	allowedUsers         map[string]struct{}
	ignoredUsers         map[string]struct{}
	allowedRepositories  map[int]struct{}
	ignoredRepositories  map[int]struct{}
	ignoredWebUrlsRegexp *regexp.Regexp
}

func New(cfg *config.Settings) *AccessManager {
	manager := &AccessManager{}
	fillLookupCache(cfg.AllowedUsers, &manager.allowedUsers)
	fillLookupCache(cfg.IgnoredUsers, &manager.ignoredUsers)
	fillLookupCache(cfg.AllowedRepositories, &manager.allowedRepositories)
	fillLookupCache(cfg.IgnoredRepositories, &manager.ignoredRepositories)
	if len(cfg.IgnoredWebUrlsRegexp) > 0 {
		manager.ignoredWebUrlsRegexp = regexp.MustCompile(strings.Join(cfg.IgnoredWebUrlsRegexp, "|"))
	}
	return manager
}

func fillLookupCache[T comparable, K []T, P map[T]struct{}](source K, target *P) {
	if len(source) == 0 {
		return
	}
	result := make(P, len(source))

	for _, item := range source {
		result[item] = struct{}{}
	}
	*target = result
}

func (manager *AccessManager) InAllowedUsers(username string) bool {
	if len(manager.allowedUsers) == 0 {
		return true
	}
	_, ok := manager.allowedUsers[username]
	return ok
}

func (manager *AccessManager) InIgnoredUsers(username string) bool {
	_, ok := manager.ignoredUsers[username]
	return ok
}

func (manager *AccessManager) IsUserAllowed(username string) bool {
	return !manager.InIgnoredUsers(username) && manager.InAllowedUsers(username)
}

func (manager *AccessManager) InIgnoredRepositories(repoID int) bool {
	_, ok := manager.ignoredRepositories[repoID]
	return ok
}

func (manager *AccessManager) InAllowedRepositories(repoID int) bool {
	if len(manager.allowedRepositories) == 0 {
		return true
	}
	_, ok := manager.allowedRepositories[repoID]
	return ok
}

func (manager *AccessManager) IsRepositoryAllowed(repoID int) bool {
	return !manager.InIgnoredRepositories(repoID) && manager.InAllowedRepositories(repoID)
}

func (manager *AccessManager) IsWebUrlAllowed(weburl string) bool {
	if manager.ignoredWebUrlsRegexp == nil {
		return true
	}
	return !manager.ignoredWebUrlsRegexp.MatchString(weburl)
}
