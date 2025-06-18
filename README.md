# mcp-k8s

A Model Context Protocol (MCP) server that provides read-only tools for safely interacting with Kubernetes clusters. This server is designed to be safe for production use, deliberately excluding any tools that could mutate cluster state or cause side effects.

## Design Philosophy

This MCP server is intentionally **read-only** and focuses on observability and debugging. It does not include tools that could:

- Modify or delete resources (no `kubectl apply`, `delete`, `patch`, etc.)
- Execute commands in containers (no `kubectl exec`)
- Scale deployments or modify replica counts
- Create port forwards or expose services
- Drain nodes or perform other cluster maintenance operations

This makes it safe to use for debugging production issues without risk of accidental changes.

## Tools

- **`list_k8s_resources`** - List Kubernetes resources of any type with custom formatting for common resource types (pods, deployments, services, etc.)
- **`get_k8s_resource`** - Fetch a single Kubernetes resource with optional Go template formatting for advanced output customization
- **`get_k8s_metrics`** - Get CPU and memory usage metrics for nodes or pods, similar to `kubectl top`, with optional filtering by name (CPU in millicores, memory in MiB). Optional `sum` parameter adds TOTAL entry to results.
- **`get_k8s_pod_logs`** - Get logs from a Kubernetes pod, similar to `kubectl logs`, with options for container selection, time filtering, tail lines, and previous container logs.

## Resources

- **`k8s://contexts`** - Lists available Kubernetes contexts from your kubeconfig file, showing context names, cluster names, and which context is currently active. This helps discover available clusters for use with the tools.

## Prompts

- **`memory_pressure_analysis`** - Analyzes pods for memory pressure issues, including:

  - Pods with memory usage close to their configured limits (>80%)
  - Pods with memory usage significantly exceeding their requests (>120%)
  - Pods that have been OOM killed

  **Arguments:**

  - `context` (required) - The Kubernetes context to use for the analysis
  - `namespace` (optional) - The namespace to analyze (defaults to all namespaces)

  The prompt guides the assistant to use the `get_k8s_metrics` and `list_k8s_resources` tools to identify problematic pods and provide actionable recommendations.

- **`workload_instability_analysis`** - Analyzes Events and pod logs for signs of workload instability, including:

  - Warning Events and failed operations (FailedMount, FailedScheduling, etc.)
  - Error patterns in pod logs (ERROR, FATAL, PANIC messages)
  - Authentication/authorization failures and network connectivity issues
  - Resource exhaustion indicators and application crashes

  **Arguments:**

  - `context` (required) - The Kubernetes context to use for the analysis
  - `namespace` (required) - The namespace to analyze for workload instability

  The prompt guides the assistant to systematically analyze Events and pod logs across all containers, providing a prioritized summary from critical to informational findings.
