// Package pipeline owns routing for the pipeline top-level command.
package pipeline

import (
	"github.com/spf13/cobra"

	"mcp-runtime/internal/cli"
)

// New returns the pipeline command.
func New(runtime *cli.Runtime) *cobra.Command {
	return NewWithManager(runtime.PipelineManager())
}

// NewWithManager returns the pipeline command using the provided manager.
func NewWithManager(mgr *cli.PipelineManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Pipeline integration commands",
		Long:  "Commands for CI/CD pipeline integration to generate and deploy CRDs",
	}

	var metadataFile string
	var metadataDir string
	var outputDir string
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate CRD files from metadata",
		Long: `Generate Kubernetes CRD files from metadata/registry files.
This command reads server definitions and creates CRD YAML files that
the operator will use to deploy MCP servers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return mgr.GenerateCRDsFromMetadata(metadataFile, metadataDir, outputDir)
		},
	}
	generateCmd.Flags().StringVar(&metadataFile, "file", "", "Path to metadata file (YAML)")
	generateCmd.Flags().StringVar(&metadataDir, "dir", ".mcp", "Directory containing metadata files")
	generateCmd.Flags().StringVar(&outputDir, "output", "manifests", "Output directory for CRD files")

	var manifestsDir string
	var namespace string
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy CRD files to cluster",
		Long: `Deploy generated CRD files to the Kubernetes cluster.
This applies all CRD manifests to the cluster, which triggers
the operator to create the necessary Kubernetes resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return mgr.DeployCRDs(manifestsDir, namespace)
		},
	}
	deployCmd.Flags().StringVar(&manifestsDir, "dir", "manifests", "Directory containing CRD files")
	deployCmd.Flags().StringVar(&namespace, "namespace", "", "Namespace to deploy to (overrides metadata)")

	cmd.AddCommand(generateCmd, deployCmd)
	return cmd
}
