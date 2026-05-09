package cmd

import (
	"fmt"

	"charm.land/log/v2"
	"github.com/Gaurav-Gosain/cider/internal/server"
	"github.com/Gaurav-Gosain/cider/pkg/fm"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	var (
		host         string
		port         int
		instructions string
		apiKey       string
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start an OpenAI-compatible API server",
		Long:  "Starts an HTTP server that exposes Apple's on-device Foundation Model via an OpenAI-compatible API.",
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

			log.Info("Foundation Model is available")

			cfg := server.Config{
				Host:         host,
				Port:         port,
				Instructions: instructions,
				APIKey:       apiKey,
			}

			return server.Run(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "Host to bind to")
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	cmd.Flags().StringVarP(&instructions, "instructions", "i", "", "System instructions for the model")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Optional API key for authentication")

	return cmd
}
