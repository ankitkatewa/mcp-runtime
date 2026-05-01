# CLI Internals

Package `internal/cli` implements the command behavior behind the
`mcp-runtime` binary. The top-level Cobra command folders live under
`internal/cmd`; those packages route to this package while command behavior is
split out incrementally. Both layers are intentionally internal so the CLI can
evolve without becoming a public Go API.

`go doc` is still useful for exported constructors and manager types:

```bash
go doc -all ./internal/cli
```

Most command behavior is unexported and should be understood through this page,
tests, and the command help snapshots.

## Design Principles

- Keep command wiring separate from side effects where practical.
- Prefer runner interfaces for `kubectl`, `docker`, and process execution so
  behavior can be tested without a live cluster.
- Return structured errors from shared helpers and print concise user-facing
  messages at command boundaries.
- Keep command help accurate and update golden snapshots for intentional output
  changes.
- Treat Kubernetes manifests and CRD types as source of truth; CLI structs should
  not drift from `api/v1alpha1`.

## Common Infrastructure

| File group | Responsibility |
|---|---|
| `constants.go` | namespace, deployment, service, and resource names shared by commands |
| `errors.go` | sentinel error values and wrapping helpers |
| `exec.go`, `kubectl_runner.go` | external command execution and test seams |
| `printer.go`, `output.go` | terminal output formatting |
| `asset_paths.go` | locating repository and manifest assets |
| `config.go` | environment/config defaults for registry, ingress, and setup |
| `resource_helpers.go` | shared Kubernetes resource and manifest helpers |

When adding a helper, put it near the command that owns it unless two or more
commands genuinely share it.

## Setup

Setup is split across:

- `setup.go`: Cobra command, setup orchestration, image publishing, manifest
  application, verification, and deployment diagnostics.
- `setup_plan.go`: planning and dependency injection seams used by tests.
- `setup_steps.go`: step-level helpers used by setup orchestration.

`setup --test-mode` relaxes production guardrails but still builds and pushes
the operator, gateway proxy, and Sentinel images with `latest` tags. Pull hosts
must still be reachable and trusted by node container runtimes. On k3s with the
bundled HTTP registry, that means a `registries.yaml` mirror for the exact
registry host/port used in pod image refs.

Important setup contracts:

- CRDs and namespaces are applied before runtime components.
- Ingress is installed or reused before registry routes are needed.
- Registry info is resolved before runtime images are named.
- Internal registry pushes use an in-cluster helper when direct host pushes are
  not appropriate.
- Sentinel rollouts use `MCP_DEPLOYMENT_TIMEOUT`.
- Setup verification should fail with diagnostic context instead of reporting
  success after partial deployment.

Tests: `setup_test.go`, `setup_helpers_test.go`, `setup_plan_test.go`, and
`setup_steps_test.go`.

## Cluster and Doctor

`cluster.go` provides cluster initialization, status, ingress configuration, and
provider-oriented provisioning helpers. `bootstrap.go` performs preflight checks
and has the only automated apply path for k3s CoreDNS/local-path prerequisites.

`cluster_doctor.go` is post-install diagnostics. It checks CRDs, workloads,
registry reachability, image pull failures, ingress, and platform components.
Registry protocol mismatch detection must inspect regular containers and init
containers, and it must surface failed pod inspections instead of returning a
false pass.

Tests: `cluster_test.go`, `cluster_doctor_test.go`, and bootstrap-related tests.

## Registry

`registry.go` owns registry status, info, provisioning, login, direct pushes, and
in-cluster helper pushes.

Registry endpoint precedence is intentionally shared with setup and metadata:

- explicit CLI flags
- environment variables/config
- platform-derived registry host
- bundled registry service discovery

The in-cluster push path uses a temporary helper workload and should clean up
after itself even on failure. When editing this path, verify both success and
diagnostic failure output.

Tests: `registry_test.go`, plus setup tests for runtime image publishing.

## Server and Build

`server.go` implements CRUD-style operations for `MCPServer` resources and
status/log inspection. `build.go` supports metadata-driven image builds for the
`.mcp` workflow.

Keep these flows distinct:

- `server apply` applies a manifest.
- `server build image` builds and updates metadata but does not deploy.
- `registry push` publishes images.
- `pipeline generate` renders manifests from metadata.
- `pipeline deploy` applies generated manifests.

Tests: `server_test.go`, `server_config_test.go`, `build_test.go`, and pipeline
tests.

## Pipeline

`pipeline.go` turns `.mcp` metadata into `MCPServer` manifests and applies
rendered directories. It is the CLI bridge to `pkg/metadata`.

Pipeline changes usually require checking:

- metadata schema and defaulting
- generated YAML shape
- docs for `.mcp` authoring
- examples under `examples/`

Tests: `pipeline_test.go` and `pkg/metadata` tests.

## Access

`access.go` provides commands for grants and sessions:

- `access grant list|get|apply|delete|enable|disable`
- `access session list|get|apply|delete|revoke|unrevoke`

The implementation patches `spec.disabled` for grants and `spec.revoked` for
sessions. Input validation should prevent invalid names/namespaces before they
reach `kubectl`.

Tests: `access_test.go`.

## Sentinel and Platform API

`sentinel.go`, `auth.go`, `platform_client.go`, and `platform_ingress.go` provide
CLI access to Sentinel APIs, auth flows, and platform ingress resolution. These
commands should stay aligned with `services/api` routes and the public docs.

Tests: `sentinel_test.go`, `auth_test.go`, `platform_client_test.go`, and
`platform_ingress_test.go`.

## Status

`status.go` prints high-level platform health by querying Kubernetes. It should
be quick, readable, and conservative. Deeper diagnosis belongs in
`cluster doctor`.

Tests: `status_test.go`.

## Adding a Command

1. Add the command implementation in the closest existing file or a new focused
   file under `internal/cli`.
2. Add or update the thin routing package under `internal/cmd/<command>`.
3. Register the top-level command from `internal/cmd/commands.go`.
4. Add tests with mocked runners or fake dependencies.
5. Build the CLI and inspect `--help`.
6. Update golden snapshots if help/output changes intentionally.
7. Update user docs when behavior is user-facing.

Run:

```bash
go test ./internal/cmd/... ./internal/cli/... ./cmd/mcp-runtime ./test/golden/... -count=1
```
