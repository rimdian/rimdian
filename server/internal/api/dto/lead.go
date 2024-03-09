package dto

import "time"

type LeadStage struct {
	ID          string     `json:"id"`
	Label       string     `json:"label"`
	Status      string     `json:"status"` // open | converted | lost
	Color       string     `json:"color"`
	DeletedAt   *time.Time `json:"deleted_at"`
	MigrateToID *string    `json:"migrateToId"` // used when delete a status, migrate existing leads to new stage ID
}
