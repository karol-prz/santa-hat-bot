package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	responseMap = map[string]string{
		"ping":   "pong",
		"pong":   "ping",
		"/hello": "Hello World!",
		"/start": "Send my an image or a link to one and I'll put Xmas hats on it.",
		"/help":  "Hi, I do my best to photoshop Xmas hats onto any picture or link to a picture that you send."+
					"\n\nUse the command /xmas with a photo attached or a link to a photo."+
					"\nOr reply to an photo or link with /xmas!",
	}
	bot *tb.Bot
)

// RunTelegramBot starts the bot with handler
func RunTelegramBot() {
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	bot = b

	if err != nil {
		log.Fatal(err)
		return
	}

	bot.Handle(tb.OnText, mappedResponse)
	bot.Handle("/stop", func(m *tb.Message) {
		if m.Sender.ID != 356006294 {
			return
		}
		bot.Reply(m, "Guess I'll stop now.")
		fmt.Println("Telegram Bot is now stopping.")
		bot.Stop()
	})
	bot.Handle("/xmas", processXmas)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Telegram Bot is now running.  Press CTRL-C to exit.")
	// graceful shutdown
	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc
		bot.Stop()
	}()

	bot.Start()
}

func mappedResponse(m *tb.Message) {
	if text, ok := responseMap[m.Text]; ok {
		bot.Reply(m, text)
	}
}

func processXmas(m *tb.Message){
	var url string
	var err error
	if m.IsReply() && m.ReplyTo.Photo != nil {
		url, err = bot.FileURLByID(m.ReplyTo.Photo.FileID)
		if err != nil{
			log.Print(err.Error())
			return
		}
	} else if m.IsReply() && IsImageExt(m.ReplyTo.Text){
		url = m.ReplyTo.Text
	} else if m.Photo != nil {
		url, err = bot.FileURLByID(m.ReplyTo.Photo.FileID)
		if err != nil{
			log.Print(err.Error())
			return
		}
	} else if IsImageExt(m.Payload){
		url = m.Payload
	} else {
		bot.Reply(m, "Can't find image there. Try replying to an image or an image link.")
		return
	}
	XMasHatsOn(url, m)
}

// XMasHatsOn puts xmas hats on everyone in the pic and sends it back
func XMasHatsOn(url string, m *tb.Message) {
	input, err := GetDataFromURL(url)
	if err != nil {
		log.Println("Error downloading file")
		return
	}
	ext := GetImageExt(url)
	output, err := XMassify(input, faceDetector, ext)
	if err != nil {
		log.Println(err.Error())
		return
	}
	filereader := &tb.Photo{File: tb.FromReader(bytes.NewReader(output))}
	bot.Reply(m, filereader)
}