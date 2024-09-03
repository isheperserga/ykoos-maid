package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"

	"yk-dc-bot/internal/bot"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/database"
	_ "yk-dc-bot/internal/handlers"
	"yk-dc-bot/internal/henrikapi"
	"yk-dc-bot/internal/logger"
	"yk-dc-bot/internal/redisclient"
	"yk-dc-bot/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := fx.New(
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			database.NewPostgresDB,
			redisclient.NewRedisClient,
			henrikapi.NewHenrikDevAPI,
			service.NewService,
			bot.NewDiscordBot,
		),
		fx.Invoke(runBot),
	)

	if err := app.Start(ctx); err != nil {
		os.Exit(1)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if err := app.Stop(ctx); err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
	}
}

func runBot(lc fx.Lifecycle, bot *bot.DiscordBot, log *logger.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting bot...")
			return bot.Run(ctx)
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping bot...")
			return bot.Stop(ctx)
		},
	})
}
