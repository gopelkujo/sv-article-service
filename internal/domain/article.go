// Package domain contains core entities and repository contracts.
// Implementations live in other packages; this layer has no I/O.
package domain

import "time"

// Article status values stored in posts.status.
const (
	StatusPublish = "publish"
	StatusDraft   = "draft"
	StatusThrash  = "thrash"
)

// ValidStatuses lists all allowed article status values.
var ValidStatuses = []string{StatusPublish, StatusDraft, StatusThrash}

// Article is the domain entity mapped to the posts table.
type Article struct {
	ID          int
	Title       string
	Content     string
	Category    string
	CreatedDate time.Time
	UpdatedDate *time.Time
	Status      string
}

// IsValidStatus reports whether status is one of the allowed values.
func IsValidStatus(status string) bool {
	for _, allowed := range ValidStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}
