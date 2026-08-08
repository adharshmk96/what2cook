package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"what2cook-api/internal/config"
	"what2cook-api/internal/db"
	"what2cook-api/internal/mail"
	"what2cook-api/internal/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()

		gdb, err := db.Open(cfg.DB.Path)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		if err := db.AutoMigrate(gdb); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}

		mailer := mail.New(cfg.SMTP, cfg.Server.PublicURL)

		srv, err := server.New(cfg, gdb, mailer)
		if err != nil {
			return fmt.Errorf("create server: %w", err)
		}

		log.Printf("listening on %s", cfg.Server.Addr)
		return srv.Run(cfg.Server.Addr)
	},
}
