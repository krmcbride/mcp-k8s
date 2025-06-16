package mapper

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServiceListContent represents Service-specific fields for list display
type ServiceListContent struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace,omitempty"`
	Type       string   `json:"type,omitempty"`
	ClusterIP  string   `json:"clusterIP,omitempty"`
	ExternalIP []string `json:"externalIP,omitempty"`
	Port       string   `json:"port,omitempty"`
	Age        string   `json:"age,omitempty"`
}

func init() {
	// Register Service mapper
	Register(
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		mapServiceResource,
	)
}

func mapServiceResource(item unstructured.Unstructured) interface{} {
	service := ServiceListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}

	// Extract Service-specific fields from spec
	if serviceType, found, _ := unstructured.NestedString(item.Object, "spec", "type"); found {
		service.Type = serviceType
	}

	if clusterIP, found, _ := unstructured.NestedString(item.Object, "spec", "clusterIP"); found {
		service.ClusterIP = clusterIP
	}

	// Extract external IPs
	if externalIPs, found, _ := unstructured.NestedSlice(item.Object, "spec", "externalIPs"); found {
		for _, ip := range externalIPs {
			if ipStr, ok := ip.(string); ok {
				service.ExternalIP = append(service.ExternalIP, ipStr)
			}
		}
	}

	// Extract ports (simplified - just show first port)
	if ports, found, _ := unstructured.NestedSlice(item.Object, "spec", "ports"); found && len(ports) > 0 {
		if portMap, ok := ports[0].(map[string]interface{}); ok {
			if port, found, _ := unstructured.NestedInt64(portMap, "port"); found {
				if protocol, found, _ := unstructured.NestedString(portMap, "protocol"); found {
					service.Port = fmt.Sprintf("%d/%s", port, protocol)
				} else {
					service.Port = fmt.Sprintf("%d", port)
				}
			}
		}
	}

	// TODO: Calculate age from creation timestamp

	return service
}
