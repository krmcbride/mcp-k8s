package mapper

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DeploymentListContent represents Deployment-specific fields for list display
type DeploymentListContent struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Ready     string `json:"ready,omitempty"`
	UpToDate  int64  `json:"upToDate,omitempty"`
	Available int64  `json:"available,omitempty"`
	Age       string `json:"age,omitempty"`
}

func init() {
	// Register Deployment mapper
	Register(
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		mapDeploymentResource,
	)
}

func mapDeploymentResource(item unstructured.Unstructured) any {
	deployment := DeploymentListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}

	// Extract Deployment-specific fields from status
	if replicas, found, _ := unstructured.NestedInt64(item.Object, "status", "replicas"); found {
		if readyReplicas, found, _ := unstructured.NestedInt64(item.Object, "status", "readyReplicas"); found {
			deployment.Ready = fmt.Sprintf("%d/%d", readyReplicas, replicas)
		} else {
			deployment.Ready = fmt.Sprintf("0/%d", replicas)
		}
	}

	if upToDate, found, _ := unstructured.NestedInt64(item.Object, "status", "updatedReplicas"); found {
		deployment.UpToDate = upToDate
	}

	if available, found, _ := unstructured.NestedInt64(item.Object, "status", "availableReplicas"); found {
		deployment.Available = available
	}

	// TODO: Calculate age from creation timestamp

	return deployment
}
