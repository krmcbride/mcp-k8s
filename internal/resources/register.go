package resources

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterResources(s *server.MCPServer) {
	// Register resources
	RegisterContextsResource(s)
}
