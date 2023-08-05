// Package models define various models used in project.
package models

import "time"

// Attempt is a type which describes last user reaction attempt.
type Attempt struct {
	// UserId is a Telegram user id.
	UserId int64

	// Time is a last reaction attempt time.
	Time time.Time
}
