// Package model contains data models for the application
package model

// FileData represents the structure for file data with UUID and URL information
type FileData struct {
	// UUID is the unique identifier for the file
	UUID string `json:"uuid"`
	// URLData contains the URL-related information
	URLData
}
