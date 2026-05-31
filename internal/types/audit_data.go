package types

import "time"

type ActionType string

const (
	ActionCreate ActionType = "shorten"
	ActionFollow ActionType = "follow"
)

type AuditData struct {
	TimestampEvent time.Time  `json:"ts"`
	Action         ActionType `json:"action"`
	UserID         string     `json:"user_id"`
	URL            string     `json:"url"`
}
