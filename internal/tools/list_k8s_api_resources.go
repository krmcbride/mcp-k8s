package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/krmcbride/mcp-k8s/internal/k8s"
)

const (
	apiResourcesContextProperty = "context"
	apiResourcesGroupProperty   = "group"
)

type listK8sAPIResourcesParams struct {
	Context string
	Group   string
}

type APIResourceInfo struct {
	Name       string   `json:"name"`
	ShortNames []string `json:"shortNames,omitempty"`
	APIVersion string   `json:"apiVersion"`
	Namespaced bool     `json:"namespaced"`
	Kind       string   `json:"kind"`
}

func RegisterListK8sAPIResourcesMCPTool(s *server.MCPServer) {
	s.AddTool(newListK8sAPIResourcesMCPTool(), listK8sAPIResourcesHandler)
}

// Tool schema
func newListK8sAPIResourcesMCPTool() mcp.Tool {
	return mcp.NewTool("list_k8s_api_resources",
		mcp.WithDescription("List available Kubernetes API resources (equivalent to `kubectl api-resources`)"),
		mcp.WithString(apiResourcesContextProperty,
			mcp.Description("The Kubernetes context to use. To discover available contexts or resolve cluster aliases use the kubeconfig://contexts MCP resource."),
			mcp.Required(),
		),
		mcp.WithString(apiResourcesGroupProperty,
			mcp.Description("Filter by API group. If not specified, returns resources from all groups."),
		),
	)
}

// Tool handler
func listK8sAPIResourcesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract and validate parameters
	params, err := extractListK8sAPIResourcesParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get discovery client
	discoveryClient, err := k8s.GetDiscoveryClientForContext(params.Context)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create discovery client: %v", err)), nil
	}

	// Get all API resources - this can return partial results even with error
	_, resourceLists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		// Continue with partial results if any resource lists were discovered
		if len(resourceLists) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get API resources: %v", err)), nil
		}
	}

	var apiResources []APIResourceInfo

	// Process each resource list (each represents a group/version)
	for _, resourceList := range resourceLists {
		if resourceList == nil {
			continue
		}

		// Skip if group filter is specified and doesn't match
		if params.Group != "" && !matchesGroup(resourceList.GroupVersion, params.Group) {
			continue
		}

		// Process each resource in this group/version
		for _, resource := range resourceList.APIResources {
			// Skip subresources (those with '/' in the name)
			if strings.Contains(resource.Name, "/") {
				continue
			}

			apiResource := APIResourceInfo{
				Name:       resource.Name,
				ShortNames: resource.ShortNames,
				APIVersion: resourceList.GroupVersion,
				Namespaced: resource.Namespaced,
				Kind:       resource.Kind,
			}

			apiResources = append(apiResources, apiResource)
		}
	}

	// Return as JSON
	return toJSONToolResult(apiResources)
}

func extractListK8sAPIResourcesParams(request mcp.CallToolRequest) (*listK8sAPIResourcesParams, error) {
	context, err := request.RequireString(apiResourcesContextProperty)
	if err != nil {
		return nil, err
	}

	return &listK8sAPIResourcesParams{
		Context: context,
		Group:   request.GetString(apiResourcesGroupProperty, ""),
	}, nil
}

// matchesGroup checks if the groupVersion matches the specified group filter
func matchesGroup(groupVersion, groupFilter string) bool {
	// Handle core group (empty group)
	if groupFilter == "" || groupFilter == "core" {
		return groupVersion == "v1"
	}

	// Extract group from groupVersion (e.g., "apps/v1" -> "apps")
	parts := strings.Split(groupVersion, "/")
	if len(parts) == 1 {
		// Core API (just version like "v1")
		return false
	}

	group := parts[0]
	return group == groupFilter
}
