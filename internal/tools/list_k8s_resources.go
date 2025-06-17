package tools

import (
	"context"
	"fmt"

	"github.com/krmcbride/mcp-k8s/internal/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	contextProperty   = "context"
	namespaceProperty = "namespace"
	groupProperty     = "group"
	versionProperty   = "version"
	kindProperty      = "kind"
)

type listK8sResourcesParams struct {
	Context   string
	Namespace string
	Group     string
	Version   string
	Kind      string
}

func RegisterListK8sResourcesMCPTool(s *server.MCPServer) {
	s.AddTool(newListK8sResourcesMCPTool(), listK8sResourcesHandler)
}

// Tool schema
func newListK8sResourcesMCPTool() mcp.Tool {
	return mcp.NewTool("list_k8s_resources",
		mcp.WithDescription("List Kubernetes resources"),
		mcp.WithString(contextProperty,
			mcp.Description("The Kubernetes context to use."),
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

	// List resources
	var list *unstructured.UnstructuredList
	if params.Namespace == metav1.NamespaceAll {
		list, err = dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list resources: %v", err)), nil
		}
	} else {
		list, err = dynamicClient.Resource(gvr).Namespace(params.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list resources: %v", err)), nil
		}
	}

	// Map to appropriate content structure
	content := mapToK8sResourceListContent(list, gvk)

	// Return as JSON
	return toJSONToolResult(content)
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

	return &listK8sResourcesParams{
		Context:   context,
		Namespace: request.GetString(namespaceProperty, metav1.NamespaceAll),
		Group:     request.GetString(groupProperty, ""),
		Version:   request.GetString(versionProperty, "v1"),
		Kind:      kind,
	}, nil
}
