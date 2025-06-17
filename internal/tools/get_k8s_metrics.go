package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/krmcbride/mcp-k8s/internal/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type getK8sMetricsParams struct {
	Context   string
	Kind      string
	Namespace string
}

// NodeMetrics represents CPU and memory usage for a node
type NodeMetrics struct {
	Name        string `json:"name"`
	CPUUsage    string `json:"cpuUsage"`
	MemoryUsage string `json:"memoryUsage"`
}

// PodMetrics represents CPU and memory usage for a pod
type PodMetrics struct {
	Name        string             `json:"name"`
	Namespace   string             `json:"namespace"`
	CPUUsage    string             `json:"cpuUsage"`
	MemoryUsage string             `json:"memoryUsage"`
	Containers  []ContainerMetrics `json:"containers"`
}

// ContainerMetrics represents CPU and memory usage for a container
type ContainerMetrics struct {
	Name        string `json:"name"`
	CPUUsage    string `json:"cpuUsage"`
	MemoryUsage string `json:"memoryUsage"`
}

func RegisterGetK8sMetricsMCPTool(s *server.MCPServer) {
	s.AddTool(newGetK8sMetricsMCPTool(), getK8sMetricsHandler)
}

// Tool schema
func newGetK8sMetricsMCPTool() mcp.Tool {
	return mcp.NewTool("get_k8s_metrics",
		mcp.WithDescription("Get Kubernetes resource metrics (CPU/memory usage) for nodes or pods, similar to kubectl top"),
		mcp.WithString(contextProperty,
			mcp.Description("The Kubernetes context to use."),
			mcp.Required(),
		),
		mcp.WithString(kindProperty,
			mcp.Description("The resource type to get metrics for. Must be 'node' or 'pod'."),
			mcp.Required(),
		),
		mcp.WithString(namespaceProperty,
			mcp.Description("The Kubernetes namespace to use. Ignored for nodes. If not provided for pods, shows metrics for all namespaces."),
		),
	)
}

// Tool handler
func getK8sMetricsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract and validate parameters
	params, err := extractGetK8sMetricsParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Validate kind parameter
	if params.Kind != "node" && params.Kind != "pod" {
		return mcp.NewToolResultError("kind must be 'node' or 'pod'"), nil
	}

	// Get metrics client
	metricsClient, err := k8s.GetMetricsClientForContext(params.Context)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create metrics client: %v", err)), nil
	}

	// Get metrics based on kind
	var content interface{}
	if params.Kind == "node" {
		content, err = getNodeMetrics(ctx, metricsClient)
	} else {
		content, err = getPodMetrics(ctx, metricsClient, params.Namespace)
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get %s metrics: %v", params.Kind, err)), nil
	}

	// Return as JSON
	return toJSONToolResult(content)
}

func extractGetK8sMetricsParams(request mcp.CallToolRequest) (*getK8sMetricsParams, error) {
	context, err := request.RequireString(contextProperty)
	if err != nil {
		return nil, err
	}

	kind, err := request.RequireString(kindProperty)
	if err != nil {
		return nil, err
	}

	// Normalize kind to lowercase for consistency
	kind = strings.ToLower(kind)

	return &getK8sMetricsParams{
		Context:   context,
		Kind:      kind,
		Namespace: request.GetString(namespaceProperty, metav1.NamespaceAll),
	}, nil
}

func getNodeMetrics(ctx context.Context, metricsClient metrics.Interface) ([]NodeMetrics, error) {
	nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list node metrics: %w", err)
	}

	nodeMetrics := make([]NodeMetrics, 0, len(nodeMetricsList.Items))

	for _, nodeMetric := range nodeMetricsList.Items {
		cpuUsage := formatResourceQuantity(nodeMetric.Usage["cpu"])
		memoryUsage := formatResourceQuantity(nodeMetric.Usage["memory"])

		nodeMetrics = append(nodeMetrics, NodeMetrics{
			Name:        nodeMetric.Name,
			CPUUsage:    cpuUsage,
			MemoryUsage: memoryUsage,
		})
	}

	return nodeMetrics, nil
}

func getPodMetrics(ctx context.Context, metricsClient metrics.Interface, namespace string) ([]PodMetrics, error) {
	podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod metrics: %w", err)
	}

	podMetrics := make([]PodMetrics, 0, len(podMetricsList.Items))

	for _, podMetric := range podMetricsList.Items {
		// Calculate total pod CPU and memory usage from all containers
		var totalCPU, totalMemory resource.Quantity
		containers := make([]ContainerMetrics, 0, len(podMetric.Containers))

		for _, container := range podMetric.Containers {
			cpuUsage := container.Usage["cpu"]
			memoryUsage := container.Usage["memory"]

			totalCPU.Add(cpuUsage)
			totalMemory.Add(memoryUsage)

			containers = append(containers, ContainerMetrics{
				Name:        container.Name,
				CPUUsage:    formatResourceQuantity(cpuUsage),
				MemoryUsage: formatResourceQuantity(memoryUsage),
			})
		}

		podMetrics = append(podMetrics, PodMetrics{
			Name:        podMetric.Name,
			Namespace:   podMetric.Namespace,
			CPUUsage:    formatResourceQuantity(totalCPU),
			MemoryUsage: formatResourceQuantity(totalMemory),
			Containers:  containers,
		})
	}

	return podMetrics, nil
}

// Helper function to format resource quantities in human-readable format
func formatResourceQuantity(q resource.Quantity) string {
	// Convert to millicores for CPU or bytes for memory
	if q.Format == resource.DecimalSI {
		// CPU in millicores
		return fmt.Sprintf("%dm", q.MilliValue())
	}
	// Memory in bytes, convert to Mi
	bytes := q.Value()
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%dMi", bytes/(1024*1024))
	}
	return fmt.Sprintf("%d", bytes)
}
