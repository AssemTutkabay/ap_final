package domain

import "time"
import "encoding/json"

type AuditLog struct {
	ActorUserID *string         `json:"actorUserId,omitempty"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entityType"`
	EntityID    *string         `json:"entityId,omitempty"`
	At          time.Time       `json:"at"`
	Details     json.RawMessage `json:"details,omitempty"`
}
