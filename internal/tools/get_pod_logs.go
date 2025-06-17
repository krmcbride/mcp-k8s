package tools

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/krmcbride/mcp-k8s/internal/k8s"
)

type getPodLogsParams struct {
	Context   string
	Namespace string
	Name      string
	Container string
	Since     string
	SinceTime string
	Tail      int64
	Previous  bool
}

func RegisterGetK8sPodLogsMCPTool(s *server.MCPServer) {
	s.AddTool(newGetK8sPodLogsMCPTool(), getK8sPodLogsHandler)
}

// Tool schema
func newGetK8sPodLogsMCPTool() mcp.Tool {
	return mcp.NewTool("get_k8s_pod_logs",
		mcp.WithDescription("Get logs from a Kubernetes pod, similar to kubectl logs"),
		mcp.WithString(contextProperty,
			mcp.Description("The Kubernetes context to use."),
			mcp.Required(),
		),
		mcp.WithString(namespaceProperty,
			mcp.Description("The Kubernetes namespace of the pod."),
			mcp.Required(),
		),
		mcp.WithString(nameProperty,
			mcp.Description("The name of the pod to get logs from."),
			mcp.Required(),
		),
		mcp.WithString("container",
			mcp.Description("Optional container name. If not specified, uses the first container."),
		),
		mcp.WithString("since",
			mcp.Description("Return logs since a relative time (e.g., '5m', '1h', '30s'). Cannot be used with sinceTime."),
		),
		mcp.WithString("sinceTime",
			mcp.Description("Return logs since an RFC3339 timestamp. Cannot be used with since."),
		),
		mcp.WithNumber("tail",
			mcp.Description("Number of lines to return from the end of the log. Defaults to 10."),
		),
		mcp.WithBoolean("previous",
			mcp.Description("Return logs from the previous terminated container instance."),
		),
	)
}

// Tool handler
func getK8sPodLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract and validate parameters
	params, err := extractGetK8sPodLogsParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Validate mutual exclusion of since and sinceTime
	if params.Since != "" && params.SinceTime != "" {
		return mcp.NewToolResultError("cannot specify both 'since' and 'sinceTime' parameters"), nil
	}

	// Get Kubernetes clientset for pod logs
	clientset, err := k8s.GetClientsetForContext(params.Context)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create Kubernetes clientset: %v", err)), nil
	}

	// Build log options
	logOptions := &corev1.PodLogOptions{
		Previous: params.Previous,
	}

	if params.Container != "" {
		logOptions.Container = params.Container
	}

	if params.Tail > 0 {
		logOptions.TailLines = &params.Tail
	}

	// Handle since/sinceTime
	if params.Since != "" {
		duration, err := parseDuration(params.Since)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid 'since' duration: %v", err)), nil
		}
		logOptions.SinceSeconds = &duration
	} else if params.SinceTime != "" {
		sinceTime, err := time.Parse(time.RFC3339, params.SinceTime)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid 'sinceTime' format (expected RFC3339): %v", err)), nil
		}
		metaTime := metav1.NewTime(sinceTime)
		logOptions.SinceTime = &metaTime
	}

	// Get pod logs
	req := clientset.CoreV1().Pods(params.Namespace).GetLogs(params.Name, logOptions)
	logs, err := req.Stream(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get pod logs: %v", err)), nil
	}
	defer func() {
		_ = logs.Close() // Ignore close error
	}()

	// Read logs
	logData, err := io.ReadAll(logs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read pod logs: %v", err)), nil
	}

	// Return logs as text
	return mcp.NewToolResultText(string(logData)), nil
}

func extractGetK8sPodLogsParams(request mcp.CallToolRequest) (*getPodLogsParams, error) {
	context, err := request.RequireString(contextProperty)
	if err != nil {
		return nil, err
	}

	namespace, err := request.RequireString(namespaceProperty)
	if err != nil {
		return nil, err
	}

	name, err := request.RequireString(nameProperty)
	if err != nil {
		return nil, err
	}

	// Handle tail parameter - default to 10
	tail := int64(request.GetInt("tail", 10))

	return &getPodLogsParams{
		Context:   context,
		Namespace: namespace,
		Name:      name,
		Container: request.GetString("container", ""),
		Since:     request.GetString("since", ""),
		SinceTime: request.GetString("sinceTime", ""),
		Tail:      tail,
		Previous:  request.GetBool("previous", false),
	}, nil
}

// parseDuration converts duration strings like "5m", "1h", "30s" to seconds
func parseDuration(durationStr string) (int64, error) {
	if durationStr == "" {
		return 0, nil
	}

	// Parse using Go's time.ParseDuration
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %w", err)
	}

	return int64(duration.Seconds()), nil
}
