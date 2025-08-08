package main

import (
	"flag"
	"log"

	"github.com/codecrafters-io/http-server-starter-go/config"
	"github.com/codecrafters-io/http-server-starter-go/server"
)

func main() {
	cfg := parseFlags()

	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatal("Failed to setup server:", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatal("Server error:", err)
	}
}

func parseFlags() *config.Config {
	cfg := config.DefaultConfig()

	flag.StringVar(&cfg.FileDir, "directory", cfg.FileDir, "specifies the directory where the files are stored, as an absolute path.")
	flag.Parse()

	return cfg
}
