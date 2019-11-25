package gencommands

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	structs "../../structs"
	tools "../../tools"
	"github.com/bwmarrin/discordgo"
)

// Info gives information about the user
func Info(s *discordgo.Session, m *discordgo.MessageCreate, cache []structs.PlayerData) {
	userRegex, _ := regexp.Compile(`info\s+(.+)`)

	user := m.Author
	nickname := "N/A"
	roles := "N/A"
	var joinDate discordgo.Timestamp
	var err error
	if len(m.Mentions) == 1 {
		user = m.Mentions[0]
	} else {
		userTest := ""
		if userRegex.MatchString(m.Content) {
			userTest = userRegex.FindStringSubmatch(m.Content)[1]
		}
		user, err = s.User(userTest)
		if err != nil {
			server, err := s.Guild(m.GuildID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "This is not a server! Use their ID directly instead.")
				return
			}
			if userTest == "" {
				for _, member := range server.Members {
					if member.User.ID == m.Author.ID {
						user, _ = s.User(member.User.ID)
						nickname = member.Nick
						joinDate = member.JoinedAt
						roles = ""
						for _, role := range member.Roles {
							discordRole, _ := s.State.Role(server.ID, role)
							roles = roles + discordRole.Name + ", "
						}
						if roles != "" {
							roles = roles[:len(roles)-2]
						}
					}
				}
			} else {
				sort.Slice(server.Members, func(i, j int) bool {
					time1, _ := server.Members[i].JoinedAt.Parse()
					time2, _ := server.Members[j].JoinedAt.Parse()
					return time1.Unix() < time2.Unix()
				})
				for _, member := range server.Members {
					if strings.HasPrefix(strings.ToLower(member.User.Username), userTest) || strings.HasPrefix(strings.ToLower(member.Nick), userTest) {
						user, _ = s.User(member.User.ID)
						nickname = member.Nick
						joinDate = member.JoinedAt
						roles = ""
						for _, role := range member.Roles {
							discordRole, _ := s.State.Role(server.ID, role)
							roles = roles + discordRole.Name + ", "
						}
						if roles != "" {
							roles = roles[:len(roles)-2]
						}
						break
					}
				}
			}
		} else {
			server, err := s.Guild(m.GuildID)
			if err == nil {
				for _, member := range server.Members {
					if member.User.ID == user.ID {
						nickname = member.Nick
						joinDate = member.JoinedAt
						roles = ""
						for _, role := range member.Roles {
							discordRole, _ := s.State.Role(server.ID, role)
							roles = roles + discordRole.Name + ", "
						}
						if roles != "" {
							roles = roles[:len(roles)-2]
						}
						break
					}
				}
			}
		}
	}

	// Reformat joinDate
	joinDateDate, _ := joinDate.Parse()

	// Created at date
	createdAt, err := discordgo.SnowflakeTimestamp(user.ID)

	// Status
	presence, err := s.State.Presence(m.GuildID, user.ID)
	status := "Offline"
	if err == nil {
		status = strings.Title(string(presence.Status))
	}

	// Obtain osu! info
	osuUsername := "N/A"
	for _, player := range cache {
		if player.Discord.ID == user.ID && player.Osu.Username != "" {
			osuUsername = player.Osu.Username
			break
		}
	}
	// Fix any blanks
	if roles == "" {
		roles = "None"
	}
	if nickname == "" {
		nickname = "None"
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    user.Username + "#" + user.Discriminator,
			IconURL: user.AvatarURL(""),
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "ID",
				Value:  user.ID,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Nickname",
				Value:  nickname,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Account Created",
				Value:  createdAt.UTC().Format(time.RFC822Z),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Date Joined",
				Value:  joinDateDate.Format(time.RFC822Z),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Status",
				Value:  status,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Linked on osu! as",
				Value:  osuUsername,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:  "Roles",
				Value: roles,
			},
		},
	})
}

// ServerInfo gives information about the server
func ServerInfo(s *discordgo.Session, m *discordgo.MessageCreate) {

	server, err := s.Guild(m.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "This is not a server!")
		return
	}

	// Stats info
	serverData := tools.GetServer(*server)
	statsInfo := strconv.Itoa(len(serverData.Nouns)) + " nouns\n" + strconv.Itoa(len(serverData.Adjectives)) + " adjectives\n" + strconv.Itoa(len(serverData.Skills)) + " skills\nAllowAnyoneAdd: " + strconv.FormatBool(serverData.AllowAnyoneStats)

	// Created at date
	createdAt, err := discordgo.SnowflakeTimestamp(server.ID)

	// Server owner
	owner, _ := s.User(server.OwnerID)

	// Channel info
	channels := server.Channels
	text := 0
	voice := 0
	category := 0
	other := 0
	for _, ch := range channels {
		switch ch.Type {
		case discordgo.ChannelTypeGuildText:
			text++
		case discordgo.ChannelTypeGuildVoice:
			voice++
		case discordgo.ChannelTypeGuildCategory:
			category++
		default:
			other++
		}
	}
	channelInfo := ""
	if text != 0 {
		channelInfo += strconv.Itoa(text) + " text\n"
	}
	if voice != 0 {
		channelInfo += strconv.Itoa(voice) + " voice\n"
	}
	if category != 0 {
		if category == 1 {
			channelInfo += strconv.Itoa(category) + " category\n"
		} else {
			channelInfo += strconv.Itoa(category) + " categories\n"
		}
	}
	if other != 0 {
		channelInfo += strconv.Itoa(other) + " other\n"
	}

	// AFK Timeout
	timeout := server.AfkTimeout
	timeoutString := "N/A"
	if timeout == int(time.Hour.Seconds()) {
		timeoutString = "1 hour"
	} else {
		timeoutString = strconv.Itoa(timeout/60) + " min"
	}

	// AFK Channel
	afkCh, err := s.Channel(server.AfkChannelID)
	afkChName := "N/A"
	if err == nil {
		afkChName = afkCh.Name
	}

	serverImg := "https://cdn.discordapp.com/icons/" + server.ID + "/" + server.Icon
	if strings.Contains(server.Icon, "a_") {
		serverImg += ".gif"
	} else {
		serverImg += ".png"
	}

	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    server.Name,
			IconURL: serverImg,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: serverImg,
		},
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "ID",
				Value:  server.ID,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Region",
				Value:  server.Region,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Server Created",
				Value:  createdAt.UTC().Format(time.RFC822Z),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Server Owner",
				Value:  owner.Username + "#" + owner.Discriminator,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "AFK Timeout",
				Value:  timeoutString,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "AFK Channel",
				Value:  afkChName,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Emoji Count",
				Value:  strconv.Itoa(len(server.Emojis)),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Member Count",
				Value:  strconv.Itoa(len(server.Members)),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Role Count",
				Value:  strconv.Itoa(len(server.Roles)),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Channels (" + strconv.Itoa(len(server.Channels)) + ")",
				Value:  channelInfo,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "`" + serverData.Prefix + "stats` Information:",
				Value:  statsInfo,
			},
		},
	})

	// Save new server data
	serverData.Time = time.Now()
	jsonCache, err := json.Marshal(serverData)
	tools.ErrRead(err)

	err = ioutil.WriteFile("./data/serverData/"+m.GuildID+".json", jsonCache, 0644)
	tools.ErrRead(err)
}
