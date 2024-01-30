package subscriber

import "github.com/rudderlabs/rudder-cp-sdk/notifications"

// Subscriber is a subscriber to workspace config updates.
type Subscriber struct {
	notifications chan notifications.WorkspaceConfigNotification
}

func New() *Subscriber {
	return &Subscriber{
		notifications: make(chan notifications.WorkspaceConfigNotification),
	}
}

func (s *Subscriber) Notify(n notifications.WorkspaceConfigNotification) {
	s.notifications <- n
}

// Notifications returns a channel that will be notified of any updates to the workspace configs.
func (s *Subscriber) Notifications() chan notifications.WorkspaceConfigNotification {
	return s.notifications
}
