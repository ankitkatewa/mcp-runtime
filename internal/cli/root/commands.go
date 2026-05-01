package root

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli/access"
	"mcp-runtime/internal/cli/auth"
	"mcp-runtime/internal/cli/bootstrap"
	"mcp-runtime/internal/cli/cluster"
	"mcp-runtime/internal/cli/pipeline"
	"mcp-runtime/internal/cli/registry"
	"mcp-runtime/internal/cli/sentinel"
	"mcp-runtime/internal/cli/server"
	"mcp-runtime/internal/cli/setup"
	"mcp-runtime/internal/cli/status"
)

// AddCommands registers every top-level mcp-runtime command on root.
func AddCommands(root *cobra.Command, logger *zap.Logger) {
	root.AddCommand(cluster.New(logger))
	root.AddCommand(registry.New(logger))
	root.AddCommand(server.New(logger))
	root.AddCommand(access.New(logger))
	root.AddCommand(auth.New(logger))
	root.AddCommand(bootstrap.New(logger))
	root.AddCommand(setup.New(logger))
	root.AddCommand(status.New(logger))
	root.AddCommand(sentinel.New(logger))
	root.AddCommand(pipeline.New(logger))
}
