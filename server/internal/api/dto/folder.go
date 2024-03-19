package dto

import (
	"net/http"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type FolderFilesParams struct {
	FolderID string
}

func (p *FolderFilesParams) FromRequest(r *http.Request) error {

	p.FolderID = r.FormValue("folder_id")

	if p.FolderID == "" {
		return eris.New("folder_id is required")
	}

	return nil
}

type FolderFilesResult struct {
	Files []entity.FolderFile `json:"files"`
}
