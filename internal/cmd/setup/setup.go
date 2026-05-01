// Package setup owns routing for the setup top-level command.
package setup

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the setup command.
func New(logger *zap.Logger) *cobra.Command {
	return cli.NewSetupCmd(logger)
}
