package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"what2cook-api/internal/config"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "what2cook-api",
		Short: "what2cook API server",
	}
)

// Execute runs the root Cobra command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: config.yaml)")
	rootCmd.AddCommand(serveCmd)
}

func initConfig() {
	if err := config.Load(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
}
