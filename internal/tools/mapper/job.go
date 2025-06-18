package mapper

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// JobListContent represents Job-specific fields for list display
type JobListContent struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace,omitempty"`
	Completions string `json:"completions,omitempty"`
	Duration    string `json:"duration,omitempty"`
	Age         string `json:"age,omitempty"`
}

func init() {
	// Register Job mapper
	Register(
		schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"},
		mapJobResource,
	)
}

func mapJobResource(item unstructured.Unstructured) any {
	job := JobListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}

	// Extract Job completion status
	var succeeded, failed int64
	if succeededCount, found, _ := unstructured.NestedInt64(item.Object, "status", "succeeded"); found {
		succeeded = succeededCount
	}
	if failedCount, found, _ := unstructured.NestedInt64(item.Object, "status", "failed"); found {
		failed = failedCount
	}

	// Get desired completions from spec
	if completions, found, _ := unstructured.NestedInt64(item.Object, "spec", "completions"); found {
		job.Completions = fmt.Sprintf("%d/%d", succeeded, completions)
	} else {
		// Jobs without completions specified show succeeded count
		job.Completions = fmt.Sprintf("%d", succeeded)
	}

	// Show failed count if any failures
	if failed > 0 {
		job.Completions += fmt.Sprintf(" (%d failed)", failed)
	}

	// Calculate duration if job has started and completed
	if startTime, found, _ := unstructured.NestedString(item.Object, "status", "startTime"); found && startTime != "" {
		if completionTime, found, _ := unstructured.NestedString(item.Object, "status", "completionTime"); found && completionTime != "" {
			// TODO: Parse timestamps and calculate duration
			job.Duration = "completed"
		} else {
			// Job is still running
			job.Duration = "running"
		}
	}

	// TODO: Calculate age from creation timestamp

	return job
}
