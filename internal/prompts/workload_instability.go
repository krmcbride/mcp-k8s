package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterWorkloadInstabilityMCPPrompt(s *server.MCPServer) {
	s.AddPrompt(newWorkloadInstabilityMCPPrompt(), workloadInstabilityHandler)
}

// Prompt schema
func newWorkloadInstabilityMCPPrompt() mcp.Prompt {
	return mcp.NewPrompt("workload_instability_analysis",
		mcp.WithPromptDescription("Analyze Events and pod logs in a namespace for signs of workload instability. Provides a prioritized summary from most critical to least critical findings."),
		mcp.WithArgument("context",
			mcp.ArgumentDescription("The Kubernetes context to use for the analysis"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("namespace",
			mcp.ArgumentDescription("The namespace to analyze for workload instability"),
			mcp.RequiredArgument(),
		),
	)
}

// Prompt handler
func workloadInstabilityHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// Extract the required context argument
	k8sContext := request.Params.Arguments["context"]
	if k8sContext == "" {
		return nil, fmt.Errorf("context argument is required")
	}

	// Extract the required namespace argument
	namespace := request.Params.Arguments["namespace"]
	if namespace == "" {
		return nil, fmt.Errorf("namespace argument is required")
	}

	// Build the prompt content with the specified context and namespace
	promptContent := fmt.Sprintf(`Analyze Events and pod logs for signs of workload instability in namespace "%s".

Use Kubernetes context: %s
Target namespace: %s

<instructions>
PHASE 1: Event Analysis
1. Use list_k8s_resources tool to get all Events in the namespace:
   - context: %s
   - namespace: %s
   - kind: Event
   
2. Analyze Events for suspicious patterns:
   - Warning type events (especially recurring ones)
   - Failed operations (FailedMount, FailedScheduling, etc.)
   - Security-related events (Unauthorized, Forbidden, etc.)
   - Resource issues (OutOfMemory, DiskPressure, etc.)
   - Network problems (NetworkNotReady, DNSConfigForming, etc.)
   - Image pull failures (ErrImagePull, ImagePullBackOff, etc.)

PHASE 2: Pod Discovery and Log Analysis
1. Use list_k8s_resources tool to get all Pods in the namespace:
   - context: %s
   - namespace: %s
   - kind: Pod

2. For each pod (perform in parallel when possible):
   - Use get_k8s_pod_logs tool with tail=50 for recent logs
   - If multi-container pods, analyze logs from all containers
   - Look for suspicious patterns in logs:
     * ERROR, FATAL, PANIC level messages
     * Authentication/authorization failures
     * Connection timeouts or network errors
     * Resource exhaustion indicators
     * Application crashes or exceptions
     * Service degradation
     * Database connection failures
     * Certificate or TLS errors

PHASE 3: Analysis and Prioritization
Create a comprehensive summary organized by criticality:

## CRITICAL ISSUES (Immediate attention required)
- Service outages or complete failures
- Resource exhaustion causing instability
- Data corruption or loss indicators
- Persistent application crashes

## HIGH PRIORITY (Address soon)
- Network connectivity issues
- Recurring errors affecting functionality
- Performance degradation indicators
- Persistent restart loops

## MEDIUM PRIORITY (Monitor and plan)
- Warning-level events that may escalate
- Resource usage approaching limits
- Intermittent connection issues
- Configuration warnings

## LOW PRIORITY (Informational)
- Normal operational events
- Debug information
- Expected temporary conditions

For each finding, include:
- Source (Event or specific pod/container logs)
- Timestamp or frequency information
- Brief description of the issue
- Potential impact assessment
- Recommended actions where applicable

Focus on actionable insights and avoid including normal operational noise.
</instructions>`, namespace, k8sContext, namespace, k8sContext, namespace, k8sContext, namespace)

	return &mcp.GetPromptResult{
		Description: "Workload instability analysis prompt for Kubernetes namespace",
		Messages: []mcp.PromptMessage{
			{
				Role:    "user",
				Content: mcp.NewTextContent(promptContent),
			},
		},
	}, nil
}
