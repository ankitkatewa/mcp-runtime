// Package cluster owns routing for the cluster top-level command.
package cluster

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the cluster command.
func New(logger *zap.Logger) *cobra.Command {
	return cli.NewClusterCmd(logger)
}
