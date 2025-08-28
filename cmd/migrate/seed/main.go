package main

import (
	"log"

	"github.com/n-korel/social-api/internal/db"
	"github.com/n-korel/social-api/internal/env"
	"github.com/n-korel/social-api/internal/store"
)

func main() {
	dsn := env.GetString("DSN", "host=localhost user=postgres password=my_pass dbname=social-api port=5432 sslmode=disable")
	connDb, err := db.New(dsn, 3, 3, "15m")

	if err != nil {
		log.Fatal(err)
	}

	defer connDb.Close()

	store := store.NewStorage(connDb)

	db.Seed(store)
}