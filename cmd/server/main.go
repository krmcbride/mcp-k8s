package main

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"

	"github.com/krmcbride/mcp-k8s/internal/prompts"
	"github.com/krmcbride/mcp-k8s/internal/resources"
	"github.com/krmcbride/mcp-k8s/internal/tools"
)

func main() {
	s := server.NewMCPServer(
		"mcp-k8s",
		"0.0.0-dev",
		server.WithToolCapabilities(false),
	)

	prompts.RegisterMCPPrompts(s)
	resources.RegisterMCPResources(s)
	tools.RegisterMCPTools(s)

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
