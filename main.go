package main

import (
	"log"
)

var (
	faceDetector *FaceDetector
)

func main(){

	fd, err := Setup()
	if err != nil {
		log.Fatal("Couldn't Setup faces:", err.Error())
		return
	}
	faceDetector = fd

	go RunDiscordBot()
	RunTelegramBot()
}