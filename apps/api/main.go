package main

import (
	"log"
	"net/http"

	"github.com/tinotenda-alfaneti/grocery-compare/internal/config"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/db"
	httpserver "github.com/tinotenda-alfaneti/grocery-compare/internal/http"
)

func main() {
	cfg := config.Load()

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer conn.Close()

	r := httpserver.NewRouter(conn, cfg.WebRoot)

	addr := ":" + cfg.HTTPPort
	log.Printf("grocery-compare API listening on %s (env=%s, db=%s)", addr, cfg.AppEnv, cfg.DBPath)
	log.Fatal(http.ListenAndServe(addr, r))
}
