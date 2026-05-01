// Package bootstrap owns routing for the bootstrap top-level command.
package bootstrap

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the bootstrap command.
func New(logger *zap.Logger) *cobra.Command {
	return cli.NewBootstrapCmd(logger)
}
