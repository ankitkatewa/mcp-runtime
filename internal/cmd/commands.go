package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cmd/access"
	"mcp-runtime/internal/cmd/auth"
	"mcp-runtime/internal/cmd/bootstrap"
	"mcp-runtime/internal/cmd/cluster"
	"mcp-runtime/internal/cmd/pipeline"
	"mcp-runtime/internal/cmd/registry"
	"mcp-runtime/internal/cmd/sentinel"
	"mcp-runtime/internal/cmd/server"
	"mcp-runtime/internal/cmd/setup"
	"mcp-runtime/internal/cmd/status"
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
