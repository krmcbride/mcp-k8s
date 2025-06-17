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
}
