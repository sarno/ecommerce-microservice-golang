package entities

// NotificationType defines the type of a notification.
type NotificationType string

const (
	// NotificationTypeEmail represents an email notification.
	NotificationTypeEmail NotificationType = "EMAIL"
	// NotificationTypePush represents a push notification.
	NotificationTypePush NotificationType = "PUSH"
)

// String returns the string representation of the NotificationType.
func (nt NotificationType) String() string {
	return string(nt)
}
