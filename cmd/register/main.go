package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"

	"yk-dc-bot/internal/apperrors"
	"yk-dc-bot/internal/commands"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/logger"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			newDiscordSession,
		),
		fx.Invoke(registerCommands),
	)

	if err := app.Start(context.TODO()); err != nil {
		fmt.Printf("Error starting application: %v\n", err)
		os.Exit(1)
	}

	app.Stop(context.TODO())
}

func newDiscordSession(cfg *config.Config) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}
	return session, nil
}

func registerCommands(session *discordgo.Session, log *logger.Logger) error {
	log.Info("Registering commands...")

	err := session.Open()
	if err != nil {
		return apperrors.Wrap(err, "DISCORD_SESSION_OPEN_ERROR", "Error opening Discord session")
	}
	defer session.Close()

	cmds := commands.GetAll()

	for _, cmd := range cmds {
		log.Info(fmt.Sprintf("Registering command: %s", cmd.Name))
		_, err := session.ApplicationCommandCreate(session.State.User.ID, "", cmd)
		if err != nil {
			return apperrors.Wrap(err, "COMMAND_REGISTRATION_ERROR", fmt.Sprintf("Error registering command %s", cmd.Name))
		}
		log.Info(fmt.Sprintf("Successfully registered command: %s", cmd.Name))
	}

	log.Info("All commands registered successfully!")
	return nil
}
