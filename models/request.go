// Package models define various models used in project
package models

// Request is a request from other bot to out server
type Request struct {
	// MessageId is a Telegram message ID
	MessageId int64 `json:"message_id"`

	// Chat is a Telegram chat entity
	Chat struct {
		// Chat -> Id is a Chat ID
		Id int64 `json:"id"`
	} `json:"chat"`

	// FromUser is a Telegram user entity
	FromUser struct {
		// FromUser -> Id is a User ID
		Id int64 `json:"id"`

		// FromUser -> FirstName is user or bot first name
		FirstName string `json:"first_name"`

		// FromUser -> LastName is user or bot last name
		LastName string `json:"last_name,omitempty"`

		// FromUser -> Username is user or bot username
		Username string `json:"username,omitempty"`
	} `json:"from_user"`

	// Reactions is a Telegram reactions list
	Reactions []struct {
		// Reactions -> Emoji is emoji itself
		Emoji string `json:"emoji"`

		// From is a Telegram user entity
		From struct {
			// From -> Id is a User ID which set reaction
			Id int64 `json:"id"`
		} `json:"from"`
	} `json:"reactions"`
}
