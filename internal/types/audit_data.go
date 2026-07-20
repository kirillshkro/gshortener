// Package types provides data structures for audit events and actions.
package types

// type user actions

type ActionType string

const (
	ActionCreate ActionType = "shorten" // ActionCreate is the action type for creating a short URL.
	ActionFollow ActionType = "follow"  // ActionFollow is the action type for following a user.
)

// auditData represents an audit event data. It contains information about the time of the event, the action performed, the user who performed it and the URL associated with that action.
type auditData struct {
	TimestampEvent int64      `json:"ts"`      // TimestampEvent is the timestamp when the event occurred.
	Action         ActionType `json:"action"`  // Action is the type of action performed.
	UserID         string     `json:"user_id"` // UserID is the ID of the user who performed the action.
	URL            string     `json:"url"`     // URL is the URL associated with the action.
}

// Event represents an audit event. It's just an alias for auditData type.
type Event = auditData
