package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "battleship",
		Description: "Play Battleship!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "host",
				Description: "Create a new game",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "join",
				Description: "Join an existing game",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "match_id",
						Description: "The match ID to join",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        "list",
				Description: "List available matches",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "place",
				Description: "Place a ship on your board",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "size",
						Description: "Ship size (2-5)",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    true,
						MinValue:    floatPtr(2),
						MaxValue:    5,
					},
					{
						Name:        "x",
						Description: "X coordinate (0-9)",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    true,
						MinValue:    floatPtr(0),
						MaxValue:    9,
					},
					{
						Name:        "y",
						Description: "Y coordinate (0-9)",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    true,
						MinValue:    floatPtr(0),
						MaxValue:    9,
					},
					{
						Name:        "vertical",
						Description: "Place ship vertically?",
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Required:    true,
					},
				},
			},
			{
				Name:        "attack",
				Description: "Attack a coordinate",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "x",
						Description: "X coordinate (0-9)",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    true,
						MinValue:    floatPtr(0),
						MaxValue:    9,
					},
					{
						Name:        "y",
						Description: "Y coordinate (0-9)",
						Type:        discordgo.ApplicationCommandOptionInteger,
						Required:    true,
						MinValue:    floatPtr(0),
						MaxValue:    9,
					},
				},
			},
			{
				Name:        "status",
				Description: "View your current game state",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	},
}

func floatPtr(f float64) *float64 {
	return &f
}

// registerCommands registers all slash commands with Discord.
func (b *DiscordBot) registerCommands() error {
	log.Println("Registering slash commands...")

	for _, cmd := range commands {
		_, err := b.session.ApplicationCommandCreate(b.appID, "", cmd)
		if err != nil {
			return err
		}
		log.Printf("Registered command: %s", cmd.Name)
	}

	return nil
}
