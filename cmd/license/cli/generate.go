package cli

import (
	"context"
	"fmt"
	"time"

	"stitchforce-server/config"
	licenseRepo "stitchforce-server/internal/repo/postgres/license"
	licenseUseCase "stitchforce-server/internal/usecase/license"
	"stitchforce-server/pkg/logger"
	"stitchforce-server/pkg/postgres"

	"github.com/spf13/cobra"
)

func newGenerateCommand(
	cfg config.Config,
	log *logger.Logger,
) *cobra.Command {
	var (
		organizationID string
		durationDays   int
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a license for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if organizationID == "" {
				return fmt.Errorf("--org is required")
			}

			if durationDays < 1 || durationDays > 3650 {
				return fmt.Errorf("duration must be between 1 and 3650 days")
			}

			ctx, cancel := context.WithTimeout(
				cmd.Context(),
				10*time.Second,
			)
			defer cancel()

			db, err := postgres.New(cfg.Postgres)
			if err != nil {
				return fmt.Errorf("postgres connection failed: %w", err)
			}
			defer db.Close()

			if err := db.DB.PingContext(ctx); err != nil {
				return fmt.Errorf("postgres ping failed: %w", err)
			}

			repository := licenseRepo.New(db.DB)
			useCase := licenseUseCase.New(repository, log)

			generated, err := useCase.Generate(
				ctx,
				organizationID,
				durationDays,
			)
			if err != nil {
				log.Error("license generation failed", "error", err)
				return err
			}

			fmt.Println()
			fmt.Println("License generated successfully!")
			fmt.Println("----------------------------------")
			fmt.Println("Organization:", organizationID)
			fmt.Println("Duration:", durationDays, "days")
			fmt.Println()
			fmt.Println("License key:")
			fmt.Println(generated.LicenseKey)
			fmt.Println()
			fmt.Println("Save this key. It cannot be retrieved later.")
			fmt.Println("----------------------------------")

			return nil
		},
	}

	cmd.Flags().StringVar(
		&organizationID,
		"org",
		"",
		"Organization UUID",
	)

	cmd.Flags().IntVar(
		&durationDays,
		"days",
		365,
		"License duration in days",
	)

	_ = cmd.MarkFlagRequired("org")

	return cmd
}
