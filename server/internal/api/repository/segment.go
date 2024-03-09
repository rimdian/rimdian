package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func segmentNodeAsSQL(parentSegmentID *string, node *entity.SegmentTreeNode, timezone string) (sql string, args []interface{}, err error) {
	if args == nil {
		args = []interface{}{}
	}
	// branch
	if node.Kind == entity.SegmentTreeNodeKindBranch {
		toJoin := []string{}
		for _, leaf := range node.Branch.Leaves {
			sqlChunk, argsChunk, err := segmentNodeAsSQL(parentSegmentID, leaf, timezone)
			if err != nil {
				return "", nil, err
			}
			sqlChunk = fmt.Sprintf(`u.id IN (%s)`, sqlChunk)
			args = append(args, argsChunk...)
			toJoin = append(toJoin, sqlChunk)
		}
		joinOperator := " AND "
		if node.Branch.Operator == "or" {
			joinOperator = " OR "
		}
		return "(" + strings.Join(toJoin, joinOperator) + ")", args, nil
	}

	// leaf

	// append an alias to the table in case many joins are done
	alias := fmt.Sprintf("%v__%v", node.Leaf.Table, rand.Intn(1000))

	// build the where clause
	whereBuilder, err := filterAsQueryBuilder(alias, node.Leaf.Filters, timezone)
	if err != nil {
		return "", nil, err
	}

	if node.Leaf.Table == "user" {

		q := sq.Select(alias + ".id").From("`user` as " + alias)

		if parentSegmentID != nil {

			if *parentSegmentID != entity.SegmentAllUsersID {
				if *parentSegmentID == "anonymous" {
					q.Where(sq.Eq{alias + ".is_authenticated": 0})
				}
				if *parentSegmentID == "authenticated" {
					q.Where(sq.Eq{alias + ".is_authenticated": 1})
				}
			}
		}

		return q.Where(whereBuilder).ToSql()
	}

	queryBuilder := sq.Select(alias + ".user_id").From(fmt.Sprintf("`%v` as %v", node.Leaf.Table, alias)).Where(whereBuilder).GroupBy(alias + ".user_id")

	if node.Leaf.Action.CountOperator == "at_least" {
		queryBuilder = queryBuilder.Having(sq.GtOrEq{"COUNT(*)": node.Leaf.Action.CountValue})
	} else if node.Leaf.Action.CountOperator == "at_most" {
		queryBuilder = queryBuilder.Having(sq.LtOrEq{"COUNT(*)": node.Leaf.Action.CountValue})
	} else if node.Leaf.Action.CountOperator == "exactly" {
		queryBuilder = queryBuilder.Having(sq.Eq{"COUNT(*)": node.Leaf.Action.CountValue})
	}

	if node.Leaf.Action.TimeframeOperator == "anytime" {
		// do nothing
	} else if node.Leaf.Action.TimeframeOperator == "in_date_range" {
		queryBuilder = queryBuilder.Having(sq.Expr("MAX("+alias+".created_at_trunc) >= TIMESTAMP(convert_tz(?, '+00:00', ?))", node.Leaf.Action.TimeframeValues[0], timezone))
		queryBuilder = queryBuilder.Having(sq.Expr("MAX("+alias+".created_at_trunc) <= TIMESTAMP(convert_tz(?, '+00:00', ?))", node.Leaf.Action.TimeframeValues[1], timezone))
	} else if node.Leaf.Action.TimeframeOperator == "before_date" {
		queryBuilder = queryBuilder.Having(sq.Expr("MAX("+alias+".created_at_trunc) < TIMESTAMP(convert_tz(?, '+00:00', ?))", node.Leaf.Action.TimeframeValues[0], timezone))
	} else if node.Leaf.Action.TimeframeOperator == "after_date" {
		queryBuilder = queryBuilder.Having(sq.Expr("MAX("+alias+".created_at_trunc) > TIMESTAMP(convert_tz(?, '+00:00', ?))", node.Leaf.Action.TimeframeValues[0], timezone))
	}

	return queryBuilder.ToSql()
}

func filterAsQueryBuilder(alias string, filters []*entity.SegmentDimensionFilter, timezone string) (whereBuilder sq.And, err error) {

	whereBuilder = sq.And{}

	for _, filter := range filters {
		fieldName := alias + "." + filter.FieldName

		switch filter.FieldType {
		case "string":
			switch filter.Operator {
			case "equals":
				whereBuilder = append(whereBuilder, sq.Eq{fieldName: filter.StringValues[0]})
			case "not_equals":
				whereBuilder = append(whereBuilder, sq.NotEq{fieldName: filter.StringValues[0]})
			case "contains":
				contains := sq.Or{}
				for _, value := range filter.StringValues {
					contains = append(contains, sq.Like{fieldName: "%" + value + "%"})
				}
				whereBuilder = append(whereBuilder, contains)
			case "not_contains":
				whereBuilder = append(whereBuilder, sq.NotLike{fieldName: "%" + filter.StringValues[0] + "%"})
			case "is_set":
				// check not null
				whereBuilder = append(whereBuilder, sq.NotEq{fieldName: nil})
			case "is_not_set":
				// check null
				whereBuilder = append(whereBuilder, sq.Eq{fieldName: nil})
			default:
				return nil, eris.Errorf("unknown operator: %v", filter.Operator)
			}
		case "number":
			switch filter.Operator {
			case "equals":
				whereBuilder = append(whereBuilder, sq.Eq{fieldName: filter.NumberValues[0]})
			case "not_equals":
				whereBuilder = append(whereBuilder, sq.NotEq{fieldName: filter.NumberValues[0]})
			case "gt":
				whereBuilder = append(whereBuilder, sq.Gt{fieldName: filter.NumberValues[0]})
			case "lt":
				whereBuilder = append(whereBuilder, sq.Lt{fieldName: filter.NumberValues[0]})
			case "gte":
				whereBuilder = append(whereBuilder, sq.GtOrEq{fieldName: filter.NumberValues[0]})
			case "lte":
				whereBuilder = append(whereBuilder, sq.LtOrEq{fieldName: filter.NumberValues[0]})
			case "is_set":
				// check not null
				whereBuilder = append(whereBuilder, sq.NotEq{fieldName: nil})
			case "is_not_set":
				// check null
				whereBuilder = append(whereBuilder, sq.Eq{fieldName: nil})
			default:
				return nil, eris.Errorf("unknown operator: %v", filter.Operator)
			}
		case "time":
			switch filter.Operator {
			case "in_date_range":
				whereBuilder = append(whereBuilder, sq.And{
					sq.Expr("? >= TIMESTAMP(convert_tz(?, '+00:00', ?))", fieldName, filter.StringValues[0], timezone),
					sq.Expr("? <= TIMESTAMP(convert_tz(?, '+00:00', ?))", fieldName, filter.StringValues[1], timezone),
				})
			case "not_in_date_range":
				whereBuilder = append(whereBuilder, sq.Or{
					sq.Expr("? < TIMESTAMP(convert_tz(?, '+00:00', ?))", fieldName, filter.StringValues[0], timezone),
					sq.Expr("? > TIMESTAMP(convert_tz(?, '+00:00', ?))", fieldName, filter.StringValues[1], timezone),
				})

			case "before_date":
				whereBuilder = append(whereBuilder, sq.Expr("? < TIMESTAMP(convert_tz(?, '+00:00', ?))", fieldName, filter.StringValues[0], timezone))
			case "after_date":
				whereBuilder = append(whereBuilder, sq.Expr("? > TIMESTAMP(convert_tz(?, '+00:00', ?))", fieldName, filter.StringValues[0], timezone))
			case "is_set":
				// check not null
				whereBuilder = append(whereBuilder, sq.NotEq{fieldName: nil})
			case "is_not_set":
				// check null
				whereBuilder = append(whereBuilder, sq.Eq{fieldName: nil})
			default:
				return nil, eris.Errorf("unknown operator: %v", filter.Operator)
			}
		default:
			return nil, eris.Errorf("unknown field type: %v", filter.FieldType)
		}
	}

	return whereBuilder, nil
}

func (repo *RepositoryImpl) DeleteSegment(ctx context.Context, workspaceID string, segmentID string) (err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	_, err = conn.ExecContext(ctx, "UPDATE `segment` SET status = ? WHERE id = ?", entity.SegmentStatusDeleted, segmentID)

	if err != nil {
		return eris.Wrap(err, "DeleteSegment")
	}

	return
}

func (repo *RepositoryImpl) PreviewSegment(ctx context.Context, workspaceID string, parentSegmentID *string, tree *entity.SegmentTreeNode, timezone string) (count int64, sql string, args []interface{}, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	// build the query from the filter tree
	conditions, args, err := segmentNodeAsSQL(parentSegmentID, tree, timezone)

	if err != nil {
		return 0, "", nil, err
	}

	parentSegmentFilter := ""
	if parentSegmentID != nil {
		if *parentSegmentID != entity.SegmentAllUsersID {
			parentSegmentFilter = "u.is_authenticated = 1 AND"
			if *parentSegmentID == "anonymous" {
				parentSegmentFilter = "u.is_authenticated = 0 AND"
			}
		}
	}

	returnedSQL := fmt.Sprintf("%v (%v)", parentSegmentFilter, conditions)

	sql = fmt.Sprintf("SELECT count(u.id) FROM `user` u WHERE u.is_merged = 0 AND %v;", returnedSQL)

	// log.Printf("sql: %v\n", sql)
	// log.Printf("args: %+v\n", args)

	if err = sqlscan.Get(ctx, conn, &count, sql, args...); err != nil {
		err = eris.Wrapf(err, "SegmentPreview query: %v, args: %+v", sql, args)
		return
	}

	return count, returnedSQL, args, nil
}

func (repo *RepositoryImpl) ListSegments(ctx context.Context, workspaceID string, withUsersCount bool) (segments []*entity.Segment, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	segments = []*entity.Segment{}

	queryBuilder := sq.Select("segment.*").From("segment")

	if withUsersCount {
		queryBuilder = queryBuilder.LeftJoin("user_segment ON segment.id = user_segment.segment_id")
		queryBuilder = queryBuilder.GroupBy("segment.id")
		queryBuilder = queryBuilder.Column("COALESCE(COUNT(user_segment.user_id), 0) AS users_count")
	}

	// fetch segments
	query, args, err := queryBuilder.ToSql()

	if err != nil {
		err = eris.Wrapf(err, "ListSegments fetch query: %v, args: %+v", query, args)
		return
	}

	if err = sqlscan.Select(ctx, conn, &segments, query, args...); err != nil {
		err = eris.Wrapf(err, "ListSegments query: %v, args: %+v", query, args)
		return
	}

	// add a "_all" segment in the list
	allSegment := &entity.Segment{
		ID:     entity.SegmentAllUsersID,
		Name:   "All users",
		Status: entity.SegmentStatusActive,
		Color:  "default",
		// UsersCount: ,
	}

	for _, segment := range segments {
		if segment.ID == "anonymous" {
			allSegment.UsersCount += segment.UsersCount
		}
		if segment.ID == "authenticated" {
			allSegment.UsersCount += segment.UsersCount
		}
	}

	segments = append(segments, allSegment)

	return
}

func (repo *RepositoryImpl) InsertSegment(ctx context.Context, segment *entity.Segment, tx *sql.Tx) (err error) {

	query, args, err := sq.Insert("segment").Columns(
		"id",
		"name",
		"color",
		"parent_segment_id",
		"tree",
		"timezone",
		"version",
		"generated_sql",
		"generated_args",
		"status",
	).Values(
		segment.ID,
		segment.Name,
		segment.Color,
		segment.ParentSegmentID,
		segment.Tree,
		segment.Timezone,
		segment.Version,
		segment.GeneratedSQL,
		segment.GeneratedArgs,
		segment.Status,
	).ToSql()

	if err != nil {
		return eris.Wrapf(err, "InsertSegment build query for segment %+v\n", *segment)
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		if repo.IsDuplicateEntry(err) {
			return entity.ErrSegmentAlreadyExists
		}
		return eris.Wrapf(err, "InsertSegment exec query %v", query)
	}

	return
}

func (repo *RepositoryImpl) ActivateSegment(ctx context.Context, workspaceID string, segmentID string, segmentVersion int) (didActivate bool, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	result, err := conn.ExecContext(ctx, "UPDATE segment SET status = ? WHERE id = ? AND version = ?;", entity.SegmentStatusActive, segmentID, segmentVersion)

	if err != nil {
		return false, eris.Wrap(err, "ActivateSegment")
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		return false, eris.Wrap(err, "ActivateSegment")
	} else if rowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

func (repo *RepositoryImpl) GetSegment(ctx context.Context, workspaceID string, segmentID string) (segment *entity.Segment, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	segment = &entity.Segment{}

	if err = sqlscan.Get(ctx, conn, segment, "SELECT * FROM `segment` WHERE id = ?", segmentID); err != nil {
		return nil, eris.Wrap(err, "GetSegment")
	}

	return
}

func (repo *RepositoryImpl) UpdateSegment(ctx context.Context, segment *entity.Segment, tx *sql.Tx) (err error) {

	query, args, err := sq.Update("segment").
		Set("name", segment.Name).
		Set("color", segment.Color).
		Set("parent_segment_id", segment.ParentSegmentID).
		Set("tree", segment.Tree).
		Set("timezone", segment.Timezone).
		Set("version", segment.Version).
		Set("generated_sql", segment.GeneratedSQL).
		Set("generated_args", segment.GeneratedArgs).
		Set("status", segment.Status).
		Where(sq.Eq{"id": segment.ID}).
		ToSql()

	if err != nil {
		return eris.Wrapf(err, "UpdateSegment build query for segment %+v\n", *segment)
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "UpdateSegment exec query %v", query)
	}

	return
}
