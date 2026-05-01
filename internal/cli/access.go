package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	accessGrantResource   = "mcpaccessgrant"
	accessSessionResource = "mcpagentsession"
)

type AccessManager struct {
	kubectl *KubectlClient
	logger  *zap.Logger
	// useKube forces kubectl; when false, the platform API is used when logged in via mcp-runtime auth.
	useKube bool
}

func NewAccessManager(kubectl *KubectlClient, logger *zap.Logger) *AccessManager {
	return &AccessManager{kubectl: kubectl, logger: logger}
}

func DefaultAccessManager(logger *zap.Logger) *AccessManager {
	return NewAccessManager(kubectlClient, logger)
}

// BindUseKubeFlag wires the shared --use-kube flag onto the command.
func (m *AccessManager) BindUseKubeFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&m.useKube, "use-kube", false, "Use kubectl and local kubeconfig instead of the platform API for supported commands")
}

func (m *AccessManager) accessListQueryNamespace(namespace string, allNamespaces bool) string {
	switch {
	case namespace != "":
		return namespace
	case allNamespaces:
		return ""
	default:
		return NamespaceMCPServers
	}
}

// ListAccessResources lists grants or sessions via the platform API when configured, else kubectl.
func (m *AccessManager) ListAccessResources(resource, namespace string, allNamespaces bool) error {
	plat, kube, err := m.platformOrKube()
	if err != nil {
		return err
	}
	if !kube {
		return m.listAccessPlatform(context.Background(), plat, resource, m.accessListQueryNamespace(namespace, allNamespaces))
	}

	args := []string{"get", resource}
	switch {
	case namespace != "":
		args = append(args, "-n", namespace)
	case allNamespaces:
		args = append(args, "-A")
	default:
		args = append(args, "-n", m.accessListQueryNamespace(namespace, allNamespaces))
	}

	if err := m.kubectl.RunWithOutput(args, os.Stdout, os.Stderr); err != nil {
		return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("failed to list %s resources: %v", resource, err), map[string]any{
			"resource":  resource,
			"namespace": namespace,
			"component": "access",
		})
	}
	return nil
}

func (m *AccessManager) listAccessPlatform(ctx context.Context, plat *platformClient, resource, nsFilter string) error {
	switch resource {
	case accessGrantResource:
		grants, err := plat.listGrants(ctx, nsFilter)
		if err != nil {
			return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("list grants: %v", err), map[string]any{"component": "access"})
		}
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "NAME\tNAMESPACE\tSERVER\tDISABLED")
		for _, g := range grants {
			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%v\n", g.Name, g.Namespace, g.ServerRef.Name, g.Disabled)
		}
		_ = tw.Flush()
		return nil
	case accessSessionResource:
		sessions, err := plat.listSessions(ctx, nsFilter)
		if err != nil {
			return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("list sessions: %v", err), map[string]any{"component": "access"})
		}
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "NAME\tNAMESPACE\tSERVER\tREVOKED")
		for _, s := range sessions {
			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%v\n", s.Name, s.Namespace, s.ServerRef.Name, s.Revoked)
		}
		_ = tw.Flush()
		return nil
	default:
		return newWithSentinel(nil, fmt.Sprintf("unsupported access resource %q", resource))
	}
}

func (m *AccessManager) GetAccessResource(resource, name, namespace string) error {
	name, namespace, err := validateAccessInput(name, namespace)
	if err != nil {
		return err
	}

	plat, kube, err := m.platformOrKube()
	if err != nil {
		return err
	}
	if !kube {
		return m.getAccessPlatform(context.Background(), plat, resource, name, namespace)
	}

	args := []string{"get", resource, name, "-n", namespace, "-o", "yaml"}
	if err := m.kubectl.RunWithOutput(args, os.Stdout, os.Stderr); err != nil {
		return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("failed to get %s %q in namespace %q: %v", resource, name, namespace, err), map[string]any{
			"resource":  resource,
			"name":      name,
			"namespace": namespace,
			"component": "access",
		})
	}
	return nil
}

func (m *AccessManager) getAccessPlatform(ctx context.Context, plat *platformClient, resource, name, namespace string) error {
	switch resource {
	case accessGrantResource:
		grant, err := plat.getGrant(ctx, namespace, name)
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(grant, "", "  ")
		_, _ = os.Stdout.Write(append(b, '\n'))
		return nil
	case accessSessionResource:
		session, err := plat.getSession(ctx, namespace, name)
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(session, "", "  ")
		_, _ = os.Stdout.Write(append(b, '\n'))
		return nil
	default:
		return newWithSentinel(nil, fmt.Sprintf("unsupported access resource %q", resource))
	}
}

func (m *AccessManager) ApplyAccessResource(file string) error {
	plat, kube, err := m.platformOrKube()
	if err != nil {
		return err
	}
	if !kube {
		if err := plat.applyAccessFromYAMLFile(context.Background(), file); err != nil {
			return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("apply access resource from file %q: %v", file, err), map[string]any{
				"file":      file,
				"component": "access",
			})
		}
		return nil
	}
	if err := applyManifestFromFile(m.kubectl, file, os.Stdout, os.Stderr); err != nil {
		return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("failed to apply access resource from file %q: %v", file, err), map[string]any{
			"file":      file,
			"component": "access",
		})
	}
	return nil
}

func (m *AccessManager) DeleteAccessResource(resource, name, namespace string) error {
	name, namespace, err := validateAccessInput(name, namespace)
	if err != nil {
		return err
	}

	plat, kube, err := m.platformOrKube()
	if err != nil {
		return err
	}
	if !kube {
		ctx := context.Background()
		switch resource {
		case accessGrantResource:
			err = plat.deleteGrant(ctx, namespace, name)
		case accessSessionResource:
			err = plat.deleteSession(ctx, namespace, name)
		default:
			return newWithSentinel(nil, fmt.Sprintf("unsupported access resource %q", resource))
		}
		if err != nil {
			return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("delete %s %q: %v", resource, name, err), map[string]any{
				"resource":  resource,
				"name":      name,
				"namespace": namespace,
				"component": "access",
			})
		}
		_, _ = fmt.Fprintf(os.Stdout, "%s %q deleted\n", resource, name)
		return nil
	}

	args := []string{"delete", resource, name, "-n", namespace}
	if err := m.kubectl.RunWithOutput(args, os.Stdout, os.Stderr); err != nil {
		return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("failed to delete %s %q in namespace %q: %v", resource, name, namespace, err), map[string]any{
			"resource":  resource,
			"name":      name,
			"namespace": namespace,
			"component": "access",
		})
	}
	return nil
}

func (m *AccessManager) ToggleAccessResource(resource, name, namespace string, value bool) error {
	name, namespace, err := validateAccessInput(name, namespace)
	if err != nil {
		return err
	}

	plat, kube, err := m.platformOrKube()
	if err != nil {
		return err
	}
	if !kube {
		ctx := context.Background()
		switch resource {
		case accessGrantResource:
			if value {
				err = plat.postGrantToggle(ctx, namespace, name, "disable")
			} else {
				err = plat.postGrantToggle(ctx, namespace, name, "enable")
			}
		case accessSessionResource:
			if value {
				err = plat.postSessionToggle(ctx, namespace, name, "revoke")
			} else {
				err = plat.postSessionToggle(ctx, namespace, name, "unrevoke")
			}
		default:
			return newWithSentinel(nil, fmt.Sprintf("unsupported access resource %q", resource))
		}
		if err != nil {
			return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("toggle %s %q: %v", resource, name, err), map[string]any{
				"resource":  resource,
				"name":      name,
				"namespace": namespace,
				"component": "access",
			})
		}
		_, _ = fmt.Fprintf(os.Stdout, "updated %s %q\n", resource, name)
		return nil
	}

	patchValue := map[string]any{"spec": map[string]any{}}
	switch resource {
	case accessGrantResource:
		patchValue["spec"].(map[string]any)["disabled"] = value
	case accessSessionResource:
		patchValue["spec"].(map[string]any)["revoked"] = value
	default:
		return newWithSentinel(nil, fmt.Sprintf("unsupported access resource %q", resource))
	}

	data, err := json.Marshal(patchValue)
	if err != nil {
		return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("failed to marshal access patch for %s %q: %v", resource, name, err), map[string]any{
			"resource":  resource,
			"name":      name,
			"namespace": namespace,
			"component": "access",
		})
	}

	args := []string{"patch", resource, name, "-n", namespace, "--type", "merge", "--patch", string(data)}
	if err := m.kubectl.RunWithOutput(args, os.Stdout, os.Stderr); err != nil {
		return wrapWithSentinelAndContext(nil, err, fmt.Sprintf("failed to patch %s %q in namespace %q: %v", resource, name, namespace, err), map[string]any{
			"resource":  resource,
			"name":      name,
			"namespace": namespace,
			"component": "access",
		})
	}
	return nil
}

func validateAccessInput(name, namespace string) (string, string, error) {
	if !validServerName.MatchString(name) {
		return "", "", newWithSentinel(nil, fmt.Sprintf("invalid resource name %q: must be lowercase alphanumeric with optional hyphens", name))
	}

	var err error
	if name, err = validateManifestValue("name", name); err != nil {
		return "", "", err
	}
	if namespace, err = validateManifestValue("namespace", namespace); err != nil {
		return "", "", err
	}

	return name, namespace, nil
}
