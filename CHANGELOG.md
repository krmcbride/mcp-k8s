# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial MCP (Model Context Protocol) server implementation for Kubernetes operations
- Kubernetes resource listing with custom formatting (`list_k8s_resources`)
- Kubernetes API resource discovery (`list_k8s_api_resources`)
- Single resource retrieval with Go template support (`get_k8s_resource`)
- CPU/memory metrics for nodes and pods (`get_k8s_metrics`)
- Pod log retrieval (`get_k8s_pod_logs`)
- Kubernetes context discovery via MCP resource (`kubeconfig://contexts`)
- Memory pressure analysis prompt
- Workload instability analysis prompt
- Comprehensive resource mappers for Pods, Deployments, Services, Events, etc.
- Vendor directory for offline development
- GitHub Actions CI pipeline with cross-platform builds
- Automated release workflow with GoReleaser

### Documentation
- Comprehensive CLAUDE.md with modular documentation structure
- Development best practices and guidelines
- CI/CD documentation and patterns
- MCP server development guides
- Kubernetes client architecture documentation

### Development
- Makefile with development, testing, and CI targets
- golangci-lint, gofumpt, and goimports-reviser integration
- Interactive MCP testing with mcptools
- Comprehensive test suite for resource mappers
- Case-insensitive resource kind lookup