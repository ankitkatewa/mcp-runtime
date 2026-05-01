// Package registry owns routing for the registry top-level command.
package registry

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the registry command.
func New(logger *zap.Logger) *cobra.Command {
	return cli.NewRegistryCmd(logger)
}
