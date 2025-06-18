package tools

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/krmcbride/mcp-k8s/internal/k8s"
)

const (
	nameProperty       = "name"
	goTemplateProperty = "go_template"
)

type getK8sResourceParams struct {
	Context    string
	Name       string
	Namespace  string
	Group      string
	Version    string
	Kind       string
	GoTemplate string
}

func RegisterGetK8sResourceMCPTool(s *server.MCPServer) {
	s.AddTool(newGetK8sResourceMCPTool(), getK8sResourceHandler)
}

// Tool schema
func newGetK8sResourceMCPTool() mcp.Tool {
	return mcp.NewTool("get_k8s_resource",
		mcp.WithDescription("Get a single Kubernetes resource with optional Go template formatting"),
		mcp.WithString(contextProperty,
			mcp.Description("The Kubernetes context to use. To discover available contexts or resolve cluster aliases use the kubeconfig://contexts MCP resource."),
			mcp.Required(),
		),
		mcp.WithString(nameProperty,
			mcp.Description("The name of the resource to fetch."),
			mcp.Required(),
		),
		mcp.WithString(namespaceProperty,
			mcp.Description("The Kubernetes namespace to use. Required for namespaced resources."),
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
		mcp.WithString(goTemplateProperty,
			mcp.Description("Optional Go template expression for formatting output (e.g., '{{.metadata.name}}: {{.status.phase}}')."),
		),
	)
}

// Tool handler
func getK8sResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract and validate parameters
	params, err := extractGetK8sResourceParams(request)
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

	// Get the specific resource
	var resource *unstructured.Unstructured
	if params.Namespace == "" {
		// Cluster-scoped resource
		resource, err = dynamicClient.Resource(gvr).Get(ctx, params.Name, metav1.GetOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get resource: %v", err)), nil
		}
	} else {
		// Namespaced resource
		resource, err = dynamicClient.Resource(gvr).Namespace(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get resource: %v", err)), nil
		}
	}

	// Apply Go template if provided
	if params.GoTemplate != "" {
		return applyGoTemplate(resource, params.GoTemplate)
	}

	// Map to appropriate content structure using custom mappers
	content := mapToK8sResourceContent(resource, gvk)

	// Return as JSON
	return toJSONToolResult(content)
}

func extractGetK8sResourceParams(request mcp.CallToolRequest) (*getK8sResourceParams, error) {
	context, err := request.RequireString(contextProperty)
	if err != nil {
		return nil, err
	}

	name, err := request.RequireString(nameProperty)
	if err != nil {
		return nil, err
	}

	kind, err := request.RequireString(kindProperty)
	if err != nil {
		return nil, err
	}

	return &getK8sResourceParams{
		Context:    context,
		Name:       name,
		Namespace:  request.GetString(namespaceProperty, ""),
		Group:      request.GetString(groupProperty, ""),
		Version:    request.GetString(versionProperty, "v1"),
		Kind:       kind,
		GoTemplate: request.GetString(goTemplateProperty, ""),
	}, nil
}

func applyGoTemplate(resource *unstructured.Unstructured, templateStr string) (*mcp.CallToolResult, error) {
	// Parse the Go template
	tmpl, err := template.New("resource").Parse(templateStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse Go template: %v", err)), nil
	}

	// Apply the template to the resource
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, resource.Object)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to execute Go template: %v", err)), nil
	}

	// Return the template output as text
	return mcp.NewToolResultText(buf.String()), nil
}
