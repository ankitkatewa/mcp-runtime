// Package bootstrap owns routing for the bootstrap top-level command.
package bootstrap

import (
	"fmt"

	"github.com/spf13/cobra"

	"mcp-runtime/internal/cli"
)

type manager struct {
	kubectl cli.KubectlRunner
}

func newManager(runtime *cli.Runtime) *manager {
	return &manager{kubectl: runtime.KubectlRunner()}
}

// New returns the bootstrap command.
func New(runtime *cli.Runtime) *cobra.Command {
	var apply bool
	var provider string
	mgr := newManager(runtime)

	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap cluster prerequisites (on-prem focused)",
		Long: `Bootstrap validates and (optionally) installs cluster prerequisites needed by mcp-runtime setup.

By design, this does not provision Kubernetes clusters end-to-end across all distributions.
Use this to prepare an existing cluster for running 'mcp-runtime setup'.

Note: bootstrap --apply is automated for k3s only and must be executed on the k3s server node (it expects local manifests under /var/lib/rancher/k3s/server/manifests).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli.Section("MCP Runtime Bootstrap")
			chosenProvider := provider
			if chosenProvider == "" || chosenProvider == "auto" {
				detectedProvider, err := cli.DetectProvider(mgr.kubectl)
				if err != nil {
					return err
				}
				chosenProvider = detectedProvider
			}
			cli.Info(fmt.Sprintf("Provider: %s", chosenProvider))

			if err := cli.RunBootstrapPreflight(mgr.kubectl); err != nil {
				return err
			}

			if !apply {
				cli.Success("Bootstrap preflight complete (no changes applied)")
				cli.Info("Next: run `./bin/mcp-runtime setup` (or `./bin/mcp-runtime setup --storage-mode hostpath` for single-node dev)")
				return nil
			}

			switch chosenProvider {
			case "k3s":
				if err := cli.BootstrapApplyK3s(mgr.kubectl); err != nil {
					return err
				}
			case "rke2", "kubeadm", "generic":
				cli.Warn("Apply mode is currently only automated for k3s. For other distributions, use the preflight output and install DNS/storage/ingress/load-balancer via your standard platform tooling.")
			default:
				cli.Warn(fmt.Sprintf("Unknown provider %q; skipping apply", chosenProvider))
			}

			cli.Success("Bootstrap complete")
			cli.Info("Next: run `./bin/mcp-runtime setup`")
			return nil
		},
	}

	cmd.Flags().BoolVar(&apply, "apply", false, "Apply safe bootstrap fixes when possible (k3s only today; run on the k3s server node)")
	cmd.Flags().StringVar(&provider, "provider", "auto", "Cluster provider hint (auto|k3s|rke2|kubeadm|generic)")
	return cmd
}
