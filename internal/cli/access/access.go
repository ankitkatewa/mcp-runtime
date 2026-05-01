// Package access owns routing for the access top-level command.
package access

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the access command.
func New(logger *zap.Logger) *cobra.Command {
	return cli.NewAccessCmd(logger)
}
