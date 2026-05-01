// Package server owns routing for the server top-level command.
package server

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the server command.
func New(logger *zap.Logger) *cobra.Command {
	return cli.NewServerCmd(logger)
}
