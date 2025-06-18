// Package mapper provides an extensible system for converting Kubernetes unstructured
// resources into structured output with resource-specific field extraction and
// case-insensitive Kind lookups.
package mapper

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceMapper is a function that maps an unstructured item to a custom content structure
type ResourceMapper func(item unstructured.Unstructured) any

// resourceMappers holds custom mappers for specific resource types
var resourceMappers = make(map[schema.GroupVersionKind]ResourceMapper)

// Register registers a custom mapper for a specific resource type.
// The GVK is normalized to ensure consistent map keys.
func Register(gvk schema.GroupVersionKind, mapper ResourceMapper) {
	// Normalize the GVK to ensure consistent keys
	normalizedGVK := normalizeGVKForLookup(gvk)
	resourceMappers[normalizedGVK] = mapper
}

// Get returns the appropriate mapper for a given GVK, handling normalization internally
func Get(gvk schema.GroupVersionKind) (ResourceMapper, bool) {
	// Normalize the GVK for our internal registry lookup
	normalizedGVK := normalizeGVKForLookup(gvk)

	// Check if we have a custom mapper for this resource type
	mapper, hasCustomMapper := resourceMappers[normalizedGVK]
	return mapper, hasCustomMapper
}

// normalizeGVKForLookup ensures consistent keys for our mapper registry.
// This normalization is applied during both registration and lookup to ensure
// that keys always match regardless of the casing used.
//
// This is NOT about Kubernetes API requirements - the k8s REST mapper handles
// various cases just fine. This is purely for our internal map key consistency.
//
// Examples:
//
//	"pod" -> "Pod"
//	"POD" -> "Pod"
//	"Pod" -> "Pod"
//	"configmap" -> "Configmap"
//	"ConfigMap" -> "Configmap"
//
// Note: This simple title-case approach means "ConfigMap" becomes "Configmap",
// but this doesn't affect functionality since the k8s REST mapper handles the
// actual Kubernetes API calls with proper names.
func normalizeGVKForLookup(gvk schema.GroupVersionKind) schema.GroupVersionKind {
	normalized := gvk
	if gvk.Kind != "" {
		normalized.Kind = strings.ToUpper(gvk.Kind[:1]) + strings.ToLower(gvk.Kind[1:])
	}
	return normalized
}

// Init initializes all custom resource mappers
func Init() {
	// All resource mappers are automatically registered via init() functions
	// in their respective files (pod.go, deployment.go, etc.)
	// No explicit initialization needed
}
