# CLI Entrypoint

Package `cmd/mcp-runtime` builds the user-facing `mcp-runtime` binary. It should
stay thin: initialize logging, assemble Cobra commands, execute the root command,
and exit with a clear status.

Useful reference command:

```bash
go doc -cmd ./cmd/mcp-runtime
```

## Responsibilities

- Define build metadata variables (`version`, `commit`, `date`) that release
  builds can set with linker flags.
- Create the root Cobra command and global flags.
- Register subcommands through the foldered `internal/cli/root` routing layer.
- Configure a console-oriented zap logger.
- Print command errors to stderr and return a non-zero process exit.

The entrypoint should not contain business logic for setup, registry, server,
access, or Sentinel behavior. Route top-level commands through
`internal/cli/root`, and keep command behavior in `internal/cli` until a focused
migration moves that domain into its own command package.

## Command Tree

The root command wires these internal command groups:

| Command | Routing package | Behavior files |
|---|---|---|
| `bootstrap` | `internal/cli/bootstrap` | `internal/cli/bootstrap.go` |
| `cluster` | `internal/cli/cluster` | `internal/cli/cluster.go` and `cluster_doctor.go` |
| `setup` | `internal/cli/setup` | `internal/cli/setup.go`, `setup_plan.go`, `setup_steps.go` |
| `status` | `internal/cli/status` | `internal/cli/status.go` |
| `registry` | `internal/cli/registry` | `internal/cli/registry.go` |
| `server` | `internal/cli/server` | `internal/cli/server.go`, `build.go` |
| `pipeline` | `internal/cli/pipeline` | `internal/cli/pipeline.go` |
| `access` | `internal/cli/access` | `internal/cli/access.go` |
| `auth` | `internal/cli/auth` | `internal/cli/auth.go` |
| `sentinel` | `internal/cli/sentinel` | `internal/cli/sentinel.go` |

When adding a command, wire it here only after the implementation has focused
package tests and help text is ready for golden snapshots.

## Contributor Contract

CLI UX changes should preserve these expectations:

- Help text is accurate and reflected in `test/golden/cli/testdata`.
- Errors are human-readable and still return non-zero exit codes.
- Logs are readable in terminals and CI.
- Global flags stay minimal; feature-specific flags belong on their command.
- Commands that shell out to external tools are testable through runner
  abstractions in `internal/cli`.
- Top-level command folders under `internal/cli/<command>` should stay thin
  while they delegate to `internal/cli`; move behavior there only as a focused
  follow-up with package-local tests.

Before changing this package, run:

```bash
go test ./cmd/mcp-runtime ./internal/cli/... ./test/golden/... -count=1
```
