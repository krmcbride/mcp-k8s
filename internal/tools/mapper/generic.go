package mapper

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GenericK8sResourceContent represents generic fields for any resource
type GenericK8sResourceContent struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// MapGenericK8sResource provides a fallback mapping for resources without custom mappers
func MapGenericK8sResource(item unstructured.Unstructured) GenericK8sResourceContent {
	return GenericK8sResourceContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}
}
