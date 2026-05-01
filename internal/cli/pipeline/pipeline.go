// Package pipeline owns routing for the pipeline top-level command.
package pipeline

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

// New returns the pipeline command.
func New(logger *zap.Logger) *cobra.Command {
	return cli.NewPipelineCmd(logger)
}
