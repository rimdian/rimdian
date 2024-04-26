package dto

import (
	"net/http"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type AppFromTokenParams struct {
	Token string `json:"token"`
}

type AppFromToken struct {
	App *entity.App `json:"app"`
}

type AppListResult struct {
	Apps []*entity.App `json:"apps"`
}

type AppListParams struct {
	WorkspaceID string `json:"workspace_id"`
}

func (params *AppListParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	if params.WorkspaceID == "" {
		return eris.New("app list: workspace_id is required")
	}

	return nil
}

type AppGetParams struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

func (params *AppGetParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	if params.WorkspaceID == "" {
		return eris.New("app get: workspace_id is required")
	}

	params.ID = r.FormValue("id")
	if params.ID == "" {
		return eris.New("app get: id is required")
	}
	return nil
}

type AppInstall struct {
	WorkspaceID string              `json:"workspace_id"`
	Manifest    *entity.AppManifest `json:"manifest,omitempty"`
	SecretKey   *string             `json:"secret_key,omitempty"`
	Reinstall   bool                `json:"reinstall"`
}

type AppDelete struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

type AppMutateState struct {
	WorkspaceID string                     `json:"workspace_id"`
	ID          string                     `json:"id"`
	Mutations   []*entity.AppStateMutation `json:"mutations"`
}

func (data *AppMutateState) Validate() error {
	if data.WorkspaceID == "" {
		return eris.New("workspace_id is required")
	}

	if data.ID == "" {
		return eris.New("id is required")
	}

	if len(data.Mutations) == 0 {
		return eris.New("mutations is required")
	}

	for _, mutation := range data.Mutations {
		if mutation.Operation != "set" && mutation.Operation != "delete" {
			return eris.New("mutation operation is not valid")
		}

		if mutation.Key == "" {
			return eris.New("mutation key is required")
		}

		if mutation.Operation == "set" && mutation.Value == nil {
			return eris.Errorf("mutation value is required for mutation: %s %v", mutation.Operation, mutation.Key)
		}
	}

	return nil
}

type AppActivate struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

type AppExecQuery struct {
	WorkspaceID string        `json:"workspace_id"`
	AppID       string        `json:"app_id"`
	QueryID     string        `json:"query_id"`
	Query       string        `json:"query"`
	Args        []interface{} `json:"args"`
}

func (data *AppExecQuery) Validate() error {
	if data.WorkspaceID == "" {
		return eris.New("workspace_id is required")
	}

	if data.AppID == "" {
		return eris.New("app_id is required")
	}

	if data.QueryID == "" && data.Query == "" {
		return eris.New("query_id or query is required")
	}

	return nil
}

type AppExecQueryResult struct {
	Rows   []map[string]interface{} `json:"rows"`
	TookMs int64                    `json:"took_ms"`
}
