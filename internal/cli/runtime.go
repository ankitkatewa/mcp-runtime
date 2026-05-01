package cli

import "go.uber.org/zap"

// Runtime is the shared CLI facade for wiring common dependencies once and
// handing typed managers to the foldered command packages.
type Runtime struct {
	logger *zap.Logger
}

// NewRuntime builds the shared CLI runtime facade.
func NewRuntime(logger *zap.Logger) *Runtime {
	return &Runtime{logger: logger}
}

// Logger returns the shared logger.
func (r *Runtime) Logger() *zap.Logger {
	return r.logger
}

// KubectlRunner returns the shared kubectl runner.
func (r *Runtime) KubectlRunner() KubectlRunner {
	return DefaultKubectlRunner()
}

// AccessManager returns the access command manager.
func (r *Runtime) AccessManager() *AccessManager {
	return DefaultAccessManager(r.logger)
}

// AuthManager returns the auth command manager.
func (r *Runtime) AuthManager() *AuthManager {
	return NewAuthManager(r.logger)
}

// ClusterManager returns the cluster command manager.
func (r *Runtime) ClusterManager() *ClusterManager {
	return DefaultClusterManager(r.logger)
}

// PipelineManager returns the pipeline command manager.
func (r *Runtime) PipelineManager() *PipelineManager {
	return DefaultPipelineManager(r.logger)
}

// RegistryManager returns the registry command manager.
func (r *Runtime) RegistryManager() *RegistryManager {
	return DefaultRegistryManager(r.logger)
}

// SentinelManager returns the sentinel command manager.
func (r *Runtime) SentinelManager() *SentinelManager {
	return DefaultSentinelManager(r.logger)
}

// ServerManager returns the server command manager.
func (r *Runtime) ServerManager() *ServerManager {
	return DefaultServerManager(r.logger)
}
