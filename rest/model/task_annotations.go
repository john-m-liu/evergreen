package model

import (
	"time"
)

type ApiTaskAnnotations struct {
	Id string `bson:"_id" json:"id"`
	// Comment about the failure
	Note string `bson:"note,omitempty" json:"note,omitempty"`
	// Links to tickets definitely related.
	Issues []*IssueLink `bson:"issues,omitempty" json:"issue,omitempty"`
	// Links to tickets possibly related
	SuspectIssues []*IssueLink `bson:"suspected_issues,omitempty" json:"suspected_issues,omitempty"`
	// annotation attribution
	Source AnnotationSource `bson:"source,omitempty" json:"source,omitempty"`
	// structured data about the task (not displayed in the UI, but available in the API)
	Metadata map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
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

// func (f *APIFile) BuildFromService(h interface{}) error {
// 	switch v := h.(type) {
// 	case artifact.File:
// 		f.Name = ToStringPtr(v.Name)
// 		f.Link = ToStringPtr(v.Link)
// 		f.Visibility = ToStringPtr(v.Visibility)
// 		f.IgnoreForFetch = v.IgnoreForFetch
// 	default:
// 		return errors.Errorf("%T is not a supported type", h)
// 	}
// 	return nil
// }

// func (f *APIFile) ToService() (interface{}, error) {
// 	return artifact.File{
// 		Name:           FromStringPtr(f.Name),
// 		Link:           FromStringPtr(f.Link),
// 		Visibility:     FromStringPtr(f.Visibility),
// 		IgnoreForFetch: f.IgnoreForFetch,
// 	}, nil
// }
