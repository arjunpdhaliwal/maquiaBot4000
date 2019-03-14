package handlers

import (
	pokemoncommands "./pokemon-commands"
	"github.com/bwmarrin/discordgo"
)

// PokemonHandle handles commands that are regarding pokemon
func PokemonHandle(s *discordgo.Session, m *discordgo.MessageCreate, args []string, serverPrefix string) {
	// Check if any args were even given
	if len(args) > 1 {
		mainArg := args[1]
		switch mainArg {
		default:
			go pokemoncommands.Pokemon(s, m, args)
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "Please specify a command! Check `"+serverPrefix+"help` for more details!")
	}
}