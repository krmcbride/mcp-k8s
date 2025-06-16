package mapper

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CronJobListContent represents CronJob-specific fields for list display
type CronJobListContent struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace,omitempty"`
	Schedule     string `json:"schedule,omitempty"`
	Suspend      bool   `json:"suspend,omitempty"`
	Active       int64  `json:"active,omitempty"`
	LastSchedule string `json:"lastSchedule,omitempty"`
	Age          string `json:"age,omitempty"`
}

func init() {
	// Register CronJob mapper
	Register(
		schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "CronJob"},
		mapCronJobResource,
	)
}

func mapCronJobResource(item unstructured.Unstructured) interface{} {
	cronJob := CronJobListContent{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}

	// Extract CronJob-specific fields from spec
	if schedule, found, _ := unstructured.NestedString(item.Object, "spec", "schedule"); found {
		cronJob.Schedule = schedule
	}

	if suspend, found, _ := unstructured.NestedBool(item.Object, "spec", "suspend"); found {
		cronJob.Suspend = suspend
	}

	// Extract status fields
	if active, found, _ := unstructured.NestedSlice(item.Object, "status", "active"); found {
		cronJob.Active = int64(len(active))
	}

	if lastScheduleTime, found, _ := unstructured.NestedString(item.Object, "status", "lastScheduleTime"); found {
		cronJob.LastSchedule = lastScheduleTime
	}

	// TODO: Calculate age from creation timestamp

	return cronJob
}
