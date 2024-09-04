package handlers

import (
	"fmt"
	"strings"

	"yk-dc-bot/internal/apperrors"
	"yk-dc-bot/internal/commands"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/logger"
	"yk-dc-bot/internal/service"
	"yk-dc-bot/internal/util"

	"github.com/bwmarrin/discordgo"
)

func init() {
	commands.AddRegistration(func() {
		commands.Register(&commands.Command{
			Name:        "tracker",
			Description: "Get a player's Valorant tracker stats",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "The player's Valorant username (e.g., username#tag)",
					Required:    true,
				},
			},
			Handler: handleTrackerCommand,
		})
	})
}

func handleTrackerCommand(s *discordgo.Session, i *discordgo.InteractionCreate, svc *service.Service, log *logger.Logger, cfg *config.Config) {
	options := i.ApplicationCommandData().Options
	if len(options) < 1 {
		util.RespondToInteraction(s, i, util.InteractionResponse{
			Content:   "please provide a valid valorant username and tag (e.g., /tracker username#tag)",
			Ephemeral: true,
		})
		return
	}

	fullUsername := options[0].StringValue()
	parts := strings.Split(fullUsername, "#")
	if len(parts) != 2 {
		util.RespondToInteraction(s, i, util.InteractionResponse{
			Content:   "invalid riot id. please use the username#tag format",
			Ephemeral: true,
		})
		return
	}

	name, tag := parts[0], parts[1]

	err := util.DeferResponse(s, i, util.DeferResponseOptions{
		Ephemeral:     false,
		CustomContent: "",
		Embeds: []*discordgo.MessageEmbed{
			util.NewEmbed(util.StyleDefault, fmt.Sprintf("fetching tracker data for %s#%s", name, tag), "> please wait a moment").
				WithFooter("valorant integration").
				Build(),
		},
	})
	if err != nil {
		log.Error("Error deferring response", "error", err)
		return
	}

	tracker := util.NewProgressTracker(s, i.Interaction, fmt.Sprintf("fetching tracker data for %s#%s", name, tag), "valorant integration", util.StyleDefault)
	tracker.Start()

	playerData, err := svc.GetPlayerTrackerData(name, tag, tracker)
	if err != nil {
		errorMessage, logMessage := apperrors.HandleError(err, "getting player tracker data")
		log.Error(logMessage)
		util.SendErrorEmbed(s, i.Interaction, errorMessage, log, "valorant integration")
		return
	}

	var embed *discordgo.MessageEmbed

	if playerData.IsPrivate {
		embed = util.NewEmbed(util.StyleWarning, fmt.Sprintf("%s#%s's Tracker Stats", name, tag), "This profile is private").
			WithColor(util.ColorGold).
			Build()
	} else {
		embed = util.NewEmbed(util.StyleSuccess, fmt.Sprintf("%s#%s's Tracker Stats", name, tag), "").
			WithColor(util.ColorGreen).
			WithField("Wins / Losses", fmt.Sprintf("%s / %s (%s winrate)", playerData.Wins, playerData.Losses, playerData.WinPct), true).
			WithField("Headshot %", playerData.HsPct, true).
			WithField("K/D Ratio", playerData.KdRatio, true).
			WithField("Damage Per Round", playerData.DamagePerRound, true).
			WithField("Time Played", playerData.TimePlayed, true).
			WithField("Rank", playerData.Rank, true).
			WithThumbnail(playerData.AvatarUrl).
			WithFooter("valorant integration").
			Build()
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		log.Error("Error editing final interaction response", "error", apperrors.Wrap(err, "INTERACTION_EDIT_ERROR", "failed to edit interaction response"))
	}
}
