package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestCheckCertManagerInstalledWithKubectl(t *testing.T) {
	mock := &MockExecutor{}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := checkCertManagerInstalledWithKubectl(kubectl); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Commands) != 1 {
		t.Fatalf("expected 1 kubectl command, got %d", len(mock.Commands))
	}
	if !commandHasArgs(mock.Commands[0], "get", "crd", CertManagerCRDName) {
		t.Fatalf("unexpected args: %v", mock.Commands[0].Args)
	}
}

func TestCheckCertManagerInstalledWithKubectlError(t *testing.T) {
	mock := &MockExecutor{DefaultRunErr: errors.New("missing")}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := checkCertManagerInstalledWithKubectl(kubectl); !errors.Is(err, ErrCertManagerNotInstalled) {
		t.Fatalf("expected ErrCertManagerNotInstalled, got %v", err)
	}
}

func TestCheckCASecretWithKubectl(t *testing.T) {
	mock := &MockExecutor{}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := checkCASecretWithKubectl(kubectl); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Commands) != 1 {
		t.Fatalf("expected 1 kubectl command, got %d", len(mock.Commands))
	}
	if !commandHasArgs(mock.Commands[0], "get", "secret", certCASecretName, "-n", certManagerNamespace) {
		t.Fatalf("unexpected args: %v", mock.Commands[0].Args)
	}
}

func TestCheckCASecretWithKubectlError(t *testing.T) {
	mock := &MockExecutor{DefaultRunErr: errors.New("missing")}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := checkCASecretWithKubectl(kubectl); !errors.Is(err, ErrCASecretNotFound) {
		t.Fatalf("expected ErrCASecretNotFound, got %v", err)
	}
}

func TestApplyClusterIssuerWithKubectl(t *testing.T) {
	mock := &MockExecutor{}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := applyClusterIssuerWithKubectl(kubectl); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Commands) != 1 {
		t.Fatalf("expected 1 kubectl command, got %d", len(mock.Commands))
	}
	if !commandHasArgs(mock.Commands[0], "apply", "-f", clusterIssuerManifestPath) {
		t.Fatalf("unexpected args: %v", mock.Commands[0].Args)
	}
}

func TestApplyRegistryCertificateWithKubectl(t *testing.T) {
	orig := DefaultCLIConfig
	t.Cleanup(func() { DefaultCLIConfig = orig })
	DefaultCLIConfig = &CLIConfig{RegistryEndpoint: "10.43.39.164:5000", RegistryIngressHost: "registry.prod.example.com"}

	root := repoRootForTest(t)
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir to repo root: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	var applyCmd *MockCommand
	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "apply", "-f", "-") {
				applyCmd = cmd
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := applyRegistryCertificateWithKubectl(kubectl); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Commands) != 1 {
		t.Fatalf("expected 1 kubectl command, got %d", len(mock.Commands))
	}
	if applyCmd == nil {
		t.Fatal("expected apply command")
	}
	captured, err := io.ReadAll(applyCmd.StdinR)
	if err != nil {
		t.Fatalf("read apply stdin: %v", err)
	}
	if !strings.Contains(string(captured), "registry.prod.example.com") {
		t.Fatalf("expected rewritten registry host, got: %s", string(captured))
	}
}

func TestWaitForCertificateReadyWithKubectl(t *testing.T) {
	mock := &MockExecutor{}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	timeout := 15 * time.Second
	if err := waitForCertificateReadyWithKubectl(kubectl, registryCertificateName, NamespaceRegistry, timeout); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Commands) != 1 {
		t.Fatalf("expected 1 kubectl command, got %d", len(mock.Commands))
	}
	if !commandHasArgs(mock.Commands[0], "wait", "--for=condition=Ready", "certificate/"+registryCertificateName, "-n", NamespaceRegistry, "--timeout=15s") {
		t.Fatalf("unexpected args: %v", mock.Commands[0].Args)
	}
}

func TestCertManagerStatus(t *testing.T) {
	mock := &MockExecutor{}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	manager := NewCertManager(kubectl, zap.NewNop())

	if err := manager.Status(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Commands) != 4 {
		t.Fatalf("expected 4 kubectl commands, got %d", len(mock.Commands))
	}
}

func TestCertManagerStatusMissingCertificate(t *testing.T) {
	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "get", "certificate", registryCertificateName, "-n", NamespaceRegistry) {
				cmd.RunErr = errors.New("missing cert")
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	manager := NewCertManager(kubectl, zap.NewNop())

	if err := manager.Status(); err == nil {
		t.Fatal("expected error")
	}
}

func TestCertManagerApplyMissingCASecret(t *testing.T) {
	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "get", "secret", certCASecretName, "-n", certManagerNamespace) {
				cmd.RunErr = errors.New("missing secret")
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	manager := NewCertManager(kubectl, zap.NewNop())

	if err := manager.Apply(); err == nil {
		t.Fatal("expected error")
	}
}

func TestCertManagerApplyClusterIssuerError(t *testing.T) {
	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "apply", "-f", clusterIssuerManifestPath) {
				cmd.RunErr = errors.New("apply issuer failed")
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	manager := NewCertManager(kubectl, zap.NewNop())

	if err := manager.Apply(); err == nil {
		t.Fatal("expected error")
	}
}

func TestCertManagerApplyEnsureNamespaceError(t *testing.T) {
	origKubectl := kubectlClient
	t.Cleanup(func() { kubectlClient = origKubectl })

	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "apply", "-f", "-") {
				cmd.RunErr = errors.New("apply namespace failed")
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	kubectlClient = kubectl
	manager := NewCertManager(kubectl, zap.NewNop())

	if err := manager.Apply(); err == nil {
		t.Fatal("expected error")
	}
}

func TestCertManagerApplyRegistryCertificateError(t *testing.T) {
	origKubectl := kubectlClient
	t.Cleanup(func() { kubectlClient = origKubectl })

	// The registry certificate is applied via `kubectl apply -f - -n registry` with the
	// manifest content piped over stdin, so match on those args rather than on the
	// on-disk manifest path.
	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "apply", "-f", "-", "-n", NamespaceRegistry) {
				cmd.RunErr = errors.New("apply cert failed")
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	kubectlClient = kubectl
	manager := NewCertManager(kubectl, zap.NewNop())

	if err := manager.Apply(); err == nil {
		t.Fatal("expected error")
	}
}

func TestCertManagerWaitFailure(t *testing.T) {
	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "wait", "--for=condition=Ready", "certificate/"+registryCertificateName, "-n", NamespaceRegistry) {
				cmd.RunErr = errors.New("wait failed")
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	manager := NewCertManager(kubectl, zap.NewNop())

	if err := manager.Wait(time.Second); err == nil {
		t.Fatal("expected error")
	}
}

func TestCertManagerStatusMissingCertManager(t *testing.T) {
	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "get", "crd", CertManagerCRDName) {
				cmd.RunErr = errors.New("not found")
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	manager := NewCertManager(kubectl, zap.NewNop())

	var buf bytes.Buffer
	setDefaultPrinterWriter(t, &buf)

	if err := manager.Status(); err == nil {
		t.Fatal("expected error when cert-manager not installed")
	}
}

func TestCertManagerStatusMissingCASecret(t *testing.T) {
	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "get", "secret", certCASecretName, "-n", certManagerNamespace) {
				cmd.RunErr = errors.New("not found")
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	manager := NewCertManager(kubectl, zap.NewNop())

	var buf bytes.Buffer
	setDefaultPrinterWriter(t, &buf)

	if err := manager.Status(); err == nil {
		t.Fatal("expected error when CA secret not found")
	}
}

func TestCertManagerStatusMissingClusterIssuer(t *testing.T) {
	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "get", "clusterissuer", certClusterIssuerName) {
				cmd.RunErr = errors.New("not found")
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	manager := NewCertManager(kubectl, zap.NewNop())

	var buf bytes.Buffer
	setDefaultPrinterWriter(t, &buf)

	if err := manager.Status(); err == nil {
		t.Fatal("expected error when ClusterIssuer not found")
	}
}

func TestCertManagerApplyMissingCertManager(t *testing.T) {
	mock := &MockExecutor{
		CommandFunc: func(spec ExecSpec) *MockCommand {
			cmd := &MockCommand{Args: spec.Args}
			if commandHasArgs(spec, "get", "crd", CertManagerCRDName) {
				cmd.RunErr = errors.New("not found")
			}
			return cmd
		},
	}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	manager := NewCertManager(kubectl, zap.NewNop())

	var buf bytes.Buffer
	setDefaultPrinterWriter(t, &buf)

	if err := manager.Apply(); err == nil {
		t.Fatal("expected error when cert-manager not installed")
	}
}

func TestCheckClusterIssuerWithKubectlSuccess(t *testing.T) {
	mock := &MockExecutor{}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := checkClusterIssuerWithKubectl(kubectl); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Commands) != 1 {
		t.Fatalf("expected 1 kubectl command, got %d", len(mock.Commands))
	}
	if !commandHasArgs(mock.Commands[0], "get", "clusterissuer", certClusterIssuerName) {
		t.Fatalf("unexpected args: %v", mock.Commands[0].Args)
	}
}

func TestCheckClusterIssuerWithKubectlError(t *testing.T) {
	mock := &MockExecutor{DefaultRunErr: errors.New("not found")}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := checkClusterIssuerWithKubectl(kubectl); err == nil {
		t.Fatal("expected error when cluster issuer not found")
	}
}

func TestCheckNamedClusterIssuerWithKubectl(t *testing.T) {
	mock := &MockExecutor{}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	if err := checkNamedClusterIssuerWithKubectl(kubectl, " company-ca "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Commands) != 1 || !commandHasArgs(mock.Commands[0], "get", "clusterissuer", "company-ca") {
		t.Fatalf("unexpected command: %v", mock.Commands)
	}
}

func TestCheckNamedClusterIssuerWithKubectlError(t *testing.T) {
	mock := &MockExecutor{DefaultRunErr: errors.New("not found")}
	kubectl := &KubectlClient{exec: mock, validators: nil}
	if err := checkNamedClusterIssuerWithKubectl(kubectl, "missing"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckNamedClusterIssuerWithKubectlEmptyName(t *testing.T) {
	kubectl := &KubectlClient{exec: &MockExecutor{}, validators: nil}
	err := checkNamedClusterIssuerWithKubectl(kubectl, "  ")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !errors.Is(err, ErrClusterIssuerNotFound) {
		t.Fatalf("expected ErrClusterIssuerNotFound, got %v", err)
	}
}

func TestCheckCertificateWithKubectlError(t *testing.T) {
	mock := &MockExecutor{DefaultRunErr: errors.New("not found")}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := checkCertificateWithKubectl(kubectl, "test-cert", "test-ns"); err == nil {
		t.Fatal("expected error when certificate not found")
	}
}

func TestApplyClusterIssuerWithKubectlError(t *testing.T) {
	mock := &MockExecutor{DefaultRunErr: errors.New("apply failed")}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := applyClusterIssuerWithKubectl(kubectl); err == nil {
		t.Fatal("expected error when apply fails")
	}
}

func TestApplyRegistryCertificateWithKubectlError(t *testing.T) {
	mock := &MockExecutor{DefaultRunErr: errors.New("apply failed")}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := applyRegistryCertificateWithKubectl(kubectl); err == nil {
		t.Fatal("expected error when apply fails")
	}
}

func TestWaitForCertificateReadyWithKubectlError(t *testing.T) {
	mock := &MockExecutor{DefaultRunErr: errors.New("timeout")}
	kubectl := &KubectlClient{exec: mock, validators: nil}

	if err := waitForCertificateReadyWithKubectl(kubectl, "test-cert", "test-ns", time.Second); err == nil {
		t.Fatal("expected error when wait times out")
	}
}
