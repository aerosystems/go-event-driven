package entities

import "time"

type DataLakeEvent struct {
	ID          string    `json:"event_id"`
	Name        string    `json:"event_name"`
	Payload     []byte    `json:"event_payload"`
	PublishedAt time.Time `json:"published_at"`
}
