package util

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func GenerateEmbedFooter(action string) *discordgo.MessageEmbedFooter {
	return &discordgo.MessageEmbedFooter{
		Text:    fmt.Sprintf("%s • ykoo.cc 2024", action),
		IconURL: "https://cdn3.emoji.gg/emojis/46591-aquaannoyed.png",
	}
}

const (
	ColorRed       = 0xED4245
	ColorGreen     = 0x57F287
	ColorBlue      = 0x3498DB
	ColorGold      = 0xF1C40F
	ColorOrange    = 0xE67E22
	ColorPurple    = 0x9B59B6
	ColorPink      = 0xE91E63
	ColorTeal      = 0x1ABC9C
	ColorWhite     = 0xFFFFFF
	ColorBlack     = 0x000000
	ColorDarkGray  = 0x979C9F
	ColorLightGray = 0xBCC0C0
)

type EmbedStyle struct {
	Color        int
	AuthorIcon   string
	FooterIcon   string
	FooterText   string
	ThumbnailURL string
}

var (
	StyleDefault = EmbedStyle{
		Color: ColorBlue,
	}
	StyleSuccess = EmbedStyle{
		Color: ColorGreen,
	}
	StyleError = EmbedStyle{
		Color: ColorRed,
	}
	StyleWarning = EmbedStyle{
		Color: ColorOrange,
	}
)

type EmbedBuilder struct {
	embed *discordgo.MessageEmbed
}

func NewEmbed(style EmbedStyle, title, description string) *EmbedBuilder {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       style.Color,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if style.FooterText != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text:    style.FooterText,
			IconURL: style.FooterIcon,
		}
	}

	if style.AuthorIcon != "" {
		embed.Author = &discordgo.MessageEmbedAuthor{
			IconURL: style.AuthorIcon,
		}
	}

	if style.ThumbnailURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: style.ThumbnailURL,
		}
	}

	return &EmbedBuilder{embed: embed}
}

func (eb *EmbedBuilder) WithField(name, value string, inline bool) *EmbedBuilder {
	eb.embed.Fields = append(eb.embed.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return eb
}

func (eb *EmbedBuilder) WithImage(imageURL string) *EmbedBuilder {
	eb.embed.Image = &discordgo.MessageEmbedImage{
		URL: imageURL,
	}
	return eb
}

func (eb *EmbedBuilder) WithAuthor(name, iconURL, url string) *EmbedBuilder {
	eb.embed.Author = &discordgo.MessageEmbedAuthor{
		Name:    name,
		IconURL: iconURL,
		URL:     url,
	}
	return eb
}

func (eb *EmbedBuilder) WithFooter(action string) *EmbedBuilder {
	eb.embed.Footer = GenerateEmbedFooter(action)
	return eb
}

func (eb *EmbedBuilder) WithFooterCustom(text, iconURL string) *EmbedBuilder {
	eb.embed.Footer = &discordgo.MessageEmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return eb
}

func (eb *EmbedBuilder) WithThumbnail(url string) *EmbedBuilder {
	eb.embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: url,
	}
	return eb
}

func (eb *EmbedBuilder) Build() *discordgo.MessageEmbed {
	return eb.embed
}

func (eb *EmbedBuilder) WithColor(color int) *EmbedBuilder {
	eb.embed.Color = color
	return eb
}

func MustInt(value string) int {
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	return valueInt
}
