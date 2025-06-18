package tools

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/krmcbride/mcp-k8s/internal/tools/mapper"
)

func mapToK8sResourceListContent(list *unstructured.UnstructuredList, gvk schema.GroupVersionKind) []any {
	content := make([]any, 0, len(list.Items))

	// Get the appropriate mapper for this resource type
	resourceMapper, hasCustomMapper := mapper.Get(gvk)

	for _, item := range list.Items {
		if hasCustomMapper {
			// Use custom mapper
			content = append(content, resourceMapper(item))
		} else {
			// Fall back to generic mapper
			content = append(content, mapper.MapGenericK8sResource(item))
		}
	}
	return content
}

func mapToK8sResourceContent(resource *unstructured.Unstructured, gvk schema.GroupVersionKind) any {
	// Get the appropriate mapper for this resource type
	resourceMapper, hasCustomMapper := mapper.Get(gvk)

	if hasCustomMapper {
		// Use custom mapper
		return resourceMapper(*resource)
	} else {
		// Fall back to generic mapper
		return mapper.MapGenericK8sResource(*resource)
	}
}

func toJSONToolResult(content any) (*mcp.CallToolResult, error) {
	jsonContent, err := json.Marshal(content)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(jsonContent)), nil
}
