package main

import (
	"log"

	"github.com/briefcast/briefcast/internal/config"
	"github.com/briefcast/briefcast/internal/db"
	"github.com/briefcast/briefcast/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("loading config: ", err)
	}

	database, err := db.Open(cfg.DatabasePath, migrations.FS)
	if err != nil {
		log.Fatal("opening database: ", err)
	}
	defer database.Close()

	log.Println("worker starting")
	select {} // block forever for now
}
