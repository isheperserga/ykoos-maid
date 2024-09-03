package handlers

import (
	"fmt"
	"strings"

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
			Content:   "Please provide a Valorant username and tag (e.g., /rank username#tag)",
			Ephemeral: true,
		})
		return
	}

	fullUsername := options[0].StringValue()
	parts := strings.Split(fullUsername, "#")
	if len(parts) != 2 {
		util.RespondToInteraction(s, i, util.InteractionResponse{
			Content:   "Invalid username format. Please use the format: username#tag",
			Ephemeral: true,
		})
		return
	}

	name, tag := parts[0], parts[1]

	err := util.DeferResponse(s, i, util.DeferResponseOptions{
		Ephemeral: false,
		Embeds: []*discordgo.MessageEmbed{
			util.NewEmbed(util.StyleDefault, "Fetching Rank", "Please wait...").
				WithFooter("valorant integration").
				Build(),
		},
	})
	if err != nil {
		log.Error("Error deferring response", "error", err)
		return
	}

	rankData, err := svc.GetPlayerRankData(name, tag)
	if err != nil {
		log.Error("Error getting player rank data", "error", err)
		errorEmbed := util.NewEmbed(util.StyleError, "Error", "An error occurred while fetching player data.").
			WithFooter("Rank Error").
			Build()
		util.EditInteractionResponse(s, i.Interaction, util.InteractionResponse{
			Embeds:    []*discordgo.MessageEmbed{errorEmbed},
			Ephemeral: true,
		})
		return
	}

	rankEmbed := util.NewEmbed(util.StyleSuccess, fmt.Sprintf("%s#%s's Rank", rankData.AccountName, rankData.AccountTag), "").
		WithColor(util.ColorGold).
		WithField("Rank", rankData.Rank, true).
		WithField("RR", fmt.Sprintf("%d/100", rankData.RR), true).
		WithField("Last Game", fmt.Sprintf("%+d", rankData.LastGameRR), true).
		WithThumbnail(rankData.CardURL).
		WithFooter("valorant integration").
		Build()

	_, err = util.EditInteractionResponse(s, i.Interaction, util.InteractionResponse{
		Embeds: []*discordgo.MessageEmbed{rankEmbed},
	})
	if err != nil {
		log.Error("Error editing interaction response", "error", err)
	}
}
