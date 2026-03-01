package main

import (
	"log"
	"net/http"

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

	log.Printf("server starting on :%s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
		log.Fatal("server error: ", err)
	}
}
