package mapper

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NodeListContent represents Node-specific fields for list display
type NodeListContent struct {
	Name             string   `json:"name"`
	Status           string   `json:"status,omitempty"`
	Roles            []string `json:"roles,omitempty"`
	Age              string   `json:"age,omitempty"`
	Version          string   `json:"version,omitempty"`
	InternalIP       string   `json:"internalIP,omitempty"`
	ExternalIP       string   `json:"externalIP,omitempty"`
	OSImage          string   `json:"osImage,omitempty"`
	KernelVersion    string   `json:"kernelVersion,omitempty"`
	ContainerRuntime string   `json:"containerRuntime,omitempty"`
}

func init() {
	// Register Node mapper
	Register(
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"},
		mapNodeResource,
	)
}

func mapNodeResource(item unstructured.Unstructured) any {
	node := NodeListContent{
		Name: item.GetName(),
		// Nodes don't have namespaces
	}

	// Extract node status
	if conditions, found, _ := unstructured.NestedSlice(item.Object, "status", "conditions"); found {
		for _, condition := range conditions {
			if condMap, ok := condition.(map[string]any); ok {
				if condType, found, _ := unstructured.NestedString(condMap, "type"); found && condType == "Ready" {
					if status, found, _ := unstructured.NestedString(condMap, "status"); found {
						if status == "True" {
							node.Status = "Ready"
						} else {
							node.Status = "NotReady"
						}
					}
				}
			}
		}
	}

	// Extract roles from labels
	if labels := item.GetLabels(); labels != nil {
		var roles []string
		for key := range labels {
			if strings.HasPrefix(key, "node-role.kubernetes.io/") {
				role := strings.TrimPrefix(key, "node-role.kubernetes.io/")
				if role != "" {
					roles = append(roles, role)
				}
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "<none>")
		}
		node.Roles = roles
	}

	// Extract kubelet version
	if version, found, _ := unstructured.NestedString(item.Object, "status", "nodeInfo", "kubeletVersion"); found {
		node.Version = version
	}

	// Extract IP addresses
	if addresses, found, _ := unstructured.NestedSlice(item.Object, "status", "addresses"); found {
		for _, address := range addresses {
			if addrMap, ok := address.(map[string]any); ok {
				if addrType, found, _ := unstructured.NestedString(addrMap, "type"); found {
					if addr, found, _ := unstructured.NestedString(addrMap, "address"); found {
						switch addrType {
						case "InternalIP":
							node.InternalIP = addr
						case "ExternalIP":
							node.ExternalIP = addr
						}
					}
				}
			}
		}
	}

	// Extract system info
	if osImage, found, _ := unstructured.NestedString(item.Object, "status", "nodeInfo", "osImage"); found {
		node.OSImage = osImage
	}

	if kernelVersion, found, _ := unstructured.NestedString(item.Object, "status", "nodeInfo", "kernelVersion"); found {
		node.KernelVersion = kernelVersion
	}

	if containerRuntime, found, _ := unstructured.NestedString(item.Object, "status", "nodeInfo", "containerRuntimeVersion"); found {
		node.ContainerRuntime = containerRuntime
	}

	// TODO: Calculate age from creation timestamp

	return node
}
