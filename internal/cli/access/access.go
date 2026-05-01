// Package access owns routing for the access top-level command.
package access

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
)

const (
	grantResource   = "mcpaccessgrant"
	sessionResource = "mcpagentsession"
)

// New returns the access command.
func New(logger *zap.Logger) *cobra.Command {
	return NewWithManager(cli.DefaultAccessManager(logger))
}

// NewWithManager returns the access command using the provided manager.
func NewWithManager(mgr *cli.AccessManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access",
		Short: "Manage grants and agent sessions",
		Long: `Commands for managing MCPAccessGrant and MCPAgentSession resources that feed the gateway policy layer.

With mcp-runtime auth login, commands use the platform API by default. Use --use-kube
to target the cluster with kubectl and a kubeconfig (cluster admin path).`,
	}

	mgr.BindUseKubeFlag(cmd)

	cmd.AddCommand(newGrantCmd(mgr), newSessionCmd(mgr))
	return cmd
}

func newGrantCmd(mgr *cli.AccessManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant",
		Short: "Manage MCPAccessGrant resources",
	}
	cmd.AddCommand(newListCmd(mgr, grantResource, "grants"))
	cmd.AddCommand(newGetCmd(mgr, grantResource, "grant"))
	cmd.AddCommand(newApplyCmd(mgr, "grant"))
	cmd.AddCommand(newDeleteCmd(mgr, grantResource, "grant"))
	cmd.AddCommand(newToggleCmd(mgr, grantResource, "disable", "Disable a grant", true))
	cmd.AddCommand(newToggleCmd(mgr, grantResource, "enable", "Enable a grant", false))
	return cmd
}

func newSessionCmd(mgr *cli.AccessManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage MCPAgentSession resources",
	}
	cmd.AddCommand(newListCmd(mgr, sessionResource, "sessions"))
	cmd.AddCommand(newGetCmd(mgr, sessionResource, "session"))
	cmd.AddCommand(newApplyCmd(mgr, "session"))
	cmd.AddCommand(newDeleteCmd(mgr, sessionResource, "session"))
	cmd.AddCommand(newToggleCmd(mgr, sessionResource, "revoke", "Revoke an agent session", true))
	cmd.AddCommand(newToggleCmd(mgr, sessionResource, "unrevoke", "Clear the revoked flag on an agent session", false))
	return cmd
}

func newListCmd(mgr *cli.AccessManager, resource, label string) *cobra.Command {
	var namespace string
	var allNamespaces bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List access " + label,
		RunE: func(cmd *cobra.Command, args []string) error {
			return mgr.ListAccessResources(resource, namespace, allNamespaces)
		},
	}
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace to inspect")
	cmd.Flags().BoolVar(&allNamespaces, "all-namespaces", true, "List resources across all namespaces when no namespace is specified")
	return cmd
}

func newGetCmd(mgr *cli.AccessManager, resource, label string) *cobra.Command {
	var namespace string
	cmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Get an access " + label,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return mgr.GetAccessResource(resource, args[0], namespace)
		},
	}
	cmd.Flags().StringVar(&namespace, "namespace", cli.NamespaceMCPServers, "Namespace")
	return cmd
}

func newApplyCmd(mgr *cli.AccessManager, label string) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a " + label + " manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mgr.ApplyAccessResource(file)
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "Manifest file to apply")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func newDeleteCmd(mgr *cli.AccessManager, resource, label string) *cobra.Command {
	var namespace string
	cmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete an access " + label,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return mgr.DeleteAccessResource(resource, args[0], namespace)
		},
	}
	cmd.Flags().StringVar(&namespace, "namespace", cli.NamespaceMCPServers, "Namespace")
	return cmd
}

func newToggleCmd(mgr *cli.AccessManager, resource, use, short string, value bool) *cobra.Command {
	var namespace string
	cmd := &cobra.Command{
		Use:   use + " [name]",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return mgr.ToggleAccessResource(resource, args[0], namespace, value)
		},
	}
	cmd.Flags().StringVar(&namespace, "namespace", cli.NamespaceMCPServers, "Namespace")
	return cmd
}
