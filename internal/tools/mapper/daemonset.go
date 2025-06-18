package mapper

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DaemonSetListContent represents DaemonSet-specific fields for list display
type DaemonSetListContent struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Desired   int64  `json:"desired,omitempty"`
	Current   int64  `json:"current,omitempty"`
	Ready     int64  `json:"ready,omitempty"`
	UpToDate  int64  `json:"upToDate,omitempty"`
	Available int64  `json:"available,omitempty"`
	Age       string `json:"age,omitempty"`
}

func init() {
	// Register DaemonSet mapper
	Register(
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"},
		mapDaemonSetResource,
	)
}

func mapDaemonSetResource(item unstructured.Unstructured) any {
	daemonSet := DaemonSetListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}

	// Extract DaemonSet-specific fields from status
	if desired, found, _ := unstructured.NestedInt64(item.Object, "status", "desiredNumberScheduled"); found {
		daemonSet.Desired = desired
	}

	if current, found, _ := unstructured.NestedInt64(item.Object, "status", "currentNumberScheduled"); found {
		daemonSet.Current = current
	}

	if ready, found, _ := unstructured.NestedInt64(item.Object, "status", "numberReady"); found {
		daemonSet.Ready = ready
	}

	if upToDate, found, _ := unstructured.NestedInt64(item.Object, "status", "updatedNumberScheduled"); found {
		daemonSet.UpToDate = upToDate
	}

	if available, found, _ := unstructured.NestedInt64(item.Object, "status", "numberAvailable"); found {
		daemonSet.Available = available
	}

	// TODO: Calculate age from creation timestamp

	return daemonSet
}
