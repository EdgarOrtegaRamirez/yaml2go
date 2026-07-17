# yaml2go — AI Agent Guide

## Project Overview
yaml2go is a Go CLI tool that parses YAML files and generates idiomatic Go structs with JSON/YAML tags. It handles type inference, nested objects, arrays, and supports configurable field naming conventions.

## Build & Test
```bash
# Build
cd /root/workspace/yaml2go
go build -o yaml2go ./cmd/yaml2go/

# Run tests (25 tests)
go test ./...

# Run with example YAML
go run ./cmd/yaml2go/ tests/fixtures/config.yaml
```

## Architecture
- `cmd/yaml2go/main.go` — CLI entry, argument parsing, YAML parsing, struct generation
- No separate package layer — single-file design for simplicity

## Key Design Decisions
- All processing is local — no network access
- Uses yaml.v3 for parsing (handles YAML 1.2)
- Nested YAML objects become inline Go structs (no separate type definitions)
- Type inference uses yaml.v3 type tags (!!str, !!int, !!bool, etc.)
- Field naming: pascal (default), camel, or snake_case
- Struct tags: json and yaml (configurable, can disable with --no-tags)

## Dependencies
- gopkg.in/yaml.v3 v3.0.1 — YAML parsing (YAML 1.2)

## Test Fixtures
- Use any YAML file for testing: `go run ./cmd/yaml2go/ <file.yaml>`