package mapper

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GenericListContent represents generic fields for any resource
type GenericListContent struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// MapGenericResource provides a fallback mapping for resources without custom mappers
func MapGenericResource(item unstructured.Unstructured) GenericListContent {
	return GenericListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}
}