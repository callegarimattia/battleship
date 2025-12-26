package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// handleInteraction is the main handler for all Discord interactions.
func (b *DiscordBot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	if data.Name != "battleship" {
		return
	}

	// Get subcommand
	if len(data.Options) == 0 {
		respondError(s, i, "No subcommand provided")
		return
	}

	subcommand := data.Options[0]
	ctx := context.Background()

	// Auto-login user with Discord ID
	userID := i.Member.User.ID
	username := i.Member.User.Username

	authResp, err := b.ctrl.Login(ctx, username, "discord", userID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Failed to authenticate: %v", err))
		return
	}

	playerID := authResp.User.ID

	// Route to appropriate handler
	switch subcommand.Name {
	case "host":
		b.handleHost(ctx, s, i, playerID)
	case "join":
		b.handleJoin(ctx, s, i, playerID, subcommand.Options)
	case "list":
		b.handleList(ctx, s, i)
	case "place":
		b.handlePlace(ctx, s, i, playerID, subcommand.Options)
	case "attack":
		b.handleAttack(ctx, s, i, playerID, subcommand.Options)
	case "status":
		b.handleStatus(ctx, s, i, playerID)
	default:
		respondError(s, i, "Unknown subcommand")
	}
}

func (b *DiscordBot) handleHost(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	playerID string,
) {
	matchID, err := b.ctrl.HostGameAction(ctx, playerID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Failed to create match: %v", err))
		return
	}

	// Register player, match, and channel
	discordUserID := i.Member.User.ID
	b.registerMatch(playerID, discordUserID, matchID, i.ChannelID)

	embed := &discordgo.MessageEmbed{
		Title: "🎮 Match Created!",
		Description: fmt.Sprintf(
			"Match ID: `%s`\n\nShare this ID with your opponent so they can join!",
			matchID,
		),
		Color: 0x00ff00,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use /battleship place to set up your ships",
		},
	}

	respondEmbed(s, i, embed, false) // Public announcement
}

func (b *DiscordBot) handleJoin(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	playerID string,
	options []*discordgo.ApplicationCommandInteractionDataOption,
) {
	matchID := options[0].StringValue()

	view, err := b.ctrl.JoinGameAction(ctx, matchID, playerID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Failed to join match: %v", err))
		return
	}

	// Register player and match (channel already tracked by host)
	discordUserID := i.Member.User.ID
	b.trackPlayer(playerID, discordUserID)
	b.trackMatch(discordUserID, matchID)

	embed := &discordgo.MessageEmbed{
		Title:       "✅ Joined Match!",
		Description: fmt.Sprintf("Match ID: `%s`\n\nGame State: %s", matchID, view.State),
		Color:       0x00ff00,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use /battleship place to set up your ships",
		},
	}

	respondEmbed(s, i, embed, true) // Ephemeral
}

func (b *DiscordBot) handleList(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	matches, err := b.ctrl.ListGamesAction(ctx)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Failed to list matches: %v", err))
		return
	}

	if len(matches) == 0 {
		embed := &discordgo.MessageEmbed{
			Title:       "📋 Available Matches",
			Description: "No matches available. Use `/battleship host` to create one!",
			Color:       0xffaa00,
		}
		respondEmbed(s, i, embed, true) // Ephemeral
		return
	}

	description := ""
	for _, match := range matches {
		description += fmt.Sprintf(
			"**%s** - Host: %s (%d/2 players)\n",
			match.ID,
			match.HostName,
			match.PlayerCount,
		)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📋 Available Matches",
		Description: description,
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use /battleship join <match_id> to join a match",
		},
	}

	respondEmbed(s, i, embed, true) // Ephemeral
}

func (b *DiscordBot) handlePlace(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	playerID string,
	options []*discordgo.ApplicationCommandInteractionDataOption,
) {
	// Get active match
	discordUserID := i.Member.User.ID
	matchID, ok := b.getActiveMatch(discordUserID)
	if !ok {
		respondError(
			s,
			i,
			"You are not in an active match. Use `/battleship host` or `/battleship join` first.",
		)
		return
	}

	// Extract options
	optMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range options {
		optMap[opt.Name] = opt
	}

	size := int(optMap["size"].IntValue())
	x := int(optMap["x"].IntValue())
	y := int(optMap["y"].IntValue())
	vertical := optMap["vertical"].BoolValue()

	view, err := b.ctrl.PlaceShipAction(ctx, matchID, playerID, size, x, y, vertical)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Failed to place ship: %v", err))
		return
	}

	embed := FormatGameState(&view)
	embed.Title = "🚢 Ship Placed!"
	respondEmbed(s, i, embed, true) // Ephemeral
}

func (b *DiscordBot) handleAttack(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	playerID string,
	options []*discordgo.ApplicationCommandInteractionDataOption,
) {
	// Get active match
	discordUserID := i.Member.User.ID
	matchID, ok := b.getActiveMatch(discordUserID)
	if !ok {
		respondError(
			s,
			i,
			"You are not in an active match. Use `/battleship host` or `/battleship join` first.",
		)
		return
	}

	optMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range options {
		optMap[opt.Name] = opt
	}

	x := int(optMap["x"].IntValue())
	y := int(optMap["y"].IntValue())

	view, err := b.ctrl.AttackAction(ctx, matchID, playerID, x, y)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Failed to attack: %v", err))
		return
	}

	embed := FormatGameState(&view)
	embed.Title = fmt.Sprintf("💥 Attack at (%d, %d)!", x, y)
	respondEmbed(s, i, embed, true) // Ephemeral
}

func (b *DiscordBot) handleStatus(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	playerID string,
) {
	// Get active match
	discordUserID := i.Member.User.ID
	matchID, ok := b.getActiveMatch(discordUserID)
	if !ok {
		respondError(
			s,
			i,
			"You are not in an active match. Use `/battleship host` or `/battleship join` first.",
		)
		return
	}

	view, err := b.ctrl.GetGameStateAction(ctx, matchID, playerID)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Failed to get game state: %v", err))
		return
	}

	embed := FormatGameState(&view)
	respondEmbed(s, i, embed, true) // Ephemeral
}

// Helper functions for responding

func respondEmbed(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	embed *discordgo.MessageEmbed,
	ephemeral bool,
) {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  flags,
		},
	})
	if err != nil {
		log.Printf("Failed to respond to interaction: %v", err)
	}
}

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	embed := &discordgo.MessageEmbed{
		Title:       "❌ Error",
		Description: message,
		Color:       0xff0000,
	}
	respondEmbed(s, i, embed, true) // Errors are always ephemeral
}
