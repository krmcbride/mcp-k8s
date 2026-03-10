package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestToolsDeclareReadOnlyAnnotations(t *testing.T) {
	tests := []struct {
		name string
		tool mcp.Tool
	}{
		{name: "list_k8s_resources", tool: newListK8sResourcesMCPTool()},
		{name: "list_k8s_api_resources", tool: newListK8sAPIResourcesMCPTool()},
		{name: "get_k8s_resource", tool: newGetK8sResourceMCPTool()},
		{name: "get_k8s_metrics", tool: newGetK8sMetricsMCPTool()},
		{name: "get_k8s_pod_logs", tool: newGetK8sPodLogsMCPTool()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertBoolPtrValue(t, tt.tool.Annotations.ReadOnlyHint, true, "readOnlyHint")
			assertBoolPtrValue(t, tt.tool.Annotations.DestructiveHint, false, "destructiveHint")
			assertBoolPtrValue(t, tt.tool.Annotations.IdempotentHint, true, "idempotentHint")
			assertBoolPtrValue(t, tt.tool.Annotations.OpenWorldHint, true, "openWorldHint")
		})
	}
}

func assertBoolPtrValue(t *testing.T, value *bool, want bool, field string) {
	t.Helper()

	if value == nil {
		t.Fatalf("%s annotation was nil", field)
	}

	if *value != want {
		t.Fatalf("%s = %t, want %t", field, *value, want)
	}
}
