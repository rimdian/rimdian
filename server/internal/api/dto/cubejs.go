package dto

type CubeJSSchemas []CubeJSSchemaFile

type CubeJSSchemaFile struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}
