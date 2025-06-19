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
- **MCP Server Testing**: Use `make mcp-shell` for interactive testing rather than piping JSON-RPC directly to the server, which will hang waiting for a persistent connection

### CI Commands

- `make test-ci` - Run tests with coverage
- `make lint-ci` - Run linters for CI
- `make format-ci` - Check formatting in CI

## MCP Components

### Tools

- **`list_k8s_resources`** - List Kubernetes resources with custom formatting for common types
- **`list_k8s_api_resources`** - List available Kubernetes API resource types (equivalent to kubectl api-resources)
- **`get_k8s_resource`** - Fetch single Kubernetes resource with optional Go template formatting
- **`get_k8s_metrics`** - Get CPU/memory metrics for nodes or pods (similar to kubectl top)
- **`get_k8s_pod_logs`** - Get logs from Kubernetes pods (similar to kubectl logs)

### Resources

**Kubernetes Contexts** (`kubeconfig://contexts`)

- Exposes the current user's kubeconfig contexts as an MCP resource
- Returns JSON array with context name, cluster name, and current context indicator
- **IMPORTANT**: Use this resource to resolve cluster aliases (like 'prod', 'sandbox') to actual context names instead of running kubectl commands
- Enables discovery of available contexts for use with the tools
- Allows matching context names to cluster names for intuitive queries

### Prompts

**Memory Pressure Analysis** (`memory_pressure_analysis`)

- Analyzes pods for memory pressure issues including high usage, exceeding requests, and OOM kills
- Required argument: `context` (Kubernetes context)
- Optional argument: `namespace` (defaults to all namespaces)
- Guides assistant to use metrics and resource tools for comprehensive analysis

**Workload Instability Analysis** (`workload_instability_analysis`)

- Analyzes Events and pod logs for signs of workload instability including errors, warnings, and suspicious patterns
- Required argument: `context` (Kubernetes context)
- Required argument: `namespace` (target namespace to analyze)
- Guides assistant to systematically analyze Events and pod logs across all containers, providing prioritized findings from critical to informational

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
- Currently registers: list_k8s_resources, list_k8s_api_resources, get_k8s_resource, get_k8s_metrics, and get_k8s_pod_logs tools

**Kubernetes Client Layer** (`internal/k8s/`)

- `client.go`: Kubernetes client factory with context switching support and discovery client for API resource enumeration
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
- Event (core/v1 and events.k8s.io/v1beta1) (cluster events)
- CustomResourceDefinition (apiextensions.k8s.io/v1 and v1beta1) (CRD discovery)

Each mapper extracts resource-specific fields (e.g., replica counts, status, networking details) rather than just name/namespace.

## Adding New Resource Mappers

1. Create new file in `internal/tools/mapper/` (e.g., `configmap.go`)
2. Define content struct with json tags
3. Implement mapper function extracting relevant fields from unstructured data
4. Add init() function to register the mapper
5. Update integration test in `integration_test.go`
6. **IMPORTANT**: Update the Resource Mappers list in this documentation
7. **Update `CHANGELOG.md`** under `[Unreleased]` section with the new mapper

## Adding New MCP Tools

When adding new MCP tools, ensure documentation is updated in both locations:

1. **Implementation Steps:**

   - Create new tool file in `internal/tools/` (e.g., `new_tool.go`)
   - Register the tool in `internal/tools/register.go`
   - Add any new client functions to `internal/k8s/client.go` if needed
   - Test with `make build` and `make test`

2. **Documentation Updates (REQUIRED):**

   - Add tool to the Tools section in `CLAUDE.md` (line ~35)
   - Add tool to the Tools section in `README.md` (line ~17)
   - Update the Tool Registration section in `CLAUDE.md` (line ~84)
   - If adding new client capabilities, update Kubernetes Client Layer section (line ~88)
   - **Update `CHANGELOG.md`** under `[Unreleased]` section with the new feature

3. **Validation:**
   - Grep for the tool name across documentation files to ensure consistency
   - Verify all references are updated and accurate

## Testing Strategy

- Comprehensive unit tests in `mapper_test.go` covering case variations and edge cases
- Integration test in `integration_test.go` verifying all expected mappers are registered
- Tests clear the mapper registry to ensure isolation between test cases

## Kubernetes Integration

The system uses kubeconfig contexts for cluster access. The GVKToGVR function handles:

- Context-specific client creation
- REST mapper discovery for accurate Kind → Resource conversion
- Support for built-in resources and CRDs

## Detailed Development Guides

For comprehensive development guidance, see the following detailed guides:

@docs/development-best-practices.md

@docs/dependency-management.md

@docs/ci-best-practices.md

@docs/mcp-server-development.md

@docs/kubernetes-client-architecture.md
