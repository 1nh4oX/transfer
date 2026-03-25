package main

import (
	"log"

	"transfer/backend/internal/app"
)

func main() {
	server, err := app.NewServer()
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
