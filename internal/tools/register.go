// Package tools provides MCP tools for interacting with Kubernetes resources.
// It includes tools for listing and retrieving Kubernetes resources with
// customizable output formatting through resource mappers.
package tools

import (
	"github.com/krmcbride/mcp-k8s/internal/tools/mapper"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMCPTools(s *server.MCPServer) {
	// Initialize resource mappers
	mapper.Init()

	// Register tools
	RegisterListK8sResourcesMCPTool(s)
	RegisterGetK8sResourceMCPTool(s)
	RegisterGetK8sMetricsMCPTool(s)
}
