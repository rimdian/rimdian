package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) InsertMessageTemplate(ctx context.Context, workspaceID string, template *entity.MessageTemplate, tx *sql.Tx) (err error) {

	query, args, err := sq.Insert("message_template").
		Columns(
			"id",
			"version",
			"name",
			"channel",
			"category",
			"engine",
			"email",
			"template_macro_id",
			"utm_source",
			"utm_medium",
			"utm_campaign",
			"settings",
			"test_data",
		).
		Values(
			template.ID,
			template.Version,
			template.Name,
			template.Channel,
			template.Category,
			template.Engine,
			template.Email,
			template.TemplateMacroID,
			template.UTMSource,
			template.UTMMedium,
			template.UTMCampaign,
			template.Settings,
			template.TestData,
		).
		ToSql()

	if err != nil {
		return eris.Wrapf(err, "InsertMessageTemplate insert query: %v, args: %+v", query, args)
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "InsertMessageTemplate insert exec: %v, args: %+v", query, args)
	}

	return nil
}

func (repo *RepositoryImpl) GetMessageTemplate(ctx context.Context, workspaceID string, id string, version *int64, tx *sql.Tx) (template *entity.MessageTemplate, err error) {

	builder := sq.Select("*").From("message_template").Where(sq.Eq{"id": id})

	if version != nil {
		builder = builder.Where(sq.Eq{"version": *version})
	}

	query, args, err := builder.ToSql()

	if err != nil {
		err = eris.Wrapf(err, "GetMessageTemplate fetch query: %v, args: %+v", query, args)
		return
	}

	template = &entity.MessageTemplate{}

	if tx != nil {
		err = sqlscan.Get(ctx, tx, template, query, args...)
	} else {
		conn, errConn := repo.GetWorkspaceConnection(ctx, workspaceID)
		if errConn != nil {
			return nil, errConn
		}
		defer conn.Close()

		err = sqlscan.Get(ctx, conn, template, query, args...)
	}

	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		err = eris.Wrapf(err, "GetMessageTemplate query: %v, args: %+v", query, args)
		return
	}

	return
}

func (repo *RepositoryImpl) ListMessageTemplates(ctx context.Context, workspaceID string, params *dto.MessageTemplateListParams) (lists []*entity.MessageTemplate, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	lists = []*entity.MessageTemplate{}

	builder := sq.Select("t1.*").From("message_template as t1").InnerJoin("(SELECT id, MAX(version) AS max_version FROM message_template GROUP BY id) t2 ON t1.id = t2.id AND t1.version = t2.max_version")

	if params.Channel != nil {
		builder = builder.Where(sq.Eq{"t1.channel": *params.Channel})
	}

	if params.Category != nil {
		builder = builder.Where(sq.Eq{"t1.category": *params.Category})
	}

	// fetch lists
	query, args, err := builder.ToSql()

	// log.Println("query", query)

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
