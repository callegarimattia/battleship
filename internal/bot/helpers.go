package bot

import (
	"sync"
)

// Helper functions for tracking players, matches, and channels

// trackPlayer associates a player ID with their Discord user ID.
func (b *DiscordBot) trackPlayer(playerID, discordUserID string) {
	b.discordMu.Lock()
	b.playerToDiscord[playerID] = discordUserID
	b.discordMu.Unlock()
}

// trackMatch stores the active match for a Discord user.
func (b *DiscordBot) trackMatch(discordUserID, matchID string) {
	b.matchMu.Lock()
	b.activeMatches[discordUserID] = matchID
	b.matchMu.Unlock()
}

// trackChannel stores the channel ID for a match.
func (b *DiscordBot) trackChannel(matchID, channelID string) {
	b.channelMu.Lock()
	b.matchToChannel[matchID] = channelID
	b.channelMu.Unlock()
}

// getActiveMatch retrieves the active match for a Discord user.
func (b *DiscordBot) getActiveMatch(discordUserID string) (string, bool) {
	b.matchMu.RLock()
	defer b.matchMu.RUnlock()
	matchID, ok := b.activeMatches[discordUserID]
	return matchID, ok
}

// registerMatch is a convenience function that tracks player, match, and channel.
func (b *DiscordBot) registerMatch(playerID, discordUserID, matchID, channelID string) {
	// Use a single lock acquisition pattern for efficiency
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		b.trackPlayer(playerID, discordUserID)
	}()

	go func() {
		defer wg.Done()
		b.trackMatch(discordUserID, matchID)
	}()

	go func() {
		defer wg.Done()
		b.trackChannel(matchID, channelID)
	}()

	wg.Wait()
}
