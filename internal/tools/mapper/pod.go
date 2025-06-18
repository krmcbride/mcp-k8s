package mapper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PodListContent represents Pod-specific fields for list display
type PodListContent struct {
	Name                  string `json:"name"`
	Namespace             string `json:"namespace,omitempty"`
	Status                string `json:"status,omitempty"`
	Ready                 string `json:"ready,omitempty"`
	Restarts              int64  `json:"restarts,omitempty"`
	Age                   string `json:"age,omitempty"`
	MemoryRequestMiB      int64  `json:"memoryRequestMiB,omitempty"`
	MemoryLimitMiB        int64  `json:"memoryLimitMiB,omitempty"`
	OOMKills              int64  `json:"oomKills,omitempty"`
	LastTerminationReason string `json:"lastTerminationReason,omitempty"`
}

// parseMemoryToMiB converts Kubernetes memory strings to MiB
// Supports formats like: "128Mi", "1Gi", "512000000", "1000000k", etc.
func parseMemoryToMiB(memoryStr string) int64 {
	if memoryStr == "" {
		return 0
	}

	// Define conversion ratios to MiB
	units := map[string]int64{
		"":   1024 * 1024, // bytes to MiB
		"k":  1024,        // kilobytes to MiB
		"Ki": 1024,        // kibibytes to MiB
		"M":  1,           // megabytes to MiB (approximately)
		"Mi": 1,           // mebibytes to MiB
		"G":  1024,        // gigabytes to MiB (approximately)
		"Gi": 1024,        // gibibytes to MiB
		"T":  1024 * 1024, // terabytes to MiB (approximately)
		"Ti": 1024 * 1024, // tebibytes to MiB
	}

	// Use regex to parse number and unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([A-Za-z]*)$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(memoryStr))

	if len(matches) != 3 {
		return 0
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	unit := matches[2]
	multiplier, exists := units[unit]
	if !exists {
		return 0
	}

	// Convert to MiB
	if unit == "" || unit == "k" || unit == "Ki" {
		// Convert from smaller units
		return int64(value / float64(multiplier))
	} else {
		// Convert from larger units
		return int64(value * float64(multiplier))
	}
}

func init() {
	// Register Pod mapper
	Register(
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		mapPodResource,
	)
}

func mapPodResource(item unstructured.Unstructured) interface{} {
	pod := PodListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}

	// Extract Pod-specific fields
	if status, found, _ := unstructured.NestedString(item.Object, "status", "phase"); found {
		pod.Status = status
	}

	// Extract memory resources from container specs
	if containers, found, _ := unstructured.NestedSlice(item.Object, "spec", "containers"); found {
		var totalMemoryRequest, totalMemoryLimit int64

		for _, c := range containers {
			if containerMap, ok := c.(map[string]interface{}); ok {
				// Extract memory request
				if memReq, found, _ := unstructured.NestedString(containerMap, "resources", "requests", "memory"); found {
					totalMemoryRequest += parseMemoryToMiB(memReq)
				}
				// Extract memory limit
				if memLimit, found, _ := unstructured.NestedString(containerMap, "resources", "limits", "memory"); found {
					totalMemoryLimit += parseMemoryToMiB(memLimit)
				}
			}
		}

		pod.MemoryRequestMiB = totalMemoryRequest
		pod.MemoryLimitMiB = totalMemoryLimit
	}

	// Extract container statuses for ready count, restarts, and OOM kills
	if containers, found, _ := unstructured.NestedSlice(item.Object, "status", "containerStatuses"); found {
		ready := 0
		total := len(containers)
		restarts := int64(0)
		oomKills := int64(0)
		var lastTerminationReason string

		for _, c := range containers {
			if containerMap, ok := c.(map[string]interface{}); ok {
				if r, found, _ := unstructured.NestedBool(containerMap, "ready"); found && r {
					ready++
				}
				if rc, found, _ := unstructured.NestedInt64(containerMap, "restartCount"); found {
					restarts += rc
				}

				// Check for OOM kills in lastState.terminated
				if lastState, found, _ := unstructured.NestedMap(containerMap, "lastState"); found {
					if terminated, found, _ := unstructured.NestedMap(lastState, "terminated"); found {
						if reason, found, _ := unstructured.NestedString(terminated, "reason"); found {
							lastTerminationReason = reason
							if reason == "OOMKilled" {
								oomKills++
							}
						}
					}
				}

				// Also check current state if it's terminated
				if state, found, _ := unstructured.NestedMap(containerMap, "state"); found {
					if terminated, found, _ := unstructured.NestedMap(state, "terminated"); found {
						if reason, found, _ := unstructured.NestedString(terminated, "reason"); found {
							if reason == "OOMKilled" {
								oomKills++
							}
						}
					}
				}
			}
		}

		pod.Ready = fmt.Sprintf("%d/%d", ready, total)
		pod.Restarts = restarts
		pod.OOMKills = oomKills
		pod.LastTerminationReason = lastTerminationReason
	}

	// TODO: Calculate age from creation timestamp

	return pod
}
