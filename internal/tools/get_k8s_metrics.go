package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/krmcbride/mcp-k8s/internal/k8s"
)

type getK8sMetricsParams struct {
	Context   string
	Kind      string
	Namespace string
	Name      string
	Sum       bool
}

// NodeMetrics represents CPU and memory usage for a node
type NodeMetrics struct {
	Name               string `json:"name"`
	CPUUsageMillicores int64  `json:"cpuUsageMillicores"`
	MemoryUsageMiB     int64  `json:"memoryUsageMiB"`
}

// PodMetrics represents CPU and memory usage for a pod
type PodMetrics struct {
	Name               string             `json:"name"`
	Namespace          string             `json:"namespace"`
	CPUUsageMillicores int64              `json:"cpuUsageMillicores"`
	MemoryUsageMiB     int64              `json:"memoryUsageMiB"`
	Containers         []ContainerMetrics `json:"containers"`
}

// ContainerMetrics represents CPU and memory usage for a container
type ContainerMetrics struct {
	Name               string `json:"name"`
	CPUUsageMillicores int64  `json:"cpuUsageMillicores"`
	MemoryUsageMiB     int64  `json:"memoryUsageMiB"`
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
		mcp.WithString(nameProperty,
			mcp.Description("Optional name to filter results by specific pod or node name."),
		),
		mcp.WithBoolean("sum",
			mcp.Description("When listing multiple resources, include a TOTAL entry with the sum of all CPU and memory usage."),
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
		content, err = getNodeMetrics(ctx, metricsClient, params.Name, params.Sum)
	} else {
		content, err = getPodMetrics(ctx, metricsClient, params.Namespace, params.Name, params.Sum)
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
		Name:      request.GetString(nameProperty, ""),
		Sum:       request.GetBool("sum", false),
	}, nil
}

func getNodeMetrics(ctx context.Context, metricsClient metrics.Interface, nodeName string, includeSum bool) ([]NodeMetrics, error) {
	if nodeName != "" {
		// Get specific node - sum not applicable for single item
		nodeMetric, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get node metrics for %s: %w", nodeName, err)
		}

		processed := processNodeMetric(nodeMetric)
		return []NodeMetrics{processed}, nil
	}

	// Get all nodes
	nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list node metrics: %w", err)
	}

	var nodeMetrics []NodeMetrics
	var totalCPUMillicores, totalMemoryMiB int64

	for _, nodeMetric := range nodeMetricsList.Items {
		processed := processNodeMetric(&nodeMetric)
		nodeMetrics = append(nodeMetrics, processed)

		// Add to totals
		totalCPUMillicores += processed.CPUUsageMillicores
		totalMemoryMiB += processed.MemoryUsageMiB
	}

	// Add total entry if requested
	if includeSum {
		nodeMetrics = append(nodeMetrics, NodeMetrics{
			Name:               "TOTAL",
			CPUUsageMillicores: totalCPUMillicores,
			MemoryUsageMiB:     totalMemoryMiB,
		})
	}

	return nodeMetrics, nil
}

func getPodMetrics(ctx context.Context, metricsClient metrics.Interface, namespace string, podName string, includeSum bool) ([]PodMetrics, error) {
	if podName != "" {
		// Get specific pod - sum not applicable for single item
		podMetric, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod metrics for %s: %w", podName, err)
		}

		processed := processPodMetric(podMetric)
		return []PodMetrics{processed}, nil
	}

	// Get metrics for all pods in the namespace(s)
	podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod metrics: %w", err)
	}

	podMetrics := make([]PodMetrics, 0, len(podMetricsList.Items))
	var totalCPUMillicores, totalMemoryMiB int64

	for _, podMetric := range podMetricsList.Items {
		processed := processPodMetric(&podMetric)
		podMetrics = append(podMetrics, processed)

		// Add to totals
		totalCPUMillicores += processed.CPUUsageMillicores
		totalMemoryMiB += processed.MemoryUsageMiB
	}

	// Add total entry if requested
	if includeSum {
		// Determine namespace for total - use "ALL" for cross-namespace queries
		totalNamespace := namespace
		if namespace == metav1.NamespaceAll {
			totalNamespace = "ALL"
		}

		podMetrics = append(podMetrics, PodMetrics{
			Name:               "TOTAL",
			Namespace:          totalNamespace,
			CPUUsageMillicores: totalCPUMillicores,
			MemoryUsageMiB:     totalMemoryMiB,
			Containers:         []ContainerMetrics{}, // Empty containers for total
		})
	}

	return podMetrics, nil
}

// Helper function to convert resource usage to standard units
func convertResourceUsage(usage corev1.ResourceList) (cpuMillicores int64, memoryMiB int64) {
	cpuQuantity := usage["cpu"]
	memoryQuantity := usage["memory"]

	cpuMillicores = cpuQuantity.MilliValue()
	memoryMiB = memoryQuantity.Value() / (1024 * 1024) // Convert bytes to MiB

	return cpuMillicores, memoryMiB
}

// Helper function to process a single node metric
func processNodeMetric(nodeMetric *metricsv1beta1.NodeMetrics) NodeMetrics {
	cpuUsageMillicores, memoryUsageMiB := convertResourceUsage(nodeMetric.Usage)

	return NodeMetrics{
		Name:               nodeMetric.Name,
		CPUUsageMillicores: cpuUsageMillicores,
		MemoryUsageMiB:     memoryUsageMiB,
	}
}

// Helper function to process a single pod metric
func processPodMetric(podMetric *metricsv1beta1.PodMetrics) PodMetrics {
	// Calculate total pod CPU and memory usage from all containers
	var totalCPUMillicores, totalMemoryMiB int64
	containers := make([]ContainerMetrics, 0, len(podMetric.Containers))

	for _, container := range podMetric.Containers {
		cpuUsageMillicores, memoryUsageMiB := convertResourceUsage(container.Usage)

		totalCPUMillicores += cpuUsageMillicores
		totalMemoryMiB += memoryUsageMiB

		containers = append(containers, ContainerMetrics{
			Name:               container.Name,
			CPUUsageMillicores: cpuUsageMillicores,
			MemoryUsageMiB:     memoryUsageMiB,
		})
	}

	return PodMetrics{
		Name:               podMetric.Name,
		Namespace:          podMetric.Namespace,
		CPUUsageMillicores: totalCPUMillicores,
		MemoryUsageMiB:     totalMemoryMiB,
		Containers:         containers,
	}
}
