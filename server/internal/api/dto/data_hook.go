package dto

import "github.com/rimdian/rimdian/internal/api/entity"

type DataHook struct {
	WorkspaceID string                `json:"workspace_id"`
	ID          string                `json:"id"`
	AppID       string                `json:"app_id"`
	On          string                `json:"on"`
	Name        string                `json:"name"`
	For         []*entity.DataHookFor `json:"for"`
	JS          *string               `json:"js"`
	Enabled     bool                  `json:"enabled"`
}
