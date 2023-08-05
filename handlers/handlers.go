// Package handlers is used to handle handlers (🐳).
package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/xbt573/flood-social-rep/database"
)

// Handle is a function which adds handlers to dispatcher.
func Handle(dispatcher *ext.Dispatcher) {
	// Debug command, will be deleted on release
	// TODO: REMOVE ON RELEASE
	dispatcher.AddHandler(handlers.NewCommand("debug", debug))

	// Rating-related commands
	dispatcher.AddHandler(handlers.NewCommand("reptop", reptop))
	dispatcher.AddHandler(handlers.NewCommand("revreptop", revreptop))
	dispatcher.AddHandler(handlers.NewCommand("rep", rep))
}

// Reputation top handler
func reptop(bot *gotgbot.Bot, ctx *ext.Context) error {
	top, err := database.TopRating(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	topStr := "Топ рейтинга:"

	for _, topPlace := range top {
		member, err := bot.GetChatMember(ctx.EffectiveChat.Id, topPlace.UserId, nil)
		if err != nil {
			continue
		}

		// If username is doesn't exist - create it from first and last name
		username := member.GetUser().Username
		if username == "" {
			username = member.GetUser().FirstName
			if member.GetUser().LastName != "" {
				username = fmt.Sprintf(
					"%v %v",
					member.GetUser().FirstName,
					member.GetUser().LastName,
				)
			}
		}

		topStr += fmt.Sprintf(
			"\n%v: %v",
			username,
			topPlace.Rating,
		)
	}

	_, err = ctx.EffectiveMessage.Reply(bot, topStr, nil)
	if err != nil {
		return err
	}

	return nil
}

// Reverse reputation top handler
func revreptop(bot *gotgbot.Bot, ctx *ext.Context) error {
	top, err := database.TopReverseRating(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	topStr := "Топ рейтинга (наоборот):"

	for _, topPlace := range top {
		member, err := bot.GetChatMember(ctx.EffectiveChat.Id, topPlace.UserId, nil)
		if err != nil {
			continue
		}

		// If username is doesn't exist - create it from first and last name
		username := member.GetUser().Username
		if username == "" {
			username = member.GetUser().FirstName
			if member.GetUser().LastName != "" {
				username = fmt.Sprintf(
					"%v %v",
					member.GetUser().FirstName,
					member.GetUser().LastName,
				)
			}
		}

		topStr += fmt.Sprintf(
			"\n%v: %v",
			username,
			topPlace.Rating,
		)
	}

	_, err = ctx.EffectiveMessage.Reply(bot, topStr, nil)
	if err != nil {
		return err
	}

	return nil
}

// Reputation handler
func rep(bot *gotgbot.Bot, ctx *ext.Context) error {
	userId := ctx.EffectiveMessage.From.Id
	username := ctx.EffectiveMessage.From.Username

	// If username is doesn't exist - create it from first and last name
	if username == "" {
		username = ctx.EffectiveMessage.From.FirstName
		if ctx.EffectiveMessage.From.LastName != "" {
			username = fmt.Sprintf(
				"%v %v",
				ctx.EffectiveMessage.From.FirstName,
				ctx.EffectiveMessage.From.LastName,
			)
		}
	}

	// If command is a reply - do the same but for replied message
	if ctx.EffectiveMessage.ReplyToMessage != nil {
		// Ignore bot rep requests
		if ctx.EffectiveMessage.ReplyToMessage.From.IsBot {
			return nil
		}

		userId = ctx.EffectiveMessage.ReplyToMessage.From.Id
		username = ctx.EffectiveMessage.ReplyToMessage.From.Username
		if username == "" {
			username = ctx.EffectiveMessage.ReplyToMessage.From.FirstName
			if ctx.EffectiveMessage.ReplyToMessage.From.LastName != "" {
				username = fmt.Sprintf(
					"%v %v",
					ctx.EffectiveMessage.ReplyToMessage.From.FirstName,
					ctx.EffectiveMessage.ReplyToMessage.From.LastName,
				)
			}
		}
	}

	rating, err := database.GetUserRating(ctx.EffectiveChat.Id, userId)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("%v: %v", username, rating.Rating)

	_, err = ctx.EffectiveMessage.Reply(bot, message, nil)
	if err != nil {
		return err
	}

	return nil
}

// Debug handler
func debug(bot *gotgbot.Bot, ctx *ext.Context) error {
	str, err := json.MarshalIndent(ctx.EffectiveMessage, "", "    ")
	if err != nil {
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(bot, string(str), nil)
	if err != nil {
		return err
	}

	return nil
}
