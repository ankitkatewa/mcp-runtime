package cluster

import (
	"github.com/spf13/cobra"

	"mcp-runtime/internal/cli"
)

func newClusterDoctorCmd(mgr *cli.ClusterManager) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose MCP Runtime cluster readiness and installed components",
		Long: "Detect the Kubernetes distribution and check that the registry service, cluster DNS, " +
			"operator/CRD prerequisites, ingress (Traefik) wiring, image pulls, Sentinel, and MCPServer reconciliation are healthy. Prints remediation steps for your distribution " +
			"when something is missing. See docs/cluster-readiness.md for the full per-distribution checklist.",
		RunE: func(cmd *cobra.Command, args []string) error {
			report := cli.RunDoctorAndPrint(mgr.KubectlRunner())
			if !report.AllOK() {
				return cli.NewSetupStepFailedError()
			}
			return nil
		},
	}
}
