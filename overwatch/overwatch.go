package overwatch

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"time"
)

type UserMessageStat struct {
	UserID                 string
	Username               string
	MessagesLastDay        uint64
	MessagesLastHour       uint64
	MessagesLastFiveMins   uint64
	MessagesLastThirtySecs uint64
	Kicks                  int
}

type Overwatch struct {
	TotalMessages uint64
	UserMessages  map[string]*UserMessageStat
}

func (o *Overwatch) ProcessMessage(s *discordgo.Session, m interface{}) {
	switch m.(type) {
	case *discordgo.MessageCreate:
		err := o.handleUserStat(s, m.(*discordgo.MessageCreate))
		if err != nil {
			log.Printf("[!] Error handling Overwatch user stat: %s\n", err.Error())
		}
		break
	case *discordgo.GuildMemberAdd:
		break
	}
}

func (o *Overwatch) handleUserStat(s *discordgo.Session, m *discordgo.MessageCreate) error {
	userID := m.Author.ID
	user, ok := o.UserMessages[userID]
	if !ok {
		o.UserMessages[userID] = &UserMessageStat{
			UserID:   userID,
			Username: m.Author.Username,
		}
		user = o.UserMessages[userID]
	}

	user.MessagesLastDay++
	user.MessagesLastHour++
	user.MessagesLastFiveMins++
	user.MessagesLastThirtySecs++

	return nil
}

// this is fucking amazing code
func (o *Overwatch) Run() {
	// State of the art anti-spam loop
	go func() {
		for range time.Tick(5 * time.Second) {
			for _, user := range o.UserMessages {
				// load the threshold from the config file, dipshit
				if user.MessagesLastThirtySecs > 10 {
					// Set slow mode, kick user? add kick count?
					if user.Kicks > 2 {
						// ban that sucker
						log.Printf("[*] User %s (%s) was banned due to previous spam-related kicks", user.Username, user.UserID)
					} else {
						user.Kicks++
						// kick that sucker
						log.Printf("[*] User %s (%s) has triggered the message threshold.", user.Username, user.UserID)
					}
				}
			}
		}
	}()

	// Clear Counters
	go func() {
		for range time.Tick(24 * time.Hour) {
			for _, user := range o.UserMessages {
				if user.MessagesLastDay == 0 {
					delete(o.UserMessages, user.UserID)
				} else {
					user.MessagesLastDay = 0
				}
			}
		}
		for range time.Tick(1 * time.Hour) {
			for _, user := range o.UserMessages {
				user.MessagesLastHour = 0
			}
		}
		for range time.Tick(5 * time.Minute) {
			for _, user := range o.UserMessages {
				user.MessagesLastFiveMins = 0
			}
		}
		for range time.Tick(30 * time.Second) {
			for _, user := range o.UserMessages {
				user.MessagesLastThirtySecs = 0
			}
		}
	}()
}
