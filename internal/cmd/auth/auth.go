// Package auth owns routing for the auth top-level command.
package auth

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the auth command.
func New(logger *zap.Logger) *cobra.Command {
	return cli.NewAuthCmd(logger)
}
