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

## Development Best Practices

### Architecture-First Approach

When implementing new features, start with architectural planning:

- Identify key interfaces and integration points
- Consider consistency requirements across related operations (e.g., registration and lookup)
- Anticipate potential edge cases and normalization needs

**API Research Protocol**

- Use Context7 proactively to research exact API structures before implementation
- For unfamiliar libraries or APIs, always verify field names, types, and methods
- Don't assume API structures - verify through documentation before coding

**Incremental Development**

- Break complex features into smaller, testable increments
- Validate each component with `make build` and `make test` before proceeding
- Test intermediate states rather than implementing entire features before first validation
- **CRITICAL**: Build and test after each significant change to catch issues early
- Use `git status` to verify file changes before and after bulk operations

### Modern Go Guidelines

**Type Declarations**

- Use `any` instead of `interface{}` for better readability
- Example: `var content any` instead of `var content interface{}`
- This applies to function parameters, return types, and variable declarations

**Safe File Manipulation**

- **AVOID** batch `sed` operations on multiple files - they can corrupt files
- Use `MultiEdit` tool for multiple related changes in a single file
- Use individual `Edit` calls for changes across multiple files
- Always check `git status` before and after bulk operations
- Use `git restore <file>` to recover from file corruption
- Test build after any file manipulation to ensure no corruption

**Error Handling**

- Avoid error variable shadowing in nested scopes
- Use descriptive error variable names when redeclaring in inner scopes
- Example: Use `parseErr` instead of redeclaring `err` in parsing operations

**Enhanced Error Messaging**

- When errors relate to invalid context names, enhance them to mention the `kubeconfig://contexts` MCP resource
- Pattern: Detect context-related errors and append guidance about using MCP resources instead of kubectl commands
- Example: "context 'sandbox' does not exist. To discover available contexts or resolve cluster aliases, use the kubeconfig://contexts MCP resource instead of running kubectl commands"
- Apply this pattern in client creation functions where context errors are likely

**Safety-First Design**

- All MCP tools are deliberately read-only to prevent accidental cluster modifications
- Tools provide comprehensive data for analysis without mutation capabilities
- Resource mappers extract relevant fields while preserving original structure

**Resource and Performance Considerations**

- Consider MCP token limits when designing features (25k response limit)
- Account for memory usage patterns, especially for large resource collections
- Design with pagination and chunking in mind for scalable operations
- Document resource usage implications (e.g., "100 Events ≈ 12k tokens")

**Backward Compatibility**

- When adding new optional parameters, ensure existing usage patterns remain functional
- Test that new features don't break existing tool behavior
- Design APIs to be extensible without breaking changes

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
- **Safety Protocol**: Always run `make build` and `make test` after file modifications
- If tests fail after bulk file operations, check for file corruption with `git status`
- Use incremental approach: make one change, test, then proceed to next change

### Consistency Validation

When implementing related operations (like mapper registration and lookup):

- Ensure the same normalization/transformation is applied in all related functions
- Test that different input variations (case, format) work consistently
- Verify that registration and retrieval use identical key generation logic

## Dependency Management

### Vendored Dependencies

This project uses Go vendoring to provide direct access to dependency source code:

- **Vendor Directory**: `vendor/` contains all dependency source code locally
- **Git Ignore**: Vendor directory is excluded from version control
- **Makefile Integration**: All Go commands use `GOFLAGS="-mod=vendor"` automatically
- **Development Workflow**: Run `go mod vendor` after dependency changes to update vendor/

**Claude Development Benefits:**

- Can read actual source code from `vendor/k8s.io/client-go/` instead of relying on external documentation
- Enables precise API research by examining actual method signatures and implementations
- Provides accurate error handling patterns and usage examples directly from source
- **IMPORTANT**: Always prefer reading vendor source code over guessing API structures

**Vendor Maintenance:**

- Run `go mod vendor` after updating `go.mod`
- Vendor directory is automatically excluded from git commits
- All build/test/lint commands automatically use vendored dependencies

### Key Dependencies

- `k8s.io/client-go`: Kubernetes Go client library
- `k8s.io/apimachinery`: Core Kubernetes types and utilities
- `github.com/mark3labs/mcp-go`: MCP protocol implementation

## Enhanced MCP Server Development

### MCP Tool Development Patterns

**Tool Structure Template:**

```go
func NewToolName(serverName string) mcp.Tool {
    return mcp.NewTool(
        serverName+"_tool_name",
        "Description of what this tool does",
        map[string]mcp.ToolInputSchema{
            "required_param": {
                Type: "string",
                Description: "Enhanced description mentioning kubeconfig://contexts for context discovery",
            },
        },
        handleToolName,
    )
}
```

**Error Handling Best Practices:**

- Always enhance context-related errors with MCP resource guidance
- Use structured error responses with actionable suggestions
- Log errors to stderr only (never stdout due to stdio transport)
- Provide context about which MCP resources can help resolve issues

**Interactive Testing Workflow:**

1. Use `make mcp-shell` for interactive tool testing
2. Test edge cases like invalid contexts, missing resources
3. Verify error messages guide users to appropriate MCP resources
4. Validate tool descriptions are clear and actionable

**Tool Registration Pattern:**

- Register tools in `internal/tools/register.go`
- Update both CLAUDE.md and README.md documentation
- Add integration tests for new tools
- Follow naming convention: `{verb}_k8s_{resource_type}`

### MCP Resource Development

**Resource Naming Convention:**

- Use descriptive URI schemes: `kubeconfig://contexts`
- Provide clear resource descriptions in registration
- Enable resource discovery through intuitive naming

**Resource Implementation Guidelines:**

- Return structured data suitable for Claude analysis
- Include metadata that helps with decision making
- Design resources to reduce need for external commands
- Document resource schema and expected usage patterns

### MCP Prompt Development

**Prompt Structure Requirements:**

- Required parameters must be clearly specified
- Optional parameters should have sensible defaults
- Provide clear guidance on what analysis to perform
- Structure prompts to guide systematic investigation

**Analysis Prompt Pattern:**

```go
// Guides assistant through systematic analysis
// Required: context, optional: namespace
// Provides step-by-step investigation approach
```

## Kubernetes Client Architecture

### Client Type Selection Guide

**Discovery Client** (`internal/k8s/client.go:GetDiscoveryClientForContext`)

- **Use For**: API resource discovery, server version info, available API groups
- **Examples**: `list_k8s_api_resources` tool, API capability checking
- **Methods**: `ServerGroupsAndResources()`, `ServerVersion()`
- **Performance**: Lightweight, cached API discovery

**Dynamic Client** (`dynamic.NewForConfig`)

- **Use For**: Generic resource operations, CRD support, unstructured data
- **Examples**: `list_k8s_resources`, `get_k8s_resource` tools
- **Methods**: `Resource().List()`, `Resource().Get()`
- **Benefits**: Works with any resource type without code generation

**REST Client** (`rest.RESTClientFor`)

- **Use For**: Custom API endpoints, non-standard operations
- **Examples**: Metrics API calls, custom resource endpoints
- **Methods**: Direct HTTP operations with Kubernetes authentication
- **Use Cases**: When dynamic client doesn't provide needed functionality

**Typed Clients** (e.g., `kubernetes.NewForConfig`)

- **Use For**: Standard Kubernetes resources with strong typing
- **Examples**: When you need compile-time type safety
- **Trade-offs**: Requires updates for API changes, doesn't support CRDs
- **Current Usage**: Not used in this project (prefer dynamic client)

### Client Creation Patterns

**Context-Aware Client Factory:**

```go
// Always enhance context errors with MCP resource guidance
func GetClientForContext(context string) (dynamic.Interface, error) {
    config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
        // ... config loading
    ).ClientConfig()
    if err != nil {
        return nil, enhanceContextError(err) // Guide to kubeconfig://contexts
    }
    return dynamic.NewForConfig(config)
}
```

**Error Enhancement Pattern:**

- Detect context-related errors in client creation
- Append guidance about `kubeconfig://contexts` MCP resource
- Provide actionable suggestions instead of raw errors
- Help users discover available contexts through MCP instead of kubectl

### GVK to GVR Conversion

**REST Mapper Usage:**

- Use discovery client to build REST mapper
- Handle both built-in resources and CRDs
- Cache mapper for performance in repeated operations
- Gracefully handle API discovery failures

**Kind Resolution Strategy:**

1. Try exact Kind match first
2. Fall back to case-insensitive lookup
3. Use resource mappers for common variations
4. Provide clear error messages for unrecognized kinds
