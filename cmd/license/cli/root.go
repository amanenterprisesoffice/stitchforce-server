package cli

import (
	"stitchforce-server/config"
	"stitchforce-server/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	rootCmd *cobra.Command
	cfg     config.Config
	log     *logger.Logger
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cfg = config.Load()
	log = logger.New(cfg.Log.Level)

	rootCmd = &cobra.Command{
		Use:   "stitchforce",
		Short: "StitchForce license management CLI",
		Long:  "CLI for generating and managing StitchForce organization licenses.",
	}

	rootCmd.AddCommand(newGenerateCommand(cfg, log))
}