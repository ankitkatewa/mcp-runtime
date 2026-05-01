// Package status owns routing for the status top-level command.
package status

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the status command.
func New(logger *zap.Logger) *cobra.Command {
	return cli.NewStatusCmd(logger)
}
