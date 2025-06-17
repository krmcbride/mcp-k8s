# mcp-k8s

A Model Context Protocol (MCP) server that provides tools for interacting with Kubernetes clusters.

## Tools

- **`list_k8s_resources`** - List Kubernetes resources of any type with custom formatting for common resource types (pods, deployments, services, etc.)
- **`get_k8s_resource`** - Fetch a single Kubernetes resource with optional Go template formatting for advanced output customization
- **`get_k8s_metrics`** - Get CPU and memory usage metrics for nodes or pods, similar to `kubectl top`, with optional filtering by name (CPU in millicores, memory in MiB). Optional `sum` parameter adds TOTAL entry to results.
- **`get_k8s_pod_logs`** - Get logs from a Kubernetes pod, similar to `kubectl logs`, with options for container selection, time filtering, tail lines, and previous container logs.
