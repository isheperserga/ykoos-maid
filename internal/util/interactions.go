package util

import (
	"yk-dc-bot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

type InteractionResponse struct {
	Content    string
	Embeds     []*discordgo.MessageEmbed
	Components []discordgo.MessageComponent
	Ephemeral  bool
	TTS        bool
}

func RespondToInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, resp InteractionResponse) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    resp.Content,
			Embeds:     resp.Embeds,
			Components: resp.Components,
			Flags:      flagsFromOptions(resp),
			TTS:        resp.TTS,
		},
	})
}

func EditInteractionResponse(s *discordgo.Session, i *discordgo.Interaction, resp InteractionResponse) (*discordgo.Message, error) {
	return s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
		Content:    &resp.Content,
		Embeds:     &resp.Embeds,
		Components: &resp.Components,
	})
}

type DeferResponseOptions struct {
	Ephemeral     bool
	CustomContent string
	Embeds        []*discordgo.MessageEmbed
}

func DeferResponse(s *discordgo.Session, i *discordgo.InteractionCreate, options DeferResponseOptions) error {
	responseData := &discordgo.InteractionResponseData{
		Flags: flagsFromOptions(InteractionResponse{Ephemeral: options.Ephemeral}),
	}

	if options.CustomContent != "" {
		responseData.Content = options.CustomContent
	}

	if len(options.Embeds) > 0 {
		responseData.Embeds = options.Embeds
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: responseData,
	})
}

func FollowUpResponse(s *discordgo.Session, i *discordgo.Interaction, resp InteractionResponse) (*discordgo.Message, error) {
	return s.FollowupMessageCreate(i, true, &discordgo.WebhookParams{
		Content:    resp.Content,
		Embeds:     resp.Embeds,
		Components: resp.Components,
		Flags:      flagsFromOptions(resp),
		TTS:        resp.TTS,
	})
}

func EditFollowUpResponse(s *discordgo.Session, i *discordgo.Interaction, messageID string, resp InteractionResponse) (*discordgo.Message, error) {
	return s.FollowupMessageEdit(i, messageID, &discordgo.WebhookEdit{
		Content:    &resp.Content,
		Embeds:     &resp.Embeds,
		Components: &resp.Components,
	})
}

func flagsFromOptions(resp InteractionResponse) discordgo.MessageFlags {
	if resp.Ephemeral {
		return discordgo.MessageFlagsEphemeral
	}
	return 0
}

func SendErrorEmbed(s *discordgo.Session, i *discordgo.Interaction, errorMessage string, log *logger.Logger, footerString string) {
	errorEmbed := NewEmbed(StyleError, "Error", errorMessage).
		WithFooter(footerString).
		Build()

	_, editErr := s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{errorEmbed},
	})
	if editErr != nil {
		log.Error("Error editing interaction response with error message", "error", editErr)
	}
}
