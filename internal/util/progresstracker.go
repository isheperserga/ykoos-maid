package util

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

type ProgressUpdate struct {
	Message       string
	Done          bool
	Error         error
	PublicMessage string
}

type ProgressTracker struct {
	Session     *discordgo.Session
	Interaction *discordgo.Interaction
	Updates     chan ProgressUpdate
	EmbedStyle  EmbedStyle
	Title       string
	Footer      string
	done        chan struct{}
	once        sync.Once
}

func NewProgressTracker(s *discordgo.Session, i *discordgo.Interaction, title, footer string, style EmbedStyle) *ProgressTracker {
	return &ProgressTracker{
		Session:     s,
		Interaction: i,
		Updates:     make(chan ProgressUpdate),
		EmbedStyle:  style,
		Title:       title,
		Footer:      footer,
		done:        make(chan struct{}),
	}
}

func (pt *ProgressTracker) Start() {
	go pt.trackProgress()
}

func (pt *ProgressTracker) trackProgress() {
	defer pt.Stop()
	for {
		select {
		case update, ok := <-pt.Updates:
			if !ok {
				return
			}
			if update.Error != nil {
				errorMessage := update.Error.Error()
				if update.PublicMessage != "" {
					errorMessage = update.PublicMessage
				}
				errorEmbed := NewEmbed(StyleError, "Error", errorMessage).
					WithFooter(pt.Footer).
					Build()
				_, _ = pt.Session.InteractionResponseEdit(pt.Interaction, &discordgo.WebhookEdit{
					Embeds: &[]*discordgo.MessageEmbed{errorEmbed},
				})
				return
			}
			if update.Done {
				return
			}
			progressEmbed := NewEmbed(pt.EmbedStyle, pt.Title, update.Message).
				WithFooter(pt.Footer).
				Build()
			_, _ = pt.Session.InteractionResponseEdit(pt.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{progressEmbed},
			})
		case <-pt.done:
			return
		}
	}
}

func (pt *ProgressTracker) Stop() {
	pt.once.Do(func() {
		close(pt.done)
		close(pt.Updates)
	})
}

func (pt *ProgressTracker) SendUpdate(message string) {
	select {
	case pt.Updates <- ProgressUpdate{Message: message}:
	case <-pt.done:
	}
}

func (pt *ProgressTracker) SendError(err error, publicMessage string) {
	select {
	case pt.Updates <- ProgressUpdate{Error: err, PublicMessage: publicMessage}:
	case <-pt.done:
	}
}

func (pt *ProgressTracker) SendDone() {
	select {
	case pt.Updates <- ProgressUpdate{Done: true}:
	case <-pt.done:
	}
}
