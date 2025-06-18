package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeContext represents a Kubernetes context with its associated cluster information
type KubeContext struct {
	Name        string `json:"name"`
	ClusterName string `json:"clusterName"`
	IsCurrent   bool   `json:"isCurrent"`
}

func RegisterK8sContextsMCPResource(s *server.MCPServer) {
	s.AddResource(newK8sContextsMCPResource(), k8sContextsHandler)
}

// Resource schema
func newK8sContextsMCPResource() mcp.Resource {
	return mcp.NewResource("kubeconfig://contexts", "kubeconfig_contexts",
		mcp.WithResourceDescription("Current user's kubeconfig contexts - maps context names to cluster names for "+
			"resolving cluster aliases like 'prod' or 'sandbox' to actual cluster names and context names. Use this "+
			"resource to discover available Kubernetes contexts instead of running `kubectl config`."),
		mcp.WithMIMEType("application/json"),
	)
}

// Resource handler
func k8sContextsHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Load kubeconfig using the same rules as our k8s client
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := loadingRules.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Get the current context
	currentContext := config.CurrentContext

	// Build list of contexts with their cluster names
	contexts := make([]KubeContext, 0, len(config.Contexts))
	for name, context := range config.Contexts {
		contexts = append(contexts, KubeContext{
			Name:        name,
			ClusterName: context.Cluster,
			IsCurrent:   name == currentContext,
		})
	}

	// Convert to JSON
	jsonData, err := json.Marshal(contexts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contexts: %w", err)
	}

	// Return as MCP resource contents
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "kubeconfig://contexts",
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}
