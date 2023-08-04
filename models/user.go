// Package models define various models used in project
package models

// User is structure describing user, and it's rating (per group)
type User struct {
	// User ID, 64 bit
	UserId int64

	// User rating
	Rating int
}
