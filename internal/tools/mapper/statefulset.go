package mapper

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// StatefulSetListContent represents StatefulSet-specific fields for list display
type StatefulSetListContent struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Ready     string `json:"ready,omitempty"`
	Age       string `json:"age,omitempty"`
}

func init() {
	// Register StatefulSet mapper
	Register(
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
		mapStatefulSetResource,
	)
}

func mapStatefulSetResource(item unstructured.Unstructured) interface{} {
	statefulSet := StatefulSetListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}

	// Extract StatefulSet-specific fields from status
	if replicas, found, _ := unstructured.NestedInt64(item.Object, "spec", "replicas"); found {
		if readyReplicas, found, _ := unstructured.NestedInt64(item.Object, "status", "readyReplicas"); found {
			statefulSet.Ready = fmt.Sprintf("%d/%d", readyReplicas, replicas)
		} else {
			statefulSet.Ready = fmt.Sprintf("0/%d", replicas)
		}
	}

	// TODO: Calculate age from creation timestamp

	return statefulSet
}
