package main

import (
	"go-discord-bot/bot" //we will create this later
	//we will create this later
)

func main() {
	bot.Start()

	<-make(chan struct{})
	return
}
