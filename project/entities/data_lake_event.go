package entities

import "time"

type DataLakeEvent struct {
	EventID      string    `json:"event_id" db:"event_id"`
	EventName    string    `json:"event_name" db:"event_name"`
	EventPayload []byte    `json:"event_payload" db:"event_payload"`
	PublishedAt  time.Time `json:"published_at" db:"published_at"`
}
