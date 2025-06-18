package mapper

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// EventListContent represents Event-specific fields for list display
type EventListContent struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace,omitempty"`
	Type           string `json:"type,omitempty"`           // Normal, Warning
	Reason         string `json:"reason,omitempty"`         // Human-readable reason
	Message        string `json:"message,omitempty"`        // Detailed description
	InvolvedObject string `json:"involvedObject,omitempty"` // Object that triggered the event
	Source         string `json:"source,omitempty"`         // Event creator
	Count          int64  `json:"count,omitempty"`          // Number of occurrences
	FirstTimestamp string `json:"firstTimestamp,omitempty"` // First occurrence (core/v1)
	LastTimestamp  string `json:"lastTimestamp,omitempty"`  // Last occurrence (core/v1)
	EventTime      string `json:"eventTime,omitempty"`      // Event time (events/v1beta1)
	Age            string `json:"age,omitempty"`            // Time since first/event timestamp
}

func init() {
	// Register Event mapper for core/v1 events
	Register(
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Event"},
		mapEventResource,
	)
	// Register Event mapper for events/v1beta1
	Register(
		schema.GroupVersionKind{Group: "events.k8s.io", Version: "v1beta1", Kind: "Event"},
		mapEventResource,
	)
}

func mapEventResource(item unstructured.Unstructured) any {
	event := &EventListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}

	// Extract event type (Normal, Warning)
	if eventType, found, _ := unstructured.NestedString(item.Object, "type"); found {
		event.Type = eventType
	}

	// Extract reason
	if reason, found, _ := unstructured.NestedString(item.Object, "reason"); found {
		event.Reason = reason
	}

	// Extract message/note
	if message, found, _ := unstructured.NestedString(item.Object, "message"); found {
		event.Message = message
	} else if note, found, _ := unstructured.NestedString(item.Object, "note"); found {
		// events/v1beta1 uses 'note' instead of 'message'
		event.Message = note
	}

	// Extract involved object information
	if involvedObj, found, _ := unstructured.NestedMap(item.Object, "involvedObject"); found {
		var objInfo string
		if kind, exists := involvedObj["kind"]; exists {
			objInfo = kind.(string)
		}
		if name, exists := involvedObj["name"]; exists {
			if objInfo != "" {
				objInfo += "/" + name.(string)
			} else {
				objInfo = name.(string)
			}
		}
		event.InvolvedObject = objInfo
	}

	// Extract source information
	if source, found, _ := unstructured.NestedMap(item.Object, "source"); found {
		var sourceInfo string
		if component, exists := source["component"]; exists {
			sourceInfo = component.(string)
		}
		if host, exists := source["host"]; exists {
			if sourceInfo != "" {
				sourceInfo += "@" + host.(string)
			} else {
				sourceInfo = host.(string)
			}
		}
		event.Source = sourceInfo
	}

	// Extract count
	if count, found, _ := unstructured.NestedInt64(item.Object, "count"); found {
		event.Count = count
	}

	// Handle timestamps (different fields for different API versions)
	var ageTime time.Time

	// Try core/v1 timestamps first
	if firstTimestamp, found, _ := unstructured.NestedString(item.Object, "firstTimestamp"); found {
		event.FirstTimestamp = firstTimestamp
		if parsed, err := time.Parse(time.RFC3339, firstTimestamp); err == nil {
			ageTime = parsed
		}
	}

	if lastTimestamp, found, _ := unstructured.NestedString(item.Object, "lastTimestamp"); found {
		event.LastTimestamp = lastTimestamp
	}

	// Try events/v1beta1 eventTime
	if eventTime, found, _ := unstructured.NestedString(item.Object, "eventTime"); found {
		event.EventTime = eventTime
		// If we don't have firstTimestamp, use eventTime for age calculation
		if event.FirstTimestamp == "" {
			if parsed, err := time.Parse(time.RFC3339, eventTime); err == nil {
				ageTime = parsed
			}
		}
	}

	// Calculate age if we have a timestamp
	if !ageTime.IsZero() {
		event.Age = formatDuration(time.Since(ageTime))
	}

	return event
}

// formatDuration formats a duration in a human-readable way similar to kubectl
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "< 1m"
	}
	if d < time.Hour {
		return formatUnit(int(d.Minutes()), "m")
	}
	if d < 24*time.Hour {
		return formatUnit(int(d.Hours()), "h")
	}
	return formatUnit(int(d.Hours()/24), "d")
}

func formatUnit(value int, unit string) string {
	return fmt.Sprintf("%d%s", value, unit)
}
