package tools

import (
	"github.com/krmcbride/mcp-k8s/internal/tools/mapper"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {
	// Initialize resource mappers
	mapper.Init()

	// Register tools
	RegisterListResourcesTool(s)
	RegisterGetResourceTool(s)
}
