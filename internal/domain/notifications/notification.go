package notifications

type NotificationType string

const (
	AwaitingReview         = NotificationType("awaitingReview")
	AwaitingThreadResponse = NotificationType("awaitingThreadResponse")
	AwaitingThreadResolve  = NotificationType("awaitingThreadResolve")
	AwaitingPipelineFix    = NotificationType("awaitingPipelineFix")
)

type Notification struct {
	UserName  string
	RequestID int
	ProjectID int
	Type      NotificationType
	WebURL    string
}

func (notif Notification) Link() string {
	return notif.WebURL
}
