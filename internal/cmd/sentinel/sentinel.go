// Package sentinel owns routing for the sentinel top-level command.
package sentinel

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the sentinel command.
func New(logger *zap.Logger) *cobra.Command {
	return cli.NewSentinelCmd(logger)
}
