package dto

type DataHook struct {
	WorkspaceID string   `json:"workspace_id"`
	ID          string   `json:"id"`
	AppID       string   `json:"app_id"`
	On          string   `json:"on"`
	Name        string   `json:"name"`
	Kind        []string `json:"kind"`
	Action      []string `json:"action"`
	JS          *string  `json:"js"`
	Enabled     bool     `json:"enabled"`
}
