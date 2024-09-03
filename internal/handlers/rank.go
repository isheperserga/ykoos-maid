package handlers

import (
	"fmt"
	"strings"

	"yk-dc-bot/internal/apperrors"
	"yk-dc-bot/internal/config"
	"yk-dc-bot/internal/logger"
	"yk-dc-bot/internal/service"
	"yk-dc-bot/internal/util"

	"github.com/bwmarrin/discordgo"
)

func init() {
	CommandHandlers = append(CommandHandlers, CommandHandler{
		Name:    "rank",
		Handler: handleRankCommand,
	})
}

func handleRankCommand(s *discordgo.Session, i *discordgo.InteractionCreate, svc *service.Service, log *logger.Logger, cfg *config.Config) {
	options := i.ApplicationCommandData().Options
	if len(options) < 1 {
		util.RespondToInteraction(s, i, util.InteractionResponse{
			Content:   "please provide a valid valorant username and tag (e.g., /rank username#tag)",
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
			util.NewEmbed(util.StyleDefault, fmt.Sprintf("fetching rank for %s#%s", name, tag), "> please wait a moment").
				WithFooter("valorant integration").
				Build(),
		},
	})
	if err != nil {
		log.Error("Error deferring response", "error", err)
		return
	}

	tracker := util.NewProgressTracker(s, i.Interaction, fmt.Sprintf("fetching rank for %s#%s", name, tag), "valorant integration", util.StyleDefault)
	tracker.Start()

	rankData, err := svc.GetPlayerRankData(name, tag, tracker)
	if err != nil {
		errorMessage, logMessage := apperrors.HandleError(err, "getting player rank data")
		log.Error(logMessage)
		util.SendErrorEmbed(s, i.Interaction, errorMessage, log)
		return
	}

	rankEmbed := util.NewEmbed(util.StyleSuccess, fmt.Sprintf("%s#%s", rankData.AccountName, rankData.AccountTag), "").
		WithColor(util.ColorGold).
		WithField("rank", "> "+rankData.Rank, false).
		WithField("ranked rating", "> "+fmt.Sprintf("%d/100", rankData.RR), false).
		WithField("last game", "> "+fmt.Sprintf("%+d rr", rankData.LastGameRR), false).
		WithThumbnail(rankData.CardURL).
		WithFooter("valorant integration").
		Build()

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{rankEmbed},
	})
	if err != nil {
		log.Error("Error editing final interaction response", "error", apperrors.Wrap(err, "INTERACTION_EDIT_ERROR", "failed to edit interaction response"))
	}
}
