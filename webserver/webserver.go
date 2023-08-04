// Package webserver is responsible for fast webserver creation
package webserver

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/xbt573/flood-social-rep/database"
	"github.com/xbt573/flood-social-rep/models"
)

// New is a function for creating webserver instance
func New() *fiber.App {
	app := fiber.New(fiber.Config{
		// Remove this fucking fancy banner
		DisableStartupMessage: true,
	})

	// Fuck you CORS
	app.Use(cors.New(cors.ConfigDefault))

	app.Post("/reactions", func(ctx *fiber.Ctx) error {
		var request models.Request

		if err := ctx.BodyParser(&request); err != nil {
			return err
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
