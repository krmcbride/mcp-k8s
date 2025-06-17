package resources

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMCPResources(s *server.MCPServer) {
	// Register resources
	RegisterK8sContextsMCPResource(s)
}
