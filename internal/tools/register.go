package tools

import "github.com/mark3labs/mcp-go/server"

func RegisterTools(s *server.MCPServer) {
	// Add tool handler
	RegisterHelloworldTool(s)
}
