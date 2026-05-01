// Package status owns routing for the status top-level command.
package status

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the status command.
func New(logger *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show platform status",
		Long:  "Show the overall status of the MCP platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.ShowPlatformStatus(logger)
		},
	}
}
