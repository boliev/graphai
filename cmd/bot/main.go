package main

import "github.com/boliev/graphai/internal/app/bot"

func main() {
	botApp := bot.New()

	err := botApp.Start()
	if err != nil {
		panic(err)
	}
}
