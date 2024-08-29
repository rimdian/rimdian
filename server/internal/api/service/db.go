package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) DBAnalytics(ctx context.Context, accountID string, params *dto.DBAnalyticsParams) (result *dto.DBAnalyticsResult, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, accountID, params.WorkspaceID)

	if err != nil {
		return nil, code, eris.Wrap(err, "DBAnalytics")
	}

	payload := struct {
		CubeQuery string        `json:"cubeQuery"`
		Schemas   []interface{} `json:"schemas"`
	}{
		CubeQuery: string(params.Query),
		Schemas:   []interface{}{},
	}

	files := []dto.CubeJSSchemaFile{}
	schemasMap := entity.GenerateSchemas(workspace.InstalledApps)

	for fileName, schema := range schemasMap {
		files = append(files, dto.CubeJSSchemaFile{FileName: fileName + ".js", Content: "cube(" + fileName + "," + schema.BuildContent(fileName) + ")"})
	}

	// json encode payload
	payloadJSON, err := json.Marshal(payload)

	if err != nil {
		return nil, 500, eris.Wrap(err, "DBAnalytics")
	}

	payloadB64 := base64.StdEncoding.EncodeToString(payloadJSON)

	dir, err := GetNodeJSDir()
	if err != nil {
		return nil, 500, eris.Wrap(err, "DBAnalytics")
	}

	scriptPath := dir + "cube-to-query.js"

	// svc.Logger.Printf("scriptPath: %v, payload %v", scriptPath, payloadB64)

	// call nodejs cmd
	data, err := exec.Command("node", scriptPath, payloadB64).Output()

	if err != nil {
		return nil, 500, eris.Wrap(err, "DBAnalytics")
	}

	// decode output
	cubeResult := struct {
		SQL  string        `json:"sql"`
		Args []interface{} `json:"args"`
	}{}

	if err = json.Unmarshal(data, &cubeResult); err != nil {
		return nil, 400, eris.Wrap(err, "DBAnalytics")
	}

	// TODO: extract measures and dimensions from params.Query raw JSON
	// used to cast properly the output
	columns := []string{}

	// exec query
	result, err = svc.Repo.QueryAnalytics(ctx, workspace, schemasMap, columns, cubeResult.SQL, cubeResult.Args)

	if err != nil {
		return nil, 500, eris.Wrap(err, "DBAnalytics")
	}

	return result, 200, nil
}

func (svc *ServiceImpl) DBSelect(ctx context.Context, accountID string, params *dto.DBSelectParams) (rows []map[string]interface{}, code int, err error) {

	if params.Validate() != nil {
		return nil, 400, eris.Wrap(err, "Validate")
	}

	// fetch workspace
	workspace, err := svc.Repo.GetWorkspace(ctx, params.WorkspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "DBSelect")
	}

	// verify that token is owner of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "DBSelect")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	query := squirrel.Select(params.Columns...).From(fmt.Sprintf("`%v`", params.From)).Limit(uint64(params.Limit))

	if params.Where != "" {
		query = query.Where(params.Where, params.Args...)
	}

	if params.OrderBy != "" {
		query = query.OrderBy(params.OrderBy)
	}

	if params.GroupBy != "" {
		query = query.GroupBy(params.GroupBy)
	}

	if params.Offset != 0 {
		query = query.Offset(uint64(params.Offset))
	}

	sql, args, err := query.ToSql()

	if err != nil {
		return nil, 500, eris.Wrap(err, "DoDBSelect")
	}

	// svc.Logger.Printf("sql: %v, args: %v", sql, args)

	jsonRows, err := svc.DoDBSelect(params.WorkspaceID, sql, args)

	if err != nil {
		svc.Logger.Printf("error DBSelect output: %v, %v", string(jsonRows), err)
		return nil, 500, eris.Wrap(err, "DBSelect")
	}

	if jsonRows == nil {
		return []map[string]interface{}{}, 200, nil
	}

	// decode output
	if err = json.Unmarshal(jsonRows, &rows); err != nil {
		svc.Logger.Printf("error DBSelect output: %v", string(jsonRows))
		return nil, 400, eris.Wrap(err, "DBSelect")
	}

	return rows, 200, nil
}

// Golang SQL driver does not cast types correctly on interface{} (only []byte and then goodluck to guess).
// We use nodejs to do the query + JSON casting.
func (svc *ServiceImpl) DoDBSelect(workspaceID string, query string, args []interface{}) (output []byte, err error) {

	// TODO: test repo.SelectQueryToJSON

	payload := struct {
		Query string        `json:"query"`
		Args  []interface{} `json:"args"`
		DB    string        `json:"db"`
	}{
		DB:    svc.Config.DB_PREFIX + workspaceID,
		Query: query,
		Args:  args,
	}

	// json encode payload
	payloadJSON, err := json.Marshal(payload)

	if err != nil {
		return nil, eris.Wrap(err, "DoDBSelect")
	}

	payloadB64 := base64.StdEncoding.EncodeToString(payloadJSON)

	dir, err := GetNodeJSDir()
	if err != nil {
		return nil, eris.Wrap(err, "DoDBSelect")
	}

	scriptPath := dir + "query.js"

	// svc.Logger.Printf("scriptPath: %v, payload %v", scriptPath, payloadB64)

	// call nodejs cmd
	return exec.Command("node", scriptPath, payloadB64).Output()
}
