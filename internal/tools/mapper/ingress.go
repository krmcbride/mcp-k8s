package mapper

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IngressListContent represents Ingress-specific fields for list display
type IngressListContent struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace,omitempty"`
	Class     string   `json:"class,omitempty"`
	Hosts     []string `json:"hosts,omitempty"`
	Address   string   `json:"address,omitempty"`
	Ports     string   `json:"ports,omitempty"`
	Age       string   `json:"age,omitempty"`
}

func init() {
	// Register Ingress mapper
	Register(
		schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress"},
		mapIngressResource,
	)
}

func mapIngressResource(item unstructured.Unstructured) interface{} {
	ingress := IngressListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}

	// Extract Ingress class
	if ingressClass, found, _ := unstructured.NestedString(item.Object, "spec", "ingressClassName"); found {
		ingress.Class = ingressClass
	}

	// Extract hosts from rules
	if rules, found, _ := unstructured.NestedSlice(item.Object, "spec", "rules"); found {
		for _, rule := range rules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				if host, found, _ := unstructured.NestedString(ruleMap, "host"); found && host != "" {
					ingress.Hosts = append(ingress.Hosts, host)
				}
			}
		}
	}

	// Extract load balancer status
	if lbIngress, found, _ := unstructured.NestedSlice(item.Object, "status", "loadBalancer", "ingress"); found {
		var addresses []string
		for _, lb := range lbIngress {
			if lbMap, ok := lb.(map[string]interface{}); ok {
				if ip, found, _ := unstructured.NestedString(lbMap, "ip"); found && ip != "" {
					addresses = append(addresses, ip)
				}
				if hostname, found, _ := unstructured.NestedString(lbMap, "hostname"); found && hostname != "" {
					addresses = append(addresses, hostname)
				}
			}
		}
		if len(addresses) > 0 {
			ingress.Address = strings.Join(addresses, ",")
		}
	}

	// Default ports for Ingress
	ingress.Ports = "80,443"

	// TODO: Calculate age from creation timestamp

	return ingress
}
