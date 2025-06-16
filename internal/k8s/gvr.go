package k8s

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GVKToGVR converts a GroupVersionKind to a GroupVersionResource using the REST mapper
func GVKToGVR(context string, gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	// Get K8s clients including REST mapper
	clients, err := getClientsForContext(context)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to create k8s clients: %w", err)
	}

	// Map Kind to Resource using REST mapper
	mapping, err := clients.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to map kind to resource: %w", err)
	}

	return mapping.Resource, nil
}
