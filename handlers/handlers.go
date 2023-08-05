// Package handlers is used to handle handlers (🐳).
package handlers

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/xbt573/flood-social-rep/database"
	"sort"
)

// Handle is a function which adds handlers to dispatcher.
func Handle(dispatcher *ext.Dispatcher) {
	// Rating-related commands
	dispatcher.AddHandler(handlers.NewCommand("liketop", liketop))
	dispatcher.AddHandler(handlers.NewCommand("disliketop", disliketop))
	dispatcher.AddHandler(handlers.NewCommand("whaletop", whaletop))
	dispatcher.AddHandler(handlers.NewCommand("rep", rep))
}

// Like top handler
func liketop(bot *gotgbot.Bot, ctx *ext.Context) error {
	top, err := database.TopRating(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	sort.SliceStable(top, func(i, j int) bool {
		return top[i].Likes > top[j].Likes
	})

	if len(top) >= 10 {
		top = top[:9]
	}

	topStr := "Топ рейтинга:"

	for _, topPlace := range top {
		if topPlace.Likes == 0 {
			continue
		}

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
			"\n%v: %v 👍 %v 👎 %v 🐳",
			username,
			topPlace.Likes,
			topPlace.Dislikes,
			topPlace.Whales,
		)
	}

	_, err = ctx.EffectiveMessage.Reply(bot, topStr, nil)
	if err != nil {
		return err
	}

	return nil
}

// Dislike top handler
func disliketop(bot *gotgbot.Bot, ctx *ext.Context) error {
	top, err := database.TopRating(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	sort.SliceStable(top, func(i, j int) bool {
		return top[i].Dislikes > top[j].Dislikes
	})

	if len(top) >= 10 {
		top = top[:9]
	}

	topStr := "Топ рейтинга (наоборот):"

	for _, topPlace := range top {
		if topPlace.Dislikes == 0 {
			continue
		}

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
			"\n%v: %v 👎 %v 👍 %v 🐳",
			username,
			topPlace.Dislikes,
			topPlace.Likes,
			topPlace.Whales,
		)
	}

	_, err = ctx.EffectiveMessage.Reply(bot, topStr, nil)
	if err != nil {
		return err
	}

	return nil
}

// Whale reputation top handler
func whaletop(bot *gotgbot.Bot, ctx *ext.Context) error {
	top, err := database.TopRating(ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}

	sort.SliceStable(top, func(i, j int) bool {
		return top[i].Whales > top[j].Whales
	})

	if len(top) >= 10 {
		top = top[:9]
	}

	topStr := "Топ рейтинга по китам:"

	for _, topPlace := range top {
		if topPlace.Whales == 0 {
			continue
		}

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
			"\n%v: %v 🐳 %v 👍 %v 👎",
			username,
			topPlace.Whales,
			topPlace.Likes,
			topPlace.Dislikes,
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

	message := fmt.Sprintf(
		"\n%v: %v 👍 %v 👎 %v 🐳",
		username,
		rating.Likes,
		rating.Dislikes,
		rating.Whales,
	)

	_, err = ctx.EffectiveMessage.Reply(bot, message, nil)
	if err != nil {
		return err
	}

	return nil
}
