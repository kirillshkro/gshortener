package types

type ActionType string

const (
	ActionCreate ActionType = "shorten"
	ActionFollow ActionType = "follow"
)

type auditData struct {
	TimestampEvent int64      `json:"ts"`
	Action         ActionType `json:"action"`
	UserID         string     `json:"user_id"`
	URL            string     `json:"url"`
}

type Event = auditData
