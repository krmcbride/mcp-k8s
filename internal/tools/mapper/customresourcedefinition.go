package mapper

import (
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CustomResourceDefinitionListContent represents CRD-specific fields for list display
type CustomResourceDefinitionListContent struct {
	Name      string   `json:"name"`
	Group     string   `json:"group,omitempty"`
	Kind      string   `json:"kind,omitempty"`
	Scope     string   `json:"scope,omitempty"`
	Versions  []string `json:"versions,omitempty"`
	Age       string   `json:"age,omitempty"`
	Singular  string   `json:"singular,omitempty"`
	Plural    string   `json:"plural,omitempty"`
	ShortName string   `json:"shortName,omitempty"`
}

func init() {
	// Register both apiextensions.k8s.io/v1 and v1beta1 CRDs
	Register(schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}, mapCustomResourceDefinitionResource)
	Register(schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1beta1", Kind: "CustomResourceDefinition"}, mapCustomResourceDefinitionResource)
}

func mapCustomResourceDefinitionResource(item unstructured.Unstructured) any {
	content := CustomResourceDefinitionListContent{
		Name: item.GetName(),
		Age:  formatDuration(time.Since(item.GetCreationTimestamp().Time)),
	}

	// Extract group from spec.group
	if group, found, err := unstructured.NestedString(item.Object, "spec", "group"); err == nil && found {
		content.Group = group
	}

	// Extract scope from spec.scope
	if scope, found, err := unstructured.NestedString(item.Object, "spec", "scope"); err == nil && found {
		content.Scope = scope
	}

	// Extract names from spec.names
	if names, found, err := unstructured.NestedMap(item.Object, "spec", "names"); err == nil && found {
		if kind, ok := names["kind"].(string); ok {
			content.Kind = kind
		}
		if singular, ok := names["singular"].(string); ok {
			content.Singular = singular
		}
		if plural, ok := names["plural"].(string); ok {
			content.Plural = plural
		}

		// Extract short names (first one if multiple)
		if shortNames, ok := names["shortNames"].([]any); ok && len(shortNames) > 0 {
			if shortName, ok := shortNames[0].(string); ok {
				content.ShortName = shortName
			}
		}
	}

	// Extract versions from spec.versions
	if versions, found, err := unstructured.NestedSlice(item.Object, "spec", "versions"); err == nil && found {
		for _, v := range versions {
			if versionMap, ok := v.(map[string]any); ok {
				if name, ok := versionMap["name"].(string); ok {
					content.Versions = append(content.Versions, name)
				}
			}
		}
	}

	return content
}
