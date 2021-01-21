package gencommands

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"maquiaBot/tools"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// TTSError holds information for errors returned by 15.ai
type TTSError struct {
	Message string `json:"message"`
}

// TTS lets you create TTS using a given character from https://15.ai/
func TTS(s *discordgo.Session, m *discordgo.MessageCreate) {
	ttsRegex, _ := regexp.Compile(`(?is)tts\s+(.+)`)
	numberRegex, _ := regexp.Compile(`(?is)\d`)
	text := ""
	voice := "The Narrator"
	voiceChosen := false

	// Check if text was even sent, use last message not sent from bot otherwise
	if !ttsRegex.MatchString(m.Content) {
		msgs, err := s.ChannelMessages(m.ChannelID, -1, m.ID, "", "")
		if err == nil {
			for _, msg := range msgs {
				if msg.Author.ID != s.State.User.ID {
					text = msg.Content
					if ttsRegex.MatchString(text) {
						text = ttsRegex.FindStringSubmatch(text)[1]
					} else if strings.HasSuffix(strings.ToLower(text), "tts") {
						continue
					}
					if !voiceChosen && strings.Contains(text, "-v") {
						split := strings.Split(text, "-v")
						text = strings.TrimSpace(split[0])
						voice = strings.Title(strings.TrimSpace(split[1]))
					}
					break
				}
			}
		}
	} else {
		// Check voice
		text = ttsRegex.FindStringSubmatch(m.Content)[1]
		if strings.Contains(text, "-v") {
			split := strings.Split(text, "-v")
			text = strings.TrimSpace(split[0])
			voice = strings.Title(strings.TrimSpace(split[1]))

			if voice == "Glados" {
				voice = "GLaDOS"
			} else if voice == "Spongebob Squarepants" {
				voice = "SpongeBob SquarePants"
			}

			voiceChosen = true
		}

		// Check if anything aside for voice was sent
		if text == "" {
			msgs, err := s.ChannelMessages(m.ChannelID, -1, m.ID, "", "")
			if err == nil {
				for _, msg := range msgs {
					if msg.Author.ID != s.State.User.ID {
						text = msg.Content
						if ttsRegex.MatchString(text) {
							text = ttsRegex.FindStringSubmatch(text)[1]
						} else if strings.HasSuffix(strings.ToLower(text), "tts") {
							continue
						}
						if !voiceChosen && strings.Contains(text, "-v") {
							split := strings.Split(text, "-v")
							text = strings.TrimSpace(split[0])
							voice = strings.Title(strings.TrimSpace(split[1]))
						}
						break
					}
				}
			}
		}
	}

	// Remove quotations, remove new lines, and change numbers to words
	text = strings.Replace(text, "\"", "", -1)
	text = strings.Replace(text, "\r", " ", -1)
	text = strings.Replace(text, "\n", " ", -1)
	words := strings.Split(text, " ")
	for j, word := range words {
		if i, err := strconv.Atoi(word); err == nil {
			words[j] = tools.Ntow(int64(i))
		}
	}
	text = strings.Join(words, " ")

	if numberRegex.MatchString(text) {
		s.ChannelMessageSend(m.ChannelID, "https://15.ai/ does not input numbers! If you wish to use numbers, please seperate them from other words/letters with a space! The highest number possible is `9223372036854775807`")
		return
	}

	// Can't send more than 300 characters of text
	if len(text) > 300 {
		s.ChannelMessageSend(m.ChannelID, "Please keep the text under 300 characters!")
		return
	}
	minLength := int(math.Min(32, float64(len(text))))

	// Name checks for ones that aren't just title cases
	if voice == "Glados" {
		voice = "GLaDOS"
	} else if voice == "Spongebob Squarepants" || voice == "Spongebob" || voice == "Squarepants" {
		voice = "SpongeBob SquarePants"
	}

	// Create JSON payload
	jsonText := `{"text": "` + text + `", "character": "` + voice + `", "emotion": "Contextual", "use_diagonal": true}`
	data := []byte(jsonText)

	msg, err := s.ChannelMessageSend(m.ChannelID, "Sending text to <https://15.ai/> ...")
	if err != nil {
		return
	}

	// Send JSON payload
	req, err := http.NewRequest("POST", "https://api.15.ai/app/getAudioFile", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(req)

	go s.ChannelMessageDelete(m.ChannelID, msg.ID)

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "An error occured. The bot may be having trouble accessing te website, but make sure you are using a voice that currently exists on the website, or valid ARPAbet text https://15.ai/")
		return
	}
	defer res.Body.Close()

	// Remove the percentage information from the end of the text
	b, _ := ioutil.ReadAll(res.Body)
	if len(b) < 1 || b[0] == []byte("{")[0] {
		var ttsErr TTSError
		err = json.Unmarshal(b, &ttsErr)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error in obtaining the error message. The website's server may be down, but please also make sure you are using a voice that currently exists on the website, or valid ARPAbet text https://15.ai/")
		} else {
			s.ChannelMessageSend(m.ChannelID, "Error from https://15.ai/: "+ttsErr.Message+"\nText sent: "+text)
		}
		return
	}
	for {
		b = b[:len(b)-1]
		blen := len(b)
		if b[blen-1] == []byte("[")[0] {
			b = b[:blen-1]
			break
		}
	}

	// Send audio file
	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: "Here is the audio generated by <https://15.ai/>",
		File: &discordgo.File{
			Name:   text[:minLength] + ".wav",
			Reader: bytes.NewReader(b),
		},
	})
	return
}