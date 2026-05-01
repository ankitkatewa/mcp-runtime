// Package cmd provides the foldered command routing layer for the mcp-runtime
// binary.
//
// Each subpackage owns one top-level Cobra command boundary and currently
// delegates behavior to internal/cli. Keeping this layer separate from the CLI
// implementation lets the binary wire commands through stable folders while the
// larger command implementations can be migrated incrementally.
package cmd
