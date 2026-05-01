// Package status owns routing for the status top-level command.
package status

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

type manager struct {
	logger *zap.Logger
}

func newManager(runtime *cli.Runtime) *manager {
	return &manager{logger: runtime.Logger()}
}

// New returns the status command.
func New(runtime *cli.Runtime) *cobra.Command {
	mgr := newManager(runtime)
	return &cobra.Command{
		Use:   "status",
		Short: "Show platform status",
		Long:  "Show the overall status of the MCP platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.ShowPlatformStatus(mgr.logger)
		},
	}
}
