package tools

import (
	"context"
	"encoding/json"
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

type listResourcesParams struct {
	Context   string
	Namespace string
	Group     string
	Version   string
	Kind      string
}

func RegisterListResourcesTool(s *server.MCPServer) {
	s.AddTool(newListResourcesTool(), listResourcesHandler)
}

// Tool schema
func newListResourcesTool() mcp.Tool {
	return mcp.NewTool("list_resources",
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
func listResourcesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract and validate parameters
	params, err := extractListResourcesParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Convert GVK to GVR
	gvr, err := k8s.GVKToGVR(params.Context, schema.GroupVersionKind{
		Group:   params.Group,
		Version: params.Version,
		Kind:    params.Kind,
	})
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

	// Map to GenericListContent
	content := mapToGenericListContent(list)

	// Return as JSON
	return toJSONToolResult(content)
}

func extractListResourcesParams(request mcp.CallToolRequest) (*listResourcesParams, error) {
	context, err := request.RequireString(contextProperty)
	if err != nil {
		return nil, err
	}

	kind, err := request.RequireString(kindProperty)
	if err != nil {
		return nil, err
	}

	return &listResourcesParams{
		Context:   context,
		Namespace: request.GetString(namespaceProperty, metav1.NamespaceAll),
		Group:     request.GetString(groupProperty, ""),
		Version:   request.GetString(versionProperty, "v1"),
		Kind:      kind,
	}, nil
}

func mapToGenericListContent(list *unstructured.UnstructuredList) []GenericListContent {
	content := make([]GenericListContent, 0, len(list.Items))
	for _, item := range list.Items {
		content = append(content, GenericListContent{
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
		})
	}
	return content
}

func toJSONToolResult(content interface{}) (*mcp.CallToolResult, error) {
	jsonContent, err := json.Marshal(content)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(jsonContent)), nil
}
