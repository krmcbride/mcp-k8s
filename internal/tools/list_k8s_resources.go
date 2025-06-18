package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/krmcbride/mcp-k8s/internal/k8s"
)

const (
	contextProperty       = "context"
	namespaceProperty     = "namespace"
	groupProperty         = "group"
	versionProperty       = "version"
	kindProperty          = "kind"
	fieldSelectorProperty = "fieldSelector"
	limitProperty         = "limit"
	continueProperty      = "continue"
)

type listK8sResourcesParams struct {
	Context       string
	Namespace     string
	Group         string
	Version       string
	Kind          string
	FieldSelector string
	Limit         int64
	Continue      string
}

func RegisterListK8sResourcesMCPTool(s *server.MCPServer) {
	s.AddTool(newListK8sResourcesMCPTool(), listK8sResourcesHandler)
}

// Tool schema
func newListK8sResourcesMCPTool() mcp.Tool {
	return mcp.NewTool("list_k8s_resources",
		mcp.WithDescription("List Kubernetes resources with optional server-side filtering and pagination"),
		mcp.WithString(contextProperty,
			mcp.Description("The Kubernetes context to use. To discover available contexts or resolve cluster aliases use the kubeconfig://contexts MCP resource."),
			mcp.Required(),
		),
		mcp.WithString(namespaceProperty,
			mcp.Description("The Kubernetes namespace to use. Defaults to all namespaces."),
		),
		mcp.WithString(groupProperty,
			mcp.Description("The Kubernetes resource API Group."),
		),
		mcp.WithString(versionProperty,
			mcp.Description("The Kubernetes resource API Version."),
		),
		mcp.WithString(kindProperty,
			mcp.Description("The Kubernetes resource Kind."),
			mcp.Required(),
		),
		mcp.WithString(fieldSelectorProperty,
			mcp.Description("Field selector to filter resources server-side. Examples: 'metadata.namespace!=default', 'status.phase=Running', 'spec.nodeName=node-1'. Multiple selectors can be comma-separated."),
		),
		// NOTE: The Event mapper, which contains a good number of fields, is about 120 tokens per event, so a default
		// limit of 100 uses about half of the 25k MCP tool response token limit
		mcp.WithNumber(limitProperty,
			mcp.Description("Maximum number of resources to return per request. Use for pagination. Must be positive if provided. Defaults to 100."),
		),
		mcp.WithString(continueProperty,
			mcp.Description("Continue token from previous paginated request. Used to retrieve the next page of results."),
		),
	)
}

// Tool handler
func listK8sResourcesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract and validate parameters
	params, err := extractListK8sResourcesParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Create GVK
	gvk := schema.GroupVersionKind{
		Group:   params.Group,
		Version: params.Version,
		Kind:    params.Kind,
	}

	// Convert GVK to GVR
	gvr, err := k8s.GVKToGVR(params.Context, gvk)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get dynamic client
	dynamicClient, err := k8s.GetDynamicClientForContext(params.Context)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create dynamic client: %v", err)), nil
	}

	// Prepare list options with field selector and pagination
	listOptions := metav1.ListOptions{
		Limit: params.Limit, // Always set limit (defaults to 100)
	}
	if params.FieldSelector != "" {
		listOptions.FieldSelector = params.FieldSelector
	}
	if params.Continue != "" {
		listOptions.Continue = params.Continue
	}

	// List resources
	var list *unstructured.UnstructuredList
	if params.Namespace == metav1.NamespaceAll {
		list, err = dynamicClient.Resource(gvr).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list resources: %v", err)), nil
		}
	} else {
		list, err = dynamicClient.Resource(gvr).Namespace(params.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list resources: %v", err)), nil
		}
	}

	// Map to appropriate content structure
	items := mapToK8sResourceListContent(list, gvk)

	// Create response with pagination metadata
	response := map[string]any{
		"items": items,
	}

	// Add pagination metadata if available
	metadata := map[string]any{}
	hasMetadata := false

	// Extract continue token from list metadata
	if continueToken, found, _ := unstructured.NestedString(list.Object, "metadata", "continue"); found && continueToken != "" {
		metadata["continue"] = continueToken
		hasMetadata = true
	}

	// Extract remaining item count from list metadata
	if remainingCount, found, _ := unstructured.NestedInt64(list.Object, "metadata", "remainingItemCount"); found {
		metadata["remainingItemCount"] = remainingCount
		hasMetadata = true
	}

	if hasMetadata {
		response["metadata"] = metadata
	}

	// Return as JSON
	return toJSONToolResult(response)
}

func extractListK8sResourcesParams(request mcp.CallToolRequest) (*listK8sResourcesParams, error) {
	context, err := request.RequireString(contextProperty)
	if err != nil {
		return nil, err
	}

	kind, err := request.RequireString(kindProperty)
	if err != nil {
		return nil, err
	}

	// Extract and validate limit (default to 100)
	limit := request.GetFloat(limitProperty, 100)
	if limit < 0 {
		return nil, fmt.Errorf("limit must be positive, got %v", limit)
	}

	return &listK8sResourcesParams{
		Context:       context,
		Namespace:     request.GetString(namespaceProperty, metav1.NamespaceAll),
		Group:         request.GetString(groupProperty, ""),
		Version:       request.GetString(versionProperty, "v1"),
		Kind:          kind,
		FieldSelector: request.GetString(fieldSelectorProperty, ""),
		Limit:         int64(limit),
		Continue:      request.GetString(continueProperty, ""),
	}, nil
}
