package repository

import (
	"context"
	"log"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) QueryAnalytics(ctx context.Context, workspace *entity.Workspace, schemasMap map[string]*entity.CubeJSSchema, columns []string, query string, args []interface{}) (result *dto.DBAnalyticsResult, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

	if err != nil {
		return nil, eris.Wrap(err, "QueryAnalytics")
	}

	defer conn.Close()

	rows, err := conn.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, eris.Wrap(err, "QueryAnalytics")
	}

	defer rows.Close()

	// extract columns names
	cols, err := rows.Columns()

	if err != nil {
		return nil, eris.Wrap(err, "QueryAnalytics")
	}

	log.Printf("Columns: %+v\n", cols)

	result = &dto.DBAnalyticsResult{
		SQL:  query,
		Args: args,
		Data: nil,
	}

	return result, nil
}
