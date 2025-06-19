# Kubernetes Client Architecture

## Client Type Selection Guide

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

## Client Creation Patterns

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

## GVK to GVR Conversion

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

