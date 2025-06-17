package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMemoryPressureMCPPrompt(s *server.MCPServer) {
	s.AddPrompt(newMemoryPressureMCPPrompt(), memoryPressureHandler)
}

// Prompt schema
func newMemoryPressureMCPPrompt() mcp.Prompt {
	return mcp.NewPrompt("memory_pressure_analysis",
		mcp.WithPromptDescription("Analyze pods for memory pressure issues including high usage, exceeding requests, and OOM kills. Requires a Kubernetes context to be specified."),
		mcp.WithArgument("context",
			mcp.ArgumentDescription("The Kubernetes context to use for the analysis"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("namespace",
			mcp.ArgumentDescription("The namespace to analyze (optional, defaults to all namespaces)"),
		),
	)
}

// Prompt handler
func memoryPressureHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// Extract the required context argument
	k8sContext := request.Params.Arguments["context"]
	if k8sContext == "" {
		return nil, fmt.Errorf("context argument is required")
	}

	// Extract the optional namespace argument
	namespace := request.Params.Arguments["namespace"]

	// Build the analysis scope description
	var scopeDescription string
	if namespace != "" {
		scopeDescription = fmt.Sprintf("Analyze namespace: %s", namespace)
	} else {
		scopeDescription = "Analyze all namespaces"
	}

	// Build the prompt content with the specified context and namespace
	promptContent := fmt.Sprintf(`Analyze pods for memory pressure issues. Check for:
1. Pods with memory usage close to their limits
2. Pods with memory usage significantly exceeding their requests
3. Pods that have been OOM killed

Use Kubernetes context: %s
%s

First, fetch pod metrics to analyze memory usage patterns.

<instructions>
1. Use the get_k8s_metrics tool to fetch current memory usage
2. Use the list_k8s_resources tool to get pod resource limits and requests
3. Look for pods where:
   - Memory usage is >80%% of the memory limit (high risk of OOM)
   - Memory usage is >120%% of the memory request (may cause node pressure)
   - Container status shows OOMKilled as a reason for termination
4. Summarize findings in a table showing:
   - Pod name and namespace
   - Memory usage (current/request/limit)
   - Usage percentage of limit
   - Usage percentage of request
   - OOM kill history if any
5. Highlight critical issues and provide recommendations
</instructions>`, k8sContext, scopeDescription)

	return &mcp.GetPromptResult{
		Description: "Memory pressure analysis prompt",
		Messages: []mcp.PromptMessage{
			{
				Role:    "user",
				Content: mcp.NewTextContent(promptContent),
			},
		},
	}, nil
}
