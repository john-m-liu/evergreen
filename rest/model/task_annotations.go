package model

import (
	"time"
)

type ApiTaskAnnotation struct {
	Note *string `bson:"note,omitempty" json:"note,omitempty"`
	// Links to tickets definitely related.
	Issues []APIIssueLink `bson:"issues,omitempty" json:"issue,omitempty"`
	// Links to tickets possibly related
	SuspectIssues []APIIssueLink `bson:"suspected_issues,omitempty" json:"suspected_issues,omitempty"`
	// annotation attribution
	Source APIAnnotationSource `bson:"source,omitempty" json:"source,omitempty"`
}

type APIIssueLink struct {
	URL *string `bson:"url" json:"url"`
	// Text to be displayed
	IssueKey *string `bson:"issue_key,omitempty" json:"issue_key,omitempty"`
}

type APIAnnotationSource struct {
	Author string `bson:"author,omitempty" json:"author,omitempty"`
	// Text to be displayed
	Time time.Time `bson:"time,omitempty" json:"time,omitempty"`
}

// func (t *ApiTaskAnnotation) BuildFromService(h interface{}) error {
// 	switch v := h.(type) {
// 	case task_annotations.TaskAnnotation:
// 		t.Note = ToStringPtr(v.Note)
// 		issues := []APIIssueLink{}
// 		if err := issues.BuildFromService(v.issues); err != nil {
// 			return errors.Wrap(err, "Error converting from task_annotations.APIIssueLink to model.APIIssueLink")
// 		}
// 		t.SuspectIssues = v.SuspectIssues
// 		t.Source = v.Source
// 	default:
// 		return errors.Errorf("%T is not a supported type", h)
// 	}
// 	return nil
// }

// func (t *ApiTaskAnnotation) ToService() (interface{}, error) {
// 	return task_annotations.TaskAnnotation{
// 		Note:          FromStringPtr(t.Note),
// 		Issues:        FromStringPtr(t.Issues),
// 		SuspectIssues: FromStringPtr(t.SuspectIssues),
// 		Source:        t.Source,
// 	}, nil
// }
