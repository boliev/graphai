package main

import "github.com/boliev/graphai/internal/app/vkapi"

func main() {
	app := vkapi.NewVKApi()
	app.Run()
}
