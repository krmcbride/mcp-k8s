package mapper

import (
	"github.com/krmcbride/mcp-k8s/internal/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceMapper is a function that maps an unstructured item to a custom content structure
type ResourceMapper func(item unstructured.Unstructured) interface{}

// resourceMappers holds custom mappers for specific resource types
var resourceMappers = make(map[schema.GroupVersionKind]ResourceMapper)

// Register registers a custom mapper for a specific resource type
func Register(gvk schema.GroupVersionKind, mapper ResourceMapper) {
	resourceMappers[gvk] = mapper
}

// Get returns the appropriate mapper for a given GVK, handling normalization internally
func Get(gvk schema.GroupVersionKind) (ResourceMapper, bool) {
	// Normalize the GVK for mapper lookup
	normalizedGVK := k8s.NormalizeGVK(gvk)

	// Check if we have a custom mapper for this resource type
	mapper, hasCustomMapper := resourceMappers[normalizedGVK]
	return mapper, hasCustomMapper
}

// Init initializes all custom resource mappers
func Init() {
	// All resource mappers are automatically registered via init() functions
	// in their respective files (pod.go, deployment.go, etc.)
	// No explicit initialization needed
}
