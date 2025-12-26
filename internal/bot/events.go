package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/callegarimattia/battleship/internal/events"
)

// subscribeToEvents subscribes the bot to game events.
func (b *DiscordBot) subscribeToEvents() {
	b.eventBus.Subscribe("*", b.handleGameEvent)
}

// handleGameEvent processes game events and sends notifications.
func (b *DiscordBot) handleGameEvent(event *events.GameEvent) {
	// Don't notify the player who triggered the event
	if event.TargetID == event.PlayerID {
		return
	}

	// Get channel ID for this match
	b.channelMu.RLock()
	channelID, ok := b.matchToChannel[event.MatchID]
	b.channelMu.RUnlock()

	if !ok || channelID == "" {
		return // No channel tracked for this match
	}

	// Create appropriate embed based on event type
	embed := b.formatEventEmbed(event)
	if embed == nil {
		return
	}

	// Get Discord user ID for target player to mention them
	b.discordMu.RLock()
	discordUserID := b.playerToDiscord[event.TargetID]
	b.discordMu.RUnlock()

	// Send message to channel with mention
	content := ""
	if discordUserID != "" {
		content = fmt.Sprintf("<@%s>", discordUserID)
	}

	if err := b.sendChannelMessage(channelID, content, embed); err != nil {
		log.Printf("Failed to send message to channel %s: %v", channelID, err)
	}
}

// formatEventEmbed creates an embed for the given event.
func (b *DiscordBot) formatEventEmbed(event *events.GameEvent) *discordgo.MessageEmbed {
	switch event.Type {
	case events.EventPlayerJoined:
		return &discordgo.MessageEmbed{
			Title:       "🎮 Player Joined!",
			Description: "A player has joined your game!",
			Color:       0x00ff00,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Match ID: %s", event.MatchID),
			},
		}

	case events.EventShipPlaced:
		return &discordgo.MessageEmbed{
			Title:       "🚢 Ship Placed",
			Description: "Your opponent placed a ship!",
			Color:       0x0099ff,
		}

	case events.EventAttackMade:
		data, ok := event.Data.(events.AttackEventData)
		if !ok {
			return nil
		}
		coord := CoordinateToChess(data.X, data.Y)
		return &discordgo.MessageEmbed{
			Title: "💥 Your Turn!",
			Description: fmt.Sprintf(
				"Your opponent attacked %s. Result: %s\n\nIt's your turn!",
				coord,
				data.Result,
			),
			Color: 0xff9900,
		}

	case events.EventGameStarted:
		return &discordgo.MessageEmbed{
			Title:       "🎯 Game Started!",
			Description: "Both players have placed all ships. The battle begins!",
			Color:       0x00ff00,
		}

	case events.EventGameOver:
		data, ok := event.Data.(events.GameOverEventData)
		if !ok {
			return nil
		}
		return &discordgo.MessageEmbed{
			Title:       "🏆 Game Over!",
			Description: fmt.Sprintf("Winner: %s", data.Winner),
			Color:       0xffd700,
		}

	default:
		return nil
	}
}

// sendChannelMessage sends a message to a Discord channel.
func (b *DiscordBot) sendChannelMessage(
	channelID, content string,
	embed *discordgo.MessageEmbed,
) error {
	_, err := b.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: content,
		Embeds:  []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		return fmt.Errorf("failed to send channel message: %w", err)
	}
	return nil
}
