package cmd

import (
	"fmt"

	"charm.land/log/v2"
	"github.com/Gaurav-Gosain/cider/internal/tui"
	"github.com/Gaurav-Gosain/cider/pkg/fm"
	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	var (
		verbose      bool
		instructions string
	)

	root := &cobra.Command{
		Use:   "cider",
		Short: "Apple Foundation Models toolkit",
		Long:  "Cider exposes Apple's on-device Foundation Models via Go bindings and an OpenAI-compatible API server.\n\nRun without a subcommand to start the interactive chat TUI.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := fm.Init(); err != nil {
				return fmt.Errorf("failed to load Foundation Models library: %w", err)
			}

			model := fm.DefaultModel()
			defer model.Close()

			available, reason := model.IsAvailable()
			if !available {
				return fmt.Errorf("model unavailable: %s", reason)
			}

			sessionOpts := []fm.SessionOption{}
			if instructions != "" {
				sessionOpts = append(sessionOpts, fm.WithInstructions(instructions))
			}

			session, err := fm.NewSession(sessionOpts...)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}
			defer session.Close()

			return tui.Run(session, instructions)
		},
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose (debug) logging")
	root.Flags().StringVarP(&instructions, "instructions", "i", "", "System instructions for the model")
	root.AddCommand(serveCmd())

	return root
}
