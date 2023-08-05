// Package cmd runs Telegram bot and webserver, abstracted from package main.
package cmd

import (
	"errors"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	_ "github.com/joho/godotenv/autoload"
	"github.com/xbt573/flood-social-rep/database"
	"github.com/xbt573/flood-social-rep/handlers"
	"github.com/xbt573/flood-social-rep/webserver"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Run function runs Telegram bot and webserver.
// Returns non-nil error if something goes wrong.
func Run() error {
	slog.Info("Starting flood-social-rep")

	// Database initialization
	err := database.Init()
	if err != nil {
		slog.Error(
			"Failed database init!",
			slog.String("err", err.Error()),
		)
		return err
	}

	// Looking up environment variables
	token, exists := os.LookupEnv("BOT_TOKEN")
	if !exists {
		slog.Error("BOT_TOKEN variable does not exist!")
		return errors.New("BOT_TOKEN variable does not exist")
	}

	port, exists := os.LookupEnv("WEB_PORT")
	if !exists {
		slog.Error("WEB_PORT variable does not exist!")
		return errors.New("WEB_PORT variable does not exist")
	}

	keyEnabledStr := os.Getenv("KEY_ENABLED")

	var keyEnabled bool

	if keyEnabledStr != "false" && keyEnabledStr != "" {
		keyEnabled = true
	}

	key, exists := os.LookupEnv("KEY")
	if !exists {
		slog.Warn("KEY variable does not exist!")
	}

	// Creating bot instance
	bot, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
		Client: http.Client{},
		DefaultRequestOpts: &gotgbot.RequestOpts{
			Timeout: gotgbot.DefaultTimeout,
			APIURL:  gotgbot.DefaultAPIURL,
		},
	})

	if err != nil {
		slog.Error(
			"Failed initializing Telegram bot!",
			slog.String("err", err.Error()),
		)
		return err
	}

	updater := ext.NewUpdater(&ext.UpdaterOpts{
		Dispatcher: ext.NewDispatcher(&ext.DispatcherOpts{
			// If an error is returned by a handler, log it and continue going.
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				slog.Error(
					"error during processing!",
					slog.String("err", err.Error()),
				)
				return ext.DispatcherActionNoop
			},
			MaxRoutines: ext.DefaultMaxRoutines,
		}),
	})

	// Delegating handlers to handlers package
	handlers.Handle(updater.Dispatcher)

	// Create webserver instance
	app := webserver.New(key, keyEnabled)

	// errch is a channel for errors
	errch := make(chan error)

	// sigch is a channel for os interrupts
	sigch := make(chan os.Signal)
	signal.Notify(sigch, os.Interrupt)

	go func() {
		// Start bot
		err := updater.StartPolling(bot, &ext.PollingOpts{
			DropPendingUpdates: true,
			GetUpdatesOpts: gotgbot.GetUpdatesOpts{
				Timeout: 9,
				RequestOpts: &gotgbot.RequestOpts{
					Timeout: time.Second * 10,
				},
			},
		})

		if err != nil {
			errch <- err
		}

		updater.Idle()
	}()

	go func() {
		// Start webserver
		err := app.Listen(":" + port)
		if err != nil {
			errch <- err
			return
		}
	}()

	slog.Info("Started!")

	// Give error if found first, otherwise info about signal
	select {
	case x := <-errch:
		slog.Error(
			"Error while running!",
			slog.String("err", x.Error()),
		)

	case <-sigch:
		slog.Info("Caught exit signal, shutting down!")
	}

	// Stopping bot
	err = updater.Stop()
	if err != nil {
		slog.Error(
			"Failed to stop updater!",
			slog.String("err", err.Error()),
		)
	}

	// Stopping webserver
	err = app.Shutdown()
	if err != nil {
		slog.Error(
			"Failed to stop webserver!",
			slog.String("err", err.Error()),
		)
	}

	return nil
}
