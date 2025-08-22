package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/n-korel/social-api/internal/env"
	"github.com/n-korel/social-api/internal/store"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}
	
	cfg := config{
		addr: ":" + env.GetString("PORT", "8080"),
	}

	store := store.NewStorage(nil)

	app := &application{
		config: cfg,
		store: store,
	}


	mux := app.mount()

	log.Fatal((app.run(mux)))
}