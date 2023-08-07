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

		for _, reaction := range request.Reactions {
			err := database.AddReaction(
				request.Chat.Id,
				reaction.From.Id,
				request.FromUser.Id,
				request.MessageId,
				reaction.Emoji,
			)
			if err != nil {
				return ctx.Status(500).SendString(err.Error())
			}

			username := request.FromUser.Username
			if username == "" {
				username = request.FromUser.FirstName
				if request.FromUser.LastName != "" {
					username = username + " " + request.FromUser.LastName
				}
			}

			err = database.UpdateUsername(request.FromUser.Id, username)
			if err != nil {
				return ctx.Status(500).SendString(err.Error())
			}
		}

		return nil
	})

	return app
}
