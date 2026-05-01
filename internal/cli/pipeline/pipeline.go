// Package pipeline owns routing for the pipeline top-level command.
package pipeline

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"mcp-runtime/internal/cli"
	"mcp-runtime/pkg/metadata"
)

// filepathGlob is a test seam for filepath.Glob.
var filepathGlob = filepath.Glob

type manager struct {
	kubectl *cli.KubectlClient
	logger  *zap.Logger
}

func newManager(runtime *cli.Runtime) *manager {
	return &manager{
		kubectl: runtime.KubectlClient(),
		logger:  runtime.Logger(),
	}
}

// New returns the pipeline command.
func New(runtime *cli.Runtime) *cobra.Command {
	return NewWithManager(newManager(runtime))
}

// NewWithManager returns the pipeline command using the provided manager.
func NewWithManager(mgr *manager) *cobra.Command {
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

func (m *manager) GenerateCRDsFromMetadata(metadataFile, metadataDir, outputDir string) error {
	var registry *metadata.RegistryFile
	var err error

	if metadataFile != "" {
		m.logger.Info("Loading metadata from file", zap.String("file", metadataFile))
		registry, err = metadata.LoadFromFile(metadataFile)
	} else {
		m.logger.Info("Loading metadata from directory", zap.String("dir", metadataDir))
		registry, err = metadata.LoadFromDirectory(metadataDir)
	}

	if err != nil {
		wrappedErr := cli.WrapWithSentinel(cli.ErrLoadMetadataFailed, err, fmt.Sprintf("failed to load metadata: %v", err))
		cli.Error("Failed to load metadata")
		cli.LogStructuredError(m.logger, wrappedErr, "Failed to load metadata")
		return wrappedErr
	}

	if len(registry.Servers) == 0 {
		err := cli.ErrNoServersInMetadata
		cli.Error("No servers found in metadata")
		cli.LogStructuredError(m.logger, err, "No servers found in metadata")
		return err
	}

	if metadata.ResolveRegistryHost() == metadata.DefaultRegistryHost {
		m.logger.Warn("Using default image host registry.local for generated MCPServer image refs. If cluster pulls fail, set MCP_REGISTRY_INGRESS_HOST to your registry (e.g. ClusterIP:port) and configure containerd/k3s for HTTP, or use public DNS and TLS.")
	}

	m.logger.Info("Generating CRD files", zap.Int("count", len(registry.Servers)), zap.String("output", outputDir))

	if err := metadata.GenerateCRDsFromRegistry(registry, outputDir); err != nil {
		wrappedErr := cli.WrapWithSentinelAndContext(
			cli.ErrGenerateCRDsFailed,
			err,
			fmt.Sprintf("failed to generate CRDs: %v", err),
			map[string]any{"output_dir": outputDir, "server_count": len(registry.Servers), "component": "pipeline"},
		)
		cli.Error("Failed to generate CRDs")
		cli.LogStructuredError(m.logger, wrappedErr, "Failed to generate CRDs")
		return wrappedErr
	}

	m.logger.Info("CRD files generated successfully", zap.String("output", outputDir))

	files, _ := filepath.Glob(filepath.Join(outputDir, "*.yaml"))
	for _, file := range files {
		cli.Success(fmt.Sprintf("Generated: %s", file))
	}

	return nil
}

func (m *manager) DeployCRDs(manifestsDir, namespace string) error {
	if _, kerr := m.kubectl.CombinedOutput([]string{"version", "--request-timeout=5s"}); kerr != nil {
		if cli.HasPlatformClient() {
			return cli.NewWithSentinel(cli.ErrApplyManifestFailed, "pipeline deploy applies YAML with kubectl and needs a working kubeconfig. mcp-runtime auth is for the platform API only, not for applying manifests. Run deploy from a host with cluster access, or fix KUBECONFIG, then retry.")
		}
	}
	m.logger.Info("Deploying CRD files", zap.String("dir", manifestsDir))

	files, err := filepathGlob(filepath.Join(manifestsDir, "*.yaml"))
	if err != nil {
		wrappedErr := cli.WrapWithSentinelAndContext(
			cli.ErrListManifestFilesFailed,
			err,
			fmt.Sprintf("failed to list manifest files: %v", err),
			map[string]any{"manifest_dir": manifestsDir, "component": "pipeline"},
		)
		cli.Error("Failed to list manifest files")
		cli.LogStructuredError(m.logger, wrappedErr, "Failed to list manifest files")
		return wrappedErr
	}

	ymlFiles, err := filepathGlob(filepath.Join(manifestsDir, "*.yml"))
	if err != nil {
		wrappedErr := cli.WrapWithSentinelAndContext(
			cli.ErrListManifestFilesFailed,
			err,
			fmt.Sprintf("failed to list manifest files: %v", err),
			map[string]any{"manifest_dir": manifestsDir, "component": "pipeline"},
		)
		cli.Error("Failed to list manifest files")
		cli.LogStructuredError(m.logger, wrappedErr, "Failed to list manifest files")
		return wrappedErr
	}

	files = append(files, ymlFiles...)
	if len(files) == 0 {
		err := cli.NewWithSentinel(cli.ErrNoManifestFilesFound, fmt.Sprintf("no manifest files found in %s", manifestsDir))
		cli.Error("No manifest files found")
		cli.LogStructuredError(m.logger, err, "No manifest files found")
		return err
	}

	for _, file := range files {
		m.logger.Info("Applying manifest", zap.String("file", file))

		absPath, err := cli.ResolveRegularFilePath(file)
		if err != nil {
			wrappedErr := cli.WrapWithSentinelAndContext(
				cli.ErrApplyManifestFailed,
				err,
				fmt.Sprintf("failed to resolve %s: %v", file, err),
				map[string]any{"file": file, "namespace": namespace, "component": "pipeline"},
			)
			cli.Error("Failed to resolve manifest file")
			cli.LogStructuredError(m.logger, wrappedErr, "Failed to resolve manifest file")
			return wrappedErr
		}

		manifestBytes, err := cli.ReadFileAtPath(absPath)
		if err != nil {
			wrappedErr := cli.WrapWithSentinelAndContext(
				cli.ErrApplyManifestFailed,
				err,
				fmt.Sprintf("failed to read %s: %v", absPath, err),
				map[string]any{"file": file, "namespace": namespace, "component": "pipeline"},
			)
			cli.Error("Failed to read manifest file")
			cli.LogStructuredError(m.logger, wrappedErr, "Failed to read manifest file")
			return wrappedErr
		}

		if err := cli.ApplyManifestContentWithNamespace(m.kubectl, string(manifestBytes), namespace); err != nil {
			wrappedErr := cli.WrapWithSentinelAndContext(
				cli.ErrApplyManifestFailed,
				err,
				fmt.Sprintf("failed to apply %s: %v", file, err),
				map[string]any{"file": file, "namespace": namespace, "component": "pipeline"},
			)
			cli.Error("Failed to apply manifest")
			cli.LogStructuredError(m.logger, wrappedErr, "Failed to apply manifest")
			return wrappedErr
		}
	}

	m.logger.Info("All CRD files deployed successfully")
	return nil
}
