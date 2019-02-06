package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// RunDiscordBot starts the bot with all handlers
func RunDiscordBot() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + "NTE1MjM1MjgwNDM4MTY1NTE0.DtiJ5g.4e0YGO3o60On9PDuRCisP1gpdNU")
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Discord Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	if m.Content == "/stop" && m.Author.ID == adminID {
		s.ChannelMessageSend(m.ChannelID, "Guess I'll stop running.")
		fmt.Println("Discord Bot is now stopping.")
		s.Close()
	}

	if len(m.Attachments) > 0 {
		for _, attachment := range m.Attachments {
			if !IsImageExt(attachment.Filename) {
				continue
			}
			XMasHats(attachment.URL, s, m)
		}
	}
	if IsImageExt(m.Content) {
		XMasHats(GetURL(m.Content), s, m)
	}
}

// XMasHats puts xmas hats on everyone in the pic and sends it back
func XMasHats(url string, s *discordgo.Session, m *discordgo.MessageCreate) {
	ext := GetImageExt(url)
	input, err := GetDataFromURL(url)
	if err != nil {
		log.Print("Error downloading file")
		return
	}
	output, err := XMassify(input, faceDetector, ext)
	if err != nil {
		log.Print("Error Xmassifying image")
		return
	}
	s.ChannelFileSend(m.ChannelID, "MerryChristmas.png", bytes.NewReader(output))
}
