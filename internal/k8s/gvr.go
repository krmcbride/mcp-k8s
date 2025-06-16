package k8s

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GVKToGVR converts a GroupVersionKind to a GroupVersionResource using the Kubernetes REST mapper.
// This is essential for dynamic client operations because the API uses different naming:
// - Kind: The type name used in YAML/JSON (e.g., "Pod", "Service", "Deployment")
// - Resource: The REST endpoint name (e.g., "pods", "services", "deployments")
//
// Parameters:
//   - context: The kubeconfig context to use for the REST mapper discovery
//   - gvk: The GroupVersionKind to convert (e.g., {Group: "", Version: "v1", Kind: "Pod"})
//
// Returns:
//   - The corresponding GroupVersionResource (e.g., {Group: "", Version: "v1", Resource: "pods"})
//   - An error if the mapping fails (e.g., unknown resource type)
//
// Example usage:
//
//	gvr, err := GVKToGVR("production", schema.GroupVersionKind{Version: "v1", Kind: "pod"})
//	// Returns: {Group: "", Version: "v1", Resource: "pods"}
func GVKToGVR(context string, gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	// Get K8s clients including REST mapper
	clients, err := getClientsForContext(context)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to create k8s clients: %w", err)
	}

	// Map Kind to Resource using REST mapper
	mapping, err := clients.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to map kind to resource: %w", err)
	}

	return mapping.Resource, nil
}
