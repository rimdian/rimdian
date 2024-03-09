package repository

// import (
// 	"context"
// 	"database/sql"

// 	sq "github.com/Masterminds/squirrel"
// 	"github.com/rimdian/rimdian/internal/api/dto"
// 	"github.com/rimdian/rimdian/internal/api/entity"
// 	"github.com/rotisserie/eris"
// )

// func (repo *RepositoryImpl) ListObservabilityIncidents(ctx context.Context, params *dto.AppObservabilityIncidentParams) (incidents []*entity.App_Observability_Incident, err error) {
// 	incidents = []*entity.App_Observability_Incident{}

// 	conn, err := repo.GetWorkspaceConnection(ctx, params.WorkspaceID)

// 	if err != nil {
// 		return incidents, eris.Wrapf(err, "ListObservabilityIncidents get workspace connection for workspace %+v\n", params.WorkspaceID)
// 	}

// 	defer conn.Close()

// 	queryBuilder := sq.Select("*").From("observability_incident").Limit(100)

// 	// filters
// 	if params.ID != "" {
// 		queryBuilder = queryBuilder.Where(sq.Eq{"id": params.ID})
// 	} else if params.CheckID != "" {
// 		queryBuilder = queryBuilder.Where(sq.Eq{"check_id": params.CheckID})
// 	}
// 	if params.IsClosed != nil && *params.IsClosed {
// 		queryBuilder = queryBuilder.Where(sq.Eq{"is_closed": true})
// 	}
// 	if params.IsClosed != nil && !*params.IsClosed {
// 		queryBuilder = queryBuilder.Where(sq.Eq{"is_closed": false})
// 	}
// 	queryBuilder.OrderBy("db_created_at DESC")

// 	// fetch
// 	query, args, err := queryBuilder.ToSql()

// 	if err != nil {
// 		return incidents, eris.Wrapf(err, "ListObservabilityIncidents %+v\n", *params)
// 	}

// 	// log.Println(query)
// 	if err = sqlscan.Select(ctx, conn, &incidents, query, args...); err != nil {
// 		return incidents, eris.Wrapf(err, "ListObservabilityIncidents select incidents for workspace %+v\n", params.WorkspaceID)
// 	}

// 	return incidents, nil
// }

// func (repo *RepositoryImpl) DeleteObservabilityCheck(ctx context.Context, workspaceID string, checkID string) error {

// 	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

// 	if err != nil {
// 		return err
// 	}

// 	defer conn.Close()

// 	_, err = conn.ExecContext(ctx, "DELETE FROM observability_check WHERE id = ?", checkID)

// 	if err != nil {
// 		return eris.Wrapf(err, "DeleteObservabilityCheck: %v", err)
// 	}

// 	_, err = conn.ExecContext(ctx, "DELETE FROM observability_incident WHERE check_id = ?", checkID)

// 	if err != nil {
// 		return eris.Wrapf(err, "DeleteObservabilityCheck: %v", err)
// 	}

// 	return nil
// }

// func (repo *RepositoryImpl) DeleteObservabilityIncident(ctx context.Context, workspaceID string, incidentID string) error {

// 	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

// 	if err != nil {
// 		return err
// 	}

// 	defer conn.Close()

// 	_, err = conn.ExecContext(ctx, "DELETE FROM observability_incident WHERE id = ?", incidentID)

// 	if err != nil {
// 		return eris.Wrapf(err, "DeleteObservabilityIncident: %v", err)
// 	}

// 	return nil
// }

// func (repo *RepositoryImpl) GetObservabilityCheck(ctx context.Context, workspaceID string, checkID string) (check *entity.App_Observability_Check, err error) {

// 	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

// 	if err != nil {
// 		return nil, err
// 	}

// 	defer conn.Close()

// 	check = &entity.App_Observability_Check{}

// 	err = sqlscan.Get(ctx, conn, check, "SELECT * FROM observability_check WHERE id = ? LIMIT 1", checkID)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, eris.Wrap(entity.ErrObservabilityCheckNotFound, "GetObservabilityCheck")
// 		} else {
// 			return nil, eris.Wrap(err, "GetObservabilityCheck error")
// 		}
// 	}

// 	return check, nil
// }

// func (repo *RepositoryImpl) ListObservabilityChecks(ctx context.Context, workspaceID string) (checks []*entity.App_Observability_Check, err error) {
// 	checks = []*entity.App_Observability_Check{}

// 	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

// 	if err != nil {
// 		return checks, eris.Wrapf(err, "ListObservabilityChecks get workspace connection for workspace %+v\n", workspaceID)
// 	}

// 	defer conn.Close()

// 	// log.Println(query)
// 	if err = sqlscan.Select(ctx, conn, &checks, "SELECT * FROM observability_check"); err != nil {
// 		return checks, eris.Wrapf(err, "ListObservabilityChecks select checks for workspace %+v\n", workspaceID)
// 	}

// 	return checks, nil
// }

// func (repo *RepositoryImpl) CreateObservabilityCheck(ctx context.Context, workspaceID string, check *entity.App_Observability_Check) error {

// 	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

// 	if err != nil {
// 		return eris.Wrapf(err, "CreateObservabilityCheck get workspace connection for workspace %+v\n", workspaceID)
// 	}

// 	defer conn.Close()

// 	query, args, err := sq.Insert("observability_check").Columns(
// 		"id",
// 		"name",
// 		"measure",
// 		"time_dimension",
// 		"filters",
// 		"rolling_window_value",
// 		"rolling_window_unit",
// 		"condition_type",
// 		"threshold_position",
// 		"threshold_value",
// 		"is_active",
// 		"next_run_at",
// 		"emails",
// 	).Values(
// 		check.ID,
// 		check.Name,
// 		check.Measure,
// 		check.TimeDimension,
// 		check.Filters,
// 		check.RollingWindowValue,
// 		check.RollingWindowUnit,
// 		check.ConditionType,
// 		check.ThresholdPosition,
// 		check.ThresholdValue,
// 		check.IsActive,
// 		check.NextRunAt,
// 		check.Emails,
// 	).ToSql()

// 	if err != nil {
// 		return eris.Wrapf(err, "CreateObservabilityCheck %+v\n", *check)
// 	}

// 	_, err = conn.ExecContext(ctx, query, args...)

// 	if err != nil {
// 		if repo.IsDuplicateEntry(err) {
// 			return entity.ErrObservabilityCheckAlreadyExists
// 		}
// 		return eris.Wrapf(err, "CreateObservabilityCheck exec query %v", query)
// 	}

// 	return nil
// }

// func (repo *RepositoryImpl) UpdateObservabilityCheck(ctx context.Context, workspaceID string, check *entity.App_Observability_Check) error {

// 	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

// 	if err != nil {
// 		return eris.Wrapf(err, "UpdateObservabilityCheck get workspace connection for workspace %+v\n", workspaceID)
// 	}

// 	defer conn.Close()

// 	q := sq.Update("observability_check").Where(sq.Eq{"id": check.ID}).
// 		Set("name", check.Name).
// 		Set("measure", check.Measure).
// 		Set("time_dimension", check.TimeDimension).
// 		Set("filters", check.Filters).
// 		Set("rolling_window_value", check.RollingWindowValue).
// 		Set("rolling_window_unit", check.RollingWindowUnit).
// 		Set("condition_type", check.ConditionType).
// 		Set("threshold_position", check.ThresholdPosition).
// 		Set("threshold_value", check.ThresholdValue).
// 		Set("is_active", check.IsActive).
// 		Set("next_run_at", check.NextRunAt).
// 		Set("emails", check.Emails)

// 	query, args, err := q.ToSql()

// 	if err != nil {
// 		return eris.Wrapf(err, "UpdateObservabilityCheck %+v\n", *check)
// 	}

// 	_, err = conn.ExecContext(ctx, query, args...)

// 	if err != nil {
// 		return eris.Wrapf(err, "UpdateObservabilityCheck exec query %v", query)
// 	}

// 	return nil
// }
