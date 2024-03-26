package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) ListMessageTemplates(ctx context.Context, workspaceID string, params *dto.MessageTemplateListParams) (lists []*entity.MessageTemplate, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	lists = []*entity.MessageTemplate{}

	builder := sq.Select("*").From("message_template")

	if params.Channel != nil {
		builder = builder.Where(sq.Eq{"channel": *params.Channel})
	}

	// fetch lists
	query, args, err := builder.ToSql()

	if err != nil {
		err = eris.Wrapf(err, "ListMessageTemplates fetch query: %v, args: %+v", query, args)
		return
	}

	if err = sqlscan.Select(ctx, conn, &lists, query, args...); err != nil {
		err = eris.Wrapf(err, "ListMessageTemplates query: %v, args: %+v", query, args)
		return
	}

	return
}
