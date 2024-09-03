package bot

import (
	"context"
	"fmt"

	"yk-dc-bot/internal/apperrors"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/handlers"
	"yk-dc-bot/internal/logger"
	"yk-dc-bot/internal/service"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	Session *discordgo.Session
	Service *service.Service
	Log     *logger.Logger
	Config  *config.Config
}

func NewDiscordBot(cfg *config.Config, service *service.Service, log *logger.Logger) (*DiscordBot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		return nil, apperrors.Wrap(err, "DISCORD_SESSION_ERROR", "error creating Discord session")
	}

	bot := &DiscordBot{
		Session: session,
		Service: service,
		Log:     log,
		Config:  cfg,
	}

	bot.registerHandlers()

	return bot, nil
}

func (bot *DiscordBot) Run(ctx context.Context) error {
	err := bot.Session.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}
	bot.Log.Info("Bot is now running. Press CTRL-C to exit.")

	<-ctx.Done()
	return nil
}

func (bot *DiscordBot) Stop(ctx context.Context) error {
	return bot.Session.Close()
}

func (bot *DiscordBot) registerHandlers() {
	bot.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			for _, handler := range handlers.CommandHandlers {
				if handler.Name == i.ApplicationCommandData().Name {
					handler.Handler(s, i, bot.Service, bot.Log, bot.Config)
					return
				}
			}
		case discordgo.InteractionModalSubmit:
			for _, handler := range handlers.ModalHandlers {
				if handler.CustomID == i.ModalSubmitData().CustomID {
					handler.Handler(s, i, bot.Service, bot.Log, bot.Config)
					return
				}
			}
		}
	})

	bot.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			for _, handler := range handlers.AutocompleteHandlers {
				if handler.Name == i.ApplicationCommandData().Name {
					handler.Handler(s, i, bot.Service, bot.Log, bot.Config)
					return
				}
			}
		}
	})
}
