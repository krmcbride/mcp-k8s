package mapper

import (
	"testing"
)

func TestParseMemoryToMiB(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"", 0},
		{"128Mi", 128},
		{"1Gi", 1024},
		{"2Gi", 2048},
		{"512Mi", 512},
		{"1024Mi", 1024},
		{"500000000", 476}, // ~500MB in bytes = ~476 MiB
		{"1000Mi", 1000},
		{"1.5Gi", 1536}, // 1.5 * 1024 = 1536 MiB
		{"invalid", 0},
		{"123", 0}, // Raw number without unit treated as bytes, very small result
	}

	for _, test := range tests {
		result := parseMemoryToMiB(test.input)
		if result != test.expected {
			t.Errorf("parseMemoryToMiB(%q) = %d, expected %d", test.input, result, test.expected)
		}
	}
}
