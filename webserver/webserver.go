// Package webserver is responsible for fast webserver creation
package webserver

import (
	"github.com/gofiber/fiber/v2"
	"github.com/xbt573/flood-social-rep/database"
	"github.com/xbt573/flood-social-rep/models"
)

// New is a function for creating webserver instance
func New(key string, keyEnabled bool) *fiber.App {
	app := fiber.New(fiber.Config{
		// Remove this fucking fancy banner
		DisableStartupMessage: true,
	})

	app.Post("/reactions", func(ctx *fiber.Ctx) error {
		if query := ctx.Query("key"); query != key && keyEnabled {
			return ctx.SendStatus(403)
		}
		var request models.Request

		if err := ctx.BodyParser(&request); err != nil {
			return err
		}

		if request.FromUser.IsBot {
			return nil
		}

		for _, reaction := range request.Reactions {
			switch reaction.Emoji {
			case "👍": // Positive reactions
				fallthrough
			case "🔥":
				fallthrough
			case "❤️":
				fallthrough
			case "👏":
				fallthrough
			case "💯":
				err := database.IncrementUserRating(
					request.MessageId,
					request.Chat.Id,
					reaction.From.Id,
					request.FromUser.Id,
				)

				if err != nil {
					return err
				}

			case "🤡": // Negative reactions
				fallthrough
			case "💩":
				fallthrough
			case "🤮":
				fallthrough
			case "👎":
				err := database.DecrementUserRating(
					request.MessageId,
					request.Chat.Id,
					reaction.From.Id,
					request.FromUser.Id,
				)

				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	return app
}
