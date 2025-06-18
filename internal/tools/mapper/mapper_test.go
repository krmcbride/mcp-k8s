package mapper

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Mock mapper for testing
func mockMapper(item unstructured.Unstructured) any {
	return "mocked"
}

func TestMapperRegistrationAndLookup(t *testing.T) {
	// Clear any existing mappers before testing
	resourceMappers = make(map[schema.GroupVersionKind]ResourceMapper)

	tests := []struct {
		name        string
		registerGVK schema.GroupVersionKind
		lookupGVKs  []schema.GroupVersionKind
		expectFound bool
		description string
	}{
		{
			name: "exact match - lowercase",
			registerGVK: schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "pod",
			},
			lookupGVKs: []schema.GroupVersionKind{
				{Group: "", Version: "v1", Kind: "pod"},
				{Group: "", Version: "v1", Kind: "Pod"},
				{Group: "", Version: "v1", Kind: "POD"},
			},
			expectFound: true,
			description: "All case variations should find the mapper",
		},
		{
			name: "exact match - uppercase",
			registerGVK: schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "POD",
			},
			lookupGVKs: []schema.GroupVersionKind{
				{Group: "", Version: "v1", Kind: "pod"},
				{Group: "", Version: "v1", Kind: "Pod"},
				{Group: "", Version: "v1", Kind: "POD"},
			},
			expectFound: true,
			description: "Registration with uppercase should work with all lookups",
		},
		{
			name: "multi-word kinds",
			registerGVK: schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "ConfigMap",
			},
			lookupGVKs: []schema.GroupVersionKind{
				{Group: "", Version: "v1", Kind: "configmap"},
				{Group: "", Version: "v1", Kind: "ConfigMap"},
				{Group: "", Version: "v1", Kind: "CONFIGMAP"},
				{Group: "", Version: "v1", Kind: "configMap"},
			},
			expectFound: true,
			description: "Multi-word kinds should normalize consistently",
		},
		{
			name: "exact match with different cases",
			registerGVK: schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "deployment",
			},
			lookupGVKs: []schema.GroupVersionKind{
				{Group: "apps", Version: "v1", Kind: "Deployment"}, // Should match
				{Group: "apps", Version: "v1", Kind: "DEPLOYMENT"}, // Should match
			},
			expectFound: true,
			description: "Same group/version with different Kind cases should match",
		},
		{
			name: "different group or version",
			registerGVK: schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "deployment",
			},
			lookupGVKs: []schema.GroupVersionKind{
				{Group: "", Version: "v1", Kind: "Deployment"},     // Different group
				{Group: "apps", Version: "v2", Kind: "Deployment"}, // Different version
			},
			expectFound: false,
			description: "Different group or version should not match",
		},
		{
			name: "empty kind",
			registerGVK: schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "",
			},
			lookupGVKs: []schema.GroupVersionKind{
				{Group: "", Version: "v1", Kind: ""},
			},
			expectFound: true,
			description: "Empty kinds should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear mappers for each test
			resourceMappers = make(map[schema.GroupVersionKind]ResourceMapper)

			// Register the mapper
			Register(tt.registerGVK, mockMapper)

			// Test lookups
			for _, lookupGVK := range tt.lookupGVKs {
				mapper, found := Get(lookupGVK)

				if tt.expectFound {
					if !found {
						t.Errorf("Expected to find mapper for %v after registering %v, but didn't",
							lookupGVK, tt.registerGVK)
					}
					if mapper == nil {
						t.Errorf("Found mapper but it was nil for %v", lookupGVK)
					}
				} else {
					if found {
						t.Errorf("Expected NOT to find mapper for %v after registering %v, but did",
							lookupGVK, tt.registerGVK)
					}
				}
			}
		})
	}
}

func TestNormalizeGVKForLookup(t *testing.T) {
	tests := []struct {
		input    schema.GroupVersionKind
		expected schema.GroupVersionKind
		name     string
	}{
		{
			name: "lowercase to titlecase",
			input: schema.GroupVersionKind{
				Group: "apps", Version: "v1", Kind: "pod",
			},
			expected: schema.GroupVersionKind{
				Group: "apps", Version: "v1", Kind: "Pod",
			},
		},
		{
			name: "uppercase to titlecase",
			input: schema.GroupVersionKind{
				Group: "apps", Version: "v1", Kind: "POD",
			},
			expected: schema.GroupVersionKind{
				Group: "apps", Version: "v1", Kind: "Pod",
			},
		},
		{
			name: "mixed case to titlecase",
			input: schema.GroupVersionKind{
				Group: "apps", Version: "v1", Kind: "DePlOyMeNt",
			},
			expected: schema.GroupVersionKind{
				Group: "apps", Version: "v1", Kind: "Deployment",
			},
		},
		{
			name: "multi-word flattened",
			input: schema.GroupVersionKind{
				Group: "", Version: "v1", Kind: "ConfigMap",
			},
			expected: schema.GroupVersionKind{
				Group: "", Version: "v1", Kind: "Configmap",
			},
		},
		{
			name: "empty kind unchanged",
			input: schema.GroupVersionKind{
				Group: "", Version: "v1", Kind: "",
			},
			expected: schema.GroupVersionKind{
				Group: "", Version: "v1", Kind: "",
			},
		},
		{
			name: "group and version unchanged",
			input: schema.GroupVersionKind{
				Group: "APPS", Version: "V1", Kind: "pod",
			},
			expected: schema.GroupVersionKind{
				Group: "APPS", Version: "V1", Kind: "Pod",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeGVKForLookup(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeGVKForLookup(%v) = %v, want %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestMultipleRegistrations(t *testing.T) {
	// Clear mappers
	resourceMappers = make(map[schema.GroupVersionKind]ResourceMapper)

	// Register multiple mappers with various casings
	Register(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "pod"}, mockMapper)
	Register(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "POD"}, mockMapper)
	Register(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}, mockMapper)

	// Should only have one entry in the map (all normalized to same key)
	if len(resourceMappers) != 1 {
		t.Errorf("Expected 1 mapper after registering same kind with different cases, got %d",
			len(resourceMappers))
	}

	// All lookups should work
	for _, kind := range []string{"pod", "POD", "Pod", "pOd"} {
		_, found := Get(schema.GroupVersionKind{Group: "", Version: "v1", Kind: kind})
		if !found {
			t.Errorf("Failed to find mapper for kind %q", kind)
		}
	}
}

func TestMapperFunctionality(t *testing.T) {
	// Clear mappers
	resourceMappers = make(map[schema.GroupVersionKind]ResourceMapper)

	// Create a mapper that returns the item's name
	nameMapper := func(item unstructured.Unstructured) any {
		return item.GetName()
	}

	// Register it
	Register(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "test"}, nameMapper)

	// Get it back
	mapper, found := Get(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "TEST"})
	if !found {
		t.Fatal("Failed to find registered mapper")
	}

	// Test the mapper works
	testItem := unstructured.Unstructured{}
	testItem.SetName("test-name")

	result := mapper(testItem)
	if result != "test-name" {
		t.Errorf("Mapper returned %v, expected 'test-name'", result)
	}
}

func TestEdgeCases(t *testing.T) {
	// Clear mappers
	resourceMappers = make(map[schema.GroupVersionKind]ResourceMapper)

	t.Run("overwrite existing mapper", func(t *testing.T) {
		mapper1 := func(item unstructured.Unstructured) any { return "mapper1" }
		mapper2 := func(item unstructured.Unstructured) any { return "mapper2" }

		// Register first mapper
		Register(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "test"}, mapper1)

		// Overwrite with different casing
		Register(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "TEST"}, mapper2)

		// Should get mapper2
		mapper, found := Get(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Test"})
		if !found {
			t.Fatal("Failed to find mapper")
		}

		result := mapper(unstructured.Unstructured{})
		if result != "mapper2" {
			t.Errorf("Expected mapper2, got %v", result)
		}
	})

	t.Run("special characters in kind", func(t *testing.T) {
		// This shouldn't happen in real K8s, but let's ensure we don't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panicked with special characters: %v", r)
			}
		}()

		specialKinds := []string{
			"",          // empty
			"a",         // single char
			"123",       // numbers
			"test-kind", // with dash (invalid in K8s but shouldn't panic)
			"test.kind", // with dot
		}

		for _, kind := range specialKinds {
			Register(schema.GroupVersionKind{Group: "", Version: "v1", Kind: kind}, mockMapper)
			Get(schema.GroupVersionKind{Group: "", Version: "v1", Kind: kind})
		}
	})
}
