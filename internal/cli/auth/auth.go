// Package auth owns routing for the auth top-level command.
package auth

import (
	"github.com/spf13/cobra"

	"mcp-runtime/internal/cli"
	"mcp-runtime/pkg/authfile"
)

// New returns the auth command.
func New(runtime *cli.Runtime) *cobra.Command {
	m := runtime.AuthManager()
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Log in to the platform API and manage saved credentials",
		Long: `Authenticate to the Sentinel platform using email/password or an API token (not Kubernetes).

Use this for day-to-day deploy and registry-related flows. Cluster install and admin work
use Kubernetes and the cluster commands, not this command.

The token is stored in a local file (mode 0600) under the user config directory, unless you set ` + authfile.EnvAPIToken + `.

Optional environment:
  ` + authfile.EnvAPIURL + `      default API base for login, e.g. https://platform.example.com
  ` + authfile.EnvAPIToken + `    use this token for API calls; overrides a saved file
  MCP_RUNTIME_CONFIG_DIR    override the config directory (mainly for tests)`,
	}

	cmd.AddCommand(m.NewLoginCmd())
	cmd.AddCommand(m.NewLogoutCmd())
	cmd.AddCommand(m.NewStatusCmd())
	return cmd
}
