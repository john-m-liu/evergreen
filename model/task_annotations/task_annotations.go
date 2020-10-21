package task_annotations

import (
	"encoding/json"
	"time"
)

type TaskAnnotation struct {
	Id string `bson:"_id" json:"id"`
	// Comment about the failure
	Note string `bson:"note,omitempty" json:"note,omitempty"`
	// Links to tickets definitely related.
	Issues []IssueLink `bson:"issues,omitempty" json:"issue,omitempty"`
	// Links to tickets possibly related
	SuspectIssues []IssueLink `bson:"suspected_issues,omitempty" json:"suspected_issues,omitempty"`
	// annotation attribution
	Source AnnotationSource `bson:"source,omitempty" json:"source,omitempty"`
	// structured data about the task (not displayed in the UI, but available in the API)
	Metadata json.RawMessage `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

type IssueLink struct {
	URL string `bson:"url" json:"url"`
	// Text to be displayed
	IssueKey string `bson:"issue_key,omitempty" json:"issue_key,omitempty"`
}

type AnnotationSource struct {
	Author string `bson:"author,omitempty" json:"author,omitempty"`
	// Text to be displayed
	Time time.Time `bson:"time,omitempty" json:"time,omitempty"`
}
