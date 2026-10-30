package main

import (
	"chessbox-notification/internal/application"
	"log"
)

func main() {
	app, err := application.New()

	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}