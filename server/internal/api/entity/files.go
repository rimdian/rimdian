package entity

import (
	"database/sql/driver"
	"encoding/json"
)

type FoldersTree struct {
	ID       string         `json:"id"`
	Path     string         `json:"path"`
	Name     string         `json:"name"`
	Children []*FoldersTree `json:"children"`
}

func (x *FoldersTree) Scan(val interface{}) error {
	return json.Unmarshal(val.([]byte), &x)
}

func (x FoldersTree) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type FolderFile struct {
	ID          string `json:"id"`
	FolderID    string `json:"folder_id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	DBCreatedAt string `json:"db_created_at"`
	DBUpdatedAt string `json:"db_updated_at"`
}
