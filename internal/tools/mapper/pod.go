package mapper

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PodListContent represents Pod-specific fields for list display
type PodListContent struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Status    string `json:"status,omitempty"`
	Ready     string `json:"ready,omitempty"`
	Restarts  int64  `json:"restarts,omitempty"`
	Age       string `json:"age,omitempty"`
}

func init() {
	// Register Pod mapper
	Register(
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		mapPodResource,
	)
}

func mapPodResource(item unstructured.Unstructured) interface{} {
	pod := PodListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}

	// Extract Pod-specific fields
	if status, found, _ := unstructured.NestedString(item.Object, "status", "phase"); found {
		pod.Status = status
	}

	// Extract container statuses for ready count and restarts
	if containers, found, _ := unstructured.NestedSlice(item.Object, "status", "containerStatuses"); found {
		ready := 0
		total := len(containers)
		restarts := int64(0)

		for _, c := range containers {
			if containerMap, ok := c.(map[string]interface{}); ok {
				if r, found, _ := unstructured.NestedBool(containerMap, "ready"); found && r {
					ready++
				}
				if rc, found, _ := unstructured.NestedInt64(containerMap, "restartCount"); found {
					restarts += rc
				}
			}
		}

		pod.Ready = fmt.Sprintf("%d/%d", ready, total)
		pod.Restarts = restarts
	}

	// TODO: Calculate age from creation timestamp

	return pod
}
