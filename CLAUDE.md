# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an MCP (Model Context Protocol) server that provides tools for interacting with Kubernetes clusters. The server exposes Kubernetes operations through MCP tools that can be used by Claude and other MCP clients.

## Key Commands

### Development

- `make test` - Run all tests
- `go test ./internal/tools/mapper -v` - Run mapper tests specifically
- `go test ./internal/tools/mapper -v -run TestName` - Run specific test
- `make build` - Build the MCP server binary
- `make fmt` - Format code (runs gofumpt and goimports-reviser)
- `make lint` - Run golangci-lint

### MCP Development

- `make mcp-shell` - Run the MCP server interactively with mcptools shell for testing

### CI Commands

- `make test-ci` - Run tests with coverage
- `make lint-ci` - Run linters for CI
- `make format-ci` - Check formatting in CI

## MCP Components

### Tools

- **`list_k8s_resources`** - List Kubernetes resources with custom formatting for common types
- **`get_k8s_resource`** - Fetch single Kubernetes resource with optional Go template formatting
- **`get_k8s_metrics`** - Get CPU/memory metrics for nodes or pods (similar to kubectl top)
- **`get_k8s_pod_logs`** - Get logs from Kubernetes pods (similar to kubectl logs)

### Resources

**Kubernetes Contexts** (`k8s://contexts`)

- Exposes available kubeconfig contexts as an MCP resource
- Returns JSON array with context name, cluster name, and current context indicator
- Enables discovery of available contexts for use with the tools
- Allows matching context names to cluster names for intuitive queries

### Prompts

**Memory Pressure Analysis** (`memory_pressure_analysis`)

- Analyzes pods for memory pressure issues including high usage, exceeding requests, and OOM kills
- Required argument: `context` (Kubernetes context)
- Optional argument: `namespace` (defaults to all namespaces)
- Guides assistant to use metrics and resource tools for comprehensive analysis

## Architecture

### Core Components

**MCP Server Entry Point** (`cmd/server/main.go`)

- Creates MCP server instance using mark3labs/mcp-go
- Registers all MCP components:
  - `prompts.RegisterMCPPrompts()`
  - `resources.RegisterMCPResources()`
  - `tools.RegisterMCPTools()`
- Serves over stdio protocol

**Tool Registration** (`internal/tools/register.go`)

- Central registration point for all MCP tools
- Initializes resource mappers before registering tools
- Currently registers: list_k8s_resources, get_k8s_resource, get_k8s_metrics, and get_k8s_pod_logs tools

**Kubernetes Client Layer** (`internal/k8s/`)

- `client.go`: Kubernetes client factory with context switching support
- `gvr.go`: GVK (GroupVersionKind) to GVR (GroupVersionResource) conversion using REST mapper

**Resource Mapping System** (`internal/tools/mapper/`)

- Extensible system for converting Kubernetes unstructured resources into structured output
- Case-insensitive Kind lookup with automatic normalization
- Auto-registration via init() functions in individual resource files

### Key Design Patterns

**Resource Mapper Registration**
Each resource type (pod.go, deployment.go, etc.) auto-registers its mapper in an init() function:

```go
func init() {
    Register(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, mapDeploymentResource)
}
```

**Case-Insensitive GVK Normalization**
The mapper system normalizes Kind names to title case for consistent map keys, allowing users to specify "pod", "Pod", or "POD" interchangeably.

**Dynamic Client Usage**
Uses Kubernetes dynamic client instead of typed clientset to work with any resource type (including CRDs) without code generation.

## Resource Mappers

Currently implemented mappers for:

- Pod, Deployment, DaemonSet, StatefulSet, Job, CronJob (workloads)
- Service, Ingress (networking)
- Node (infrastructure)

Each mapper extracts resource-specific fields (e.g., replica counts, status, networking details) rather than just name/namespace.

## Adding New Resource Mappers

1. Create new file in `internal/tools/mapper/` (e.g., `configmap.go`)
2. Define content struct with json tags
3. Implement mapper function extracting relevant fields from unstructured data
4. Add init() function to register the mapper
5. Update integration test in `integration_test.go`

## Testing Strategy

- Comprehensive unit tests in `mapper_test.go` covering case variations and edge cases
- Integration test in `integration_test.go` verifying all expected mappers are registered
- Tests clear the mapper registry to ensure isolation between test cases

## Kubernetes Integration

The system uses kubeconfig contexts for cluster access. The GVKToGVR function handles:

- Context-specific client creation
- REST mapper discovery for accurate Kind → Resource conversion
- Support for built-in resources and CRDs

## Development Best Practices

### Architecture-First Approach

When implementing new features, start with architectural planning:

- Identify key interfaces and integration points
- Consider consistency requirements across related operations (e.g., registration and lookup)
- Anticipate potential edge cases and normalization needs

### Modern Go Guidelines

**Type Declarations**
- Use `any` instead of `interface{}` for better readability
- Example: `var content any` instead of `var content interface{}`
- This applies to function parameters, return types, and variable declarations

**Error Handling**
- Avoid error variable shadowing in nested scopes
- Use descriptive error variable names when redeclaring in inner scopes
- Example: Use `parseErr` instead of redeclaring `err` in parsing operations

**Safety-First Design**
- All MCP tools are deliberately read-only to prevent accidental cluster modifications
- Tools provide comprehensive data for analysis without mutation capabilities
- Resource mappers extract relevant fields while preserving original structure

**MCP Server Logging**
- **CRITICAL**: When using stdio transport, all logging MUST go to stderr only
- The MCP protocol uses stdout for communication; any output to stdout will corrupt the protocol
- Use `fmt.Fprintf(os.Stderr, ...)` or configure loggers to write to stderr
- This applies to all debug messages, status updates, and error logging
- Reference: MCP specification states servers can write logs to stderr while using stdout for protocol messages

### Documentation-Driven Development

- Write comprehensive comments for public APIs during implementation, not after
- Explain design decisions and component relationships
- Include usage examples for complex interfaces
- Maintain package comments in the primary file of each package (except `main`):
  - Add to `register.go` for packages focused on registration
  - Add to the primary interface file (e.g., `client.go`, `mapper.go`) for functionality packages
  - Update package comments when the package's purpose changes
  - Format: `// Package <name> provides...` (2-3 lines describing the package's purpose)

### Enhanced Resource Mapping

**Pod Mapper Capabilities**
- Extracts memory resource specifications (requests/limits) in standardized MiB units
- Detects OOM kill events from container status history
- Supports complex memory unit parsing (Mi, Gi, bytes, etc.)
- Provides termination reason tracking for debugging

**Memory Unit Conversion**
- Standardizes all memory values to MiB for consistency
- Handles Kubernetes memory formats: "128Mi", "1Gi", "512000000", "1000000k"
- Enables direct numerical comparison and calculation

### Test-Driven Validation

- Propose and implement tests early in the feature development process
- Use tests to validate design assumptions and catch integration issues
- Test edge cases like case variations, empty inputs, and error conditions

### Consistency Validation

When implementing related operations (like mapper registration and lookup):

- Ensure the same normalization/transformation is applied in all related functions
- Test that different input variations (case, format) work consistently
- Verify that registration and retrieval use identical key generation logic

## Dependencies

- `k8s.io/client-go`: Kubernetes Go client library
- `k8s.io/apimachinery`: Core Kubernetes types and utilities
- `github.com/mark3labs/mcp-go`: MCP protocol implementation
