// Package resources provides MCP resource registration for exposing
// Kubernetes cluster information and configuration as MCP resources.
package resources

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMCPResources(s *server.MCPServer) {
	// Register resources
	RegisterK8sContextsMCPResource(s)
}
