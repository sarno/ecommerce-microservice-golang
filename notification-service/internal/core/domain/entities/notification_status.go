package entities

// NotificationStatus defines the status of a notification.
type NotificationStatus string

const (
	// NotificationStatusPending indicates the notification is pending.
	NotificationStatusPending NotificationStatus = "PENDING"
	// NotificationStatusSent indicates the notification has been sent successfully.
	NotificationStatusSent NotificationStatus = "SENT"
	// NotificationStatusFailed indicates the notification sending has failed.
	NotificationStatusFailed NotificationStatus = "FAILED"
)

// String returns the string representation of the NotificationStatus.
func (ns NotificationStatus) String() string {
	return string(ns)
}
