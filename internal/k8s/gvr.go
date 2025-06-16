package k8s

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GVKToGVR converts a GroupVersionKind to a GroupVersionResource using the REST mapper
func GVKToGVR(context string, gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	// Normalize the Kind to ensure consistent casing
	normalizedGVK := gvk
	normalizedGVK.Kind = NormalizeKind(gvk.Kind)

	// Get K8s clients including REST mapper
	clients, err := getClientsForContext(context)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to create k8s clients: %w", err)
	}

	// Map Kind to Resource using REST mapper
	mapping, err := clients.RESTMapper.RESTMapping(normalizedGVK.GroupKind(), normalizedGVK.Version)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to map kind to resource: %w", err)
	}

	return mapping.Resource, nil
}

// NormalizeKind converts kind to title case (e.g., "pod" -> "Pod", "Pod" -> "Pod")
func NormalizeKind(kind string) string {
	if kind == "" {
		return kind
	}
	return strings.ToUpper(kind[:1]) + strings.ToLower(kind[1:])
}

// NormalizeGVK returns a new GVK with normalized Kind
func NormalizeGVK(gvk schema.GroupVersionKind) schema.GroupVersionKind {
	normalized := gvk
	normalized.Kind = NormalizeKind(gvk.Kind)
	return normalized
}
