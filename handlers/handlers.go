// Package handlers is used to handle handlers (🐳).
package handlers

import (
	"errors"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/xbt573/flood-social-rep/database"
	"sort"
	"strconv"
)

// Handle is a function which adds handlers to dispatcher.
func Handle(dispatcher *ext.Dispatcher) {
	// Rating-related commands
	dispatcher.AddHandler(handlers.NewCommand("liketop", liketop))
	dispatcher.AddHandler(handlers.NewCommand("disliketop", disliketop))
	dispatcher.AddHandler(handlers.NewCommand("whaletop", whaletop))
	dispatcher.AddHandler(handlers.NewCommand("repignore", repignore))
	dispatcher.AddHandler(handlers.NewCommand("repunignore", repunignore))
	dispatcher.AddHandler(handlers.NewCommand("rep", rep))
	dispatcher.AddHandler(handlers.NewCommand("reactions", reactions))
}

func repignore(bot *gotgbot.Bot, ctx *ext.Context) error {
	member, err := bot.GetChatMember(ctx.EffectiveChat.Id, ctx.EffectiveUser.Id, nil)
	if err != nil {
		return err
	}

	if member.GetStatus() == "member" {
		_, err := ctx.EffectiveMessage.Reply(bot, "у тебя нет прав ALO🔉🔉🔉", nil)
		if err != nil {
			return err
		}

		return nil
	}

	if ctx.EffectiveMessage.ReplyToMessage == nil {
		_, err := ctx.EffectiveMessage.Reply(bot, "Команда должна быть ответом", nil)
		if err != nil {
			return err
		}

		return nil
	}

	err = database.AddBlacklist(
		ctx.EffectiveChat.Id,
		ctx.EffectiveMessage.ReplyToMessage.From.Id,
	)
	if err != nil {
		if !errors.Is(err, database.ErrAlreadyBlacklisted) {
			return err
		}

		_, err := ctx.EffectiveMessage.Reply(bot, "Юзер уже в игноре", nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func repunignore(bot *gotgbot.Bot, ctx *ext.Context) error {
	member, err := bot.GetChatMember(ctx.EffectiveChat.Id, ctx.EffectiveUser.Id, nil)
	if err != nil {
		return err
	}

	if member.GetStatus() == "member" {
		_, err := ctx.EffectiveMessage.Reply(bot, "у тебя нет прав ALO🔉🔉🔉", nil)
		if err != nil {
			return err
		}

		return nil
	}

	if ctx.EffectiveMessage.ReplyToMessage == nil {
		_, err := ctx.EffectiveMessage.Reply(bot, "Команда должна быть ответом", nil)
		if err != nil {
			return err
		}

		return nil
	}

	err = database.RemoveBlacklist(
		ctx.EffectiveChat.Id,
		ctx.EffectiveMessage.ReplyToMessage.From.Id,
	)
	if err != nil {
		if !errors.Is(err, database.ErrNotInBlacklist) {
			return err
		}

		_, err := ctx.EffectiveMessage.Reply(bot, "Юзер уже не в игноре", nil)
		if err != nil {
			return err
		}
	}

	return nil
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

		var username string

		member, err := bot.GetChatMember(ctx.EffectiveChat.Id, topPlace.UserId, nil)
		if err != nil {
			name, err := database.GetUsername(topPlace.UserId)
			if err != nil {
				continue
			}

			username = name
		} else {
			username = member.GetUser().Username
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

		var username string

		member, err := bot.GetChatMember(ctx.EffectiveChat.Id, topPlace.UserId, nil)
		if err != nil {
			name, err := database.GetUsername(topPlace.UserId)
			if err != nil {
				continue
			}

			username = name
		} else {
			username = member.GetUser().Username
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

		var username string

		member, err := bot.GetChatMember(ctx.EffectiveChat.Id, topPlace.UserId, nil)
		if err != nil {
			name, err := database.GetUsername(topPlace.UserId)
			if err != nil {
				continue
			}

			username = name
		} else {
			username = member.GetUser().Username
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
		"%v: %v 👍 %v 👎 %v 🐳",
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

func reactions(bot *gotgbot.Bot, ctx *ext.Context) error {
	var id int64

	if ctx.EffectiveMessage.ReplyToMessage != nil {
		id = ctx.EffectiveMessage.ReplyToMessage.MessageId
	}

	args := ctx.Args()
	if len(args) > 0 {
		num, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			_, err := ctx.EffectiveMessage.Reply(bot, "Failed to parse id", nil)
			if err != nil {
				return err
			}

			return err
		}

		id = num
	}

	reactions, err := database.GetReactions(ctx.EffectiveChat.Id, id)
	if err != nil {
		return err
	}

	if len(reactions) == 0 {
		_, err := ctx.EffectiveMessage.Reply(bot, "No reactions found", nil)
		if err != nil {
			return err
		}

		return nil
	}

	var resStr string

	for _, x := range reactions {
		var username string

		member, err := bot.GetChatMember(ctx.EffectiveChat.Id, x.UserId, nil)
		if err != nil {
			name, err := database.GetUsername(x.UserId)
			if err != nil {
				continue
			}

			username = name
		} else {
			username = member.GetUser().Username
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
		}

		resStr += fmt.Sprintf("%v - %v\n", username, x.Reaction)
	}

	_, err = ctx.EffectiveMessage.Reply(bot, resStr, nil)
	if err != nil {
		return err
	}

	return nil
}
