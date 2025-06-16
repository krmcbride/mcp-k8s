package mapper

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestAllResourceMappersRegistered(t *testing.T) {
	// List of all resource types we expect to be registered
	expectedMappers := []schema.GroupVersionKind{
		{Group: "", Version: "v1", Kind: "Pod"},
		{Group: "apps", Version: "v1", Kind: "Deployment"},
		{Group: "apps", Version: "v1", Kind: "DaemonSet"},
		{Group: "apps", Version: "v1", Kind: "StatefulSet"},
		{Group: "", Version: "v1", Kind: "Service"},
		{Group: "batch", Version: "v1", Kind: "Job"},
		{Group: "batch", Version: "v1", Kind: "CronJob"},
		{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress"},
		{Group: "", Version: "v1", Kind: "Node"},
	}

	for _, gvk := range expectedMappers {
		t.Run(gvk.Kind, func(t *testing.T) {
			mapper, found := Get(gvk)
			if !found {
				t.Errorf("Expected to find mapper for %v, but didn't", gvk)
			}
			if mapper == nil {
				t.Errorf("Mapper for %v was nil", gvk)
			}
		})
	}

	// Verify we have the expected number of mappers
	expectedCount := len(expectedMappers)
	actualCount := len(resourceMappers)

	if actualCount != expectedCount {
		t.Errorf("Expected %d registered mappers, got %d", expectedCount, actualCount)

		// Debug: show what's actually registered
		t.Logf("Actually registered mappers:")
		for gvk := range resourceMappers {
			t.Logf("  %v", gvk)
		}
	}
}
