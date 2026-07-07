// Package model provides a set of data models for the application.
package model

// FileData represents a file data structure in JSON format.
type FileData struct {
	// UUID is a unique identifier for each record.
	UUID string `json:"uuid"`
	// URLData contains the original and shortened URLs.
	URLData
}
