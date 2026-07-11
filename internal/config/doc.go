// Package config is layer L2 infrastructure: the Configuration Manager implementing
// ports.ConfigPort (FR-CFG-001). It resolves the effective configuration by applying the
// precedence order of Volume 10 over the andromeda.toml layers, environment variables
// (ANDROMEDA_*), and invocation overrides, attributing every resolved value to the layer that
// supplied it. TOML parsing uses pelletier/go-toml/v2 (ADR-008); typed schema validation
// (ADR-024) grows in later epics.
//
// Precedence, highest wins (Volume 0 chapter 03 / Volume 10):
//
//	cli-flag > env > runtime-override > project > workspace > profile > global > defaults
package config
