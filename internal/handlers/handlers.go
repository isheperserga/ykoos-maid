package handlers

import (
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/logger"
	"yk-dc-bot/internal/service"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler struct {
	Name                     string
	Handler                  func(*discordgo.Session, *discordgo.InteractionCreate, *service.Service, *logger.Logger, *config.Config)
	Options                  []discordgo.ApplicationCommandOption
	DefaultMemberPermissions *int64
	DMPermission             *bool
}

type ModalHandler struct {
	CustomID string
	Handler  func(*discordgo.Session, *discordgo.InteractionCreate, *service.Service, *logger.Logger, *config.Config)
}

type AutocompleteHandler struct {
	Name    string
	Handler func(*discordgo.Session, *discordgo.InteractionCreate, *service.Service, *logger.Logger, *config.Config)
}

var (
	CommandHandlers      []CommandHandler
	ModalHandlers        []ModalHandler
	AutocompleteHandlers []AutocompleteHandler
)
