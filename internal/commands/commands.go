package commands

import (
	"sync"

	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/logger"
	"yk-dc-bot/internal/service"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *service.Service, log *logger.Logger, cfg *config.Config)

type Command struct {
	Name        string
	Description string
	Options     []*discordgo.ApplicationCommandOption
	Handler     CommandHandler
}

var (
	registry = make(map[string]*Command)
	mu       sync.RWMutex
)

func Register(cmd *Command) {
	mu.Lock()
	defer mu.Unlock()
	registry[cmd.Name] = cmd
}

func Get(name string) (*Command, bool) {
	mu.RLock()
	defer mu.RUnlock()
	cmd, ok := registry[name]
	return cmd, ok
}

func GetAll() []*discordgo.ApplicationCommand {
	mu.RLock()
	defer mu.RUnlock()
	cmds := make([]*discordgo.ApplicationCommand, 0, len(registry))
	for _, cmd := range registry {
		cmds = append(cmds, &discordgo.ApplicationCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Options:     cmd.Options,
		})
	}
	return cmds
}
