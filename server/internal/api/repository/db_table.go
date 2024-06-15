package repository

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/feiin/sqlstring"
	"github.com/forPelevin/gomoji"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

var (
	ErrExistingTableIsDifferent = eris.New("existing table structure is different")
)

func tableColumnDDL(column *entity.TableColumn) (ddl string, err error) {

	ddl = column.Name

	switch column.Type {

	case entity.ColumnTypeBoolean:
		ddl = ddl + " BOOL"
		if column.IsRequired {
			ddl = ddl + " NOT NULL"
		}
		if column.DefaultBoolean != nil && *column.DefaultBoolean {
			ddl = ddl + " DEFAULT TRUE"
		}
		if column.DefaultBoolean != nil && !*column.DefaultBoolean {
			ddl = ddl + " DEFAULT FALSE"
		}

	case entity.ColumnTypeNumber:
		ddl = ddl + " DOUBLE"
		if column.IsRequired {
			ddl = ddl + " NOT NULL"
		}
		if column.DefaultNumber != nil {
			ddl = fmt.Sprintf("%v DEFAULT %f", ddl, *column.DefaultNumber)
		}
	case entity.ColumnTypeDate:
		ddl = ddl + " DATE"
		if column.IsRequired {
			ddl = ddl + " NOT NULL"
		}
		if column.DefaultDate != nil {
			ddl = fmt.Sprintf("%v DEFAULT %v", ddl, *column.DefaultDate)
		}
	case entity.ColumnTypeDatetime:
		ddl = ddl + " DATETIME"
		if column.IsRequired {
			ddl = ddl + " NOT NULL"
		}
		if column.DefaultDateTime != nil {
			ddl = fmt.Sprintf("%v DEFAULT %v", ddl, *column.DefaultDateTime)
		}
	case entity.ColumnTypeTimestamp:
		ddl = ddl + " TIMESTAMP"
		if column.Size != nil && *column.Size == 6 {
			ddl = ddl + "(6)"
		}
		if column.IsRequired {
			ddl = ddl + " NOT NULL"
		}
		if column.DefaultTimestamp != nil {
			ddl = fmt.Sprintf("%v DEFAULT %v", ddl, *column.DefaultTimestamp)
		}
		if column.ExtraDefinition != nil {
			ddl = ddl + " " + *column.ExtraDefinition
		}
	case entity.ColumnTypeVarchar:
		ddl = ddl + fmt.Sprintf(" VARCHAR(%v) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", *column.Size)
		if column.IsRequired {
			ddl = ddl + " NOT NULL"
		}
		if column.DefaultText != nil {
			*column.DefaultText = gomoji.RemoveEmojis(*column.DefaultText)
			ddl = fmt.Sprintf("%v DEFAULT %v", ddl, sqlstring.Escape(*column.DefaultText))
		}
	case entity.ColumnTypeLongText:
		ddl = ddl + " TEXT"
		if column.IsRequired {
			ddl = ddl + " NOT NULL"
		}
	case entity.ColumnTypeJSON:
		ddl = ddl + " JSON"
		if column.IsRequired {
			ddl = ddl + " NOT NULL"
		}
		if column.DefaultJSON != nil {
			ddl = fmt.Sprintf("%v DEFAULT %v", ddl, sqlstring.Escape(*column.DefaultJSON))
		}
	default:
		return "", entity.ErrTableColumnTypeNotImplemented
	}

	return ddl, nil
}

func (repo *RepositoryImpl) RenameTable(ctx context.Context, workspaceID string, tableName string, newName string) (err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %v RENAME TO %v", tableName, newName))

	if err != nil {
		return eris.Wrapf(err, "failed to rename table %v", tableName)
	}

	return nil
}

func (repo *RepositoryImpl) DeleteTable(ctx context.Context, workspaceID string, tableName string) (err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %v", tableName))

	if err != nil {
		return eris.Wrapf(err, "failed to delete custom table %v", tableName)
	}

	return nil
}

func (repo *RepositoryImpl) CreateTable(ctx context.Context, workspace *entity.Workspace, table *entity.AppTableManifest) error {

	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

	if err != nil {
		return err
	}

	defer conn.Close()

	columns := []string{}

	// generate DDL for each column
	for _, col := range table.Columns {

		// don't install updated_at column, it's used by the API to merge fields only
		if col.Name == "updated_at" {
			continue
		}

		ddl, err := tableColumnDDL(col)
		if err != nil {
			return err
		}

		if repo.Config.DB_TYPE == "singlestore" && table.TimeSeriesColumn != nil && *table.TimeSeriesColumn == col.Name {
			ddl = ddl + " SERIES TIMESTAMP"
		}

		columns = append(columns, ddl)
	}

	query := fmt.Sprintf(`
		%v,
		PRIMARY KEY(%v)
		%v
	`, strings.Join(columns, ", "), strings.Join(table.UniqueKey, ", "), table.Indexes.ToDDL())

	// log.Printf("CreateAppTable: %v", query)

	if repo.Config.DB_TYPE == "singlestore" {
		query = fmt.Sprintf(`
		%v,
		SORT KEY(%v),
		SHARD KEY(%v)
		`, query, strings.Join(table.SortKey, ", "), strings.Join(table.ShardKey, ", "))
	}

	// naive SQL injection prevention
	query = strings.ReplaceAll(query, ";", "")
	query = strings.ReplaceAll(query, "DELETE", "")
	query = strings.ReplaceAll(query, "SELECT", "")

	query = fmt.Sprintf("CREATE TABLE %v (%v);", table.Name, query)

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "CreateAppTable")
	}

	if _, err := tx.ExecContext(ctx, query); err != nil {
		shouldRollback := true

		// check if error Error 1050: Table already exists
		if strings.Contains(err.Error(), "Error 1050") {
			// check that the table is the same
			err = repo.IsExistingTableTheSame(ctx, workspace.ID, table)

			if err == nil {
				shouldRollback = false
			} else {
				err = eris.Wrapf(err, "existing table is different: %v", table.Name)
			}
		}

		if shouldRollback {
			if rollErr := tx.Rollback(); rollErr != nil {
				log.Printf("rollback error %v", rollErr)
			}
			return eris.Wrapf(err, "CreateAppTable, query: %v", query)
		}
	}

	// update workspace
	if err := repo.UpdateWorkspace(ctx, workspace, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "CreateAppTable")
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "CreateAppTable")
	}

	return nil
}

func (repo *RepositoryImpl) MigrateTable(ctx context.Context, workspace *entity.Workspace, table *entity.AppTableManifest, suffix string) error {

	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

	if err != nil {
		return err
	}

	defer conn.Close()

	// rename old table to _migrated_YYYYMMDD_HHMMSS

	// create new table
	// TODO

	// copy data from old table to new table
	// remove suffix from new table

	return nil
}

func (repo *RepositoryImpl) IsExistingTableTheSame(ctx context.Context, workspaceID string, tableManifest *entity.AppTableManifest) error {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return err
	}

	defer conn.Close()

	// TODO: only fetch schema_info for the column we are interested in to opitmize
	tables, err := repo.ShowTables(ctx, workspaceID)

	if err != nil {
		return eris.Wrap(err, "IsExistingTableTheSame")
	}

	// find table in tables
	var existingTableInfo *entity.TableInformationSchema

	for _, table := range tables {
		if table.Name == tableManifest.Name {
			existingTableInfo = table
			break
		}
	}

	if existingTableInfo == nil {
		return eris.Errorf("exixting table information schema %v not found", tableManifest.Name)
	}

	// check if manifest columns are the same
	for _, col := range tableManifest.Columns {

		// don't install updated_at column, it's used by the API to merge fields only
		if col.Name == "updated_at" || col.Name == "db_created_at" || col.Name == "db_updated_at" {
			continue
		}

		var existingColumn *entity.ColumnInformationSchema

		for _, existingCol := range existingTableInfo.Columns {
			if existingCol.Name == col.Name {
				existingColumn = existingCol
				break
			}
		}

		if existingColumn == nil {
			return eris.Errorf("column %v not found in existing table", col.Name)
		}

		switch col.Type {

		case entity.ColumnTypeBoolean:
			if existingColumn.ColumnType != "tinyint(1)" {
				return eris.Errorf("column %v type boolean does not match existing type boolean", col.Name)
			}
			// has a default value ?
			if col.DefaultBoolean != nil {
				if existingColumn.DefaultValue == nil {
					return eris.Errorf("column %v type boolean default value does not match", col.Name)
				}
				if *col.DefaultBoolean && *existingColumn.DefaultValue != "1" {
					return eris.Errorf("column %v type boolean default value does not match", col.Name)
				}
				if !*col.DefaultBoolean && *existingColumn.DefaultValue != "0" {
					return eris.Errorf("column %v type boolean default value does not match", col.Name)
				}
			}
		case entity.ColumnTypeNumber:
			if existingColumn.DataType != "double" {
				return eris.Errorf("column %v type number does not match existing type number", col.Name)
			}
			// has a default value ?
			if col.DefaultNumber != nil {
				if existingColumn.DefaultValue == nil {
					return eris.Errorf("column %v type number default value does not match", col.Name)
				}
				// requires a float comparison
				// if *col.DefaultNumber != *existingColumn.DefaultValue {
				// 	return eris.Errorf("column %v type number default value does not match", col.Name)
				// }
			}
		case entity.ColumnTypeDate:
			if existingColumn.DataType != "date" {
				return eris.Errorf("column %v type date does not match existing type date", col.Name)
			}
			// has a default value ?
			if col.DefaultDate != nil {
				if existingColumn.DefaultValue == nil {
					return eris.Errorf("column %v type date default value does not match", col.Name)
				}
				// requires a date comparison
				// if *col.DefaultDate != *existingColumn.DefaultValue {
				// 	return eris.Errorf("column %v type date default value does not match", col.Name)
				// }
			}
		case entity.ColumnTypeDatetime:
			if existingColumn.DataType != "datetime" {
				return eris.Errorf("column %v type datetime does not match existing type datetime", col.Name)
			}
			// has a default value ?
			if col.DefaultDateTime != nil {
				if existingColumn.DefaultValue == nil {
					return eris.Errorf("column %v type datetime default value does not match", col.Name)
				}
				// requires a datetime comparison
				// if *col.DefaultDatetime != *existingColumn.DefaultValue {
				// 	return eris.Errorf("column %v type datetime default value does not match", col.Name)
				// }
			}
		case entity.ColumnTypeTimestamp:
			if existingColumn.DataType != "timestamp" {
				return eris.Errorf("column %v type timestamp does not match existing type timestamp", col.Name)
			}
			// has a default value ?
			if col.DefaultTimestamp != nil {
				if existingColumn.DefaultValue == nil {
					return eris.Errorf("column %v type timestamp default value does not match", col.Name)
				}
				// requires a timestamp comparison
				// if *col.DefaultTimestamp != *existingColumn.DefaultValue {
				// 	return eris.Errorf("column %v type timestamp default value does not match", col.Name)
				// }
			}
		case entity.ColumnTypeVarchar:

			if existingColumn.DataType != "varchar" {
				return eris.Errorf("column %v type varchar does not match existing type varchar", col.Name)
			}
			// has a default value ?
			if col.DefaultText != nil {
				if existingColumn.DefaultValue == nil {
					return eris.Errorf("column %v type varchar default value does not match", col.Name)
				}
				if *col.DefaultText != *existingColumn.DefaultValue {
					return eris.Errorf("column %v type varchar default value does not match", col.Name)
				}
			}

		case entity.ColumnTypeLongText:
			if existingColumn.DataType != "text" {
				return eris.Errorf("column %v type longtext does not match existing type longtext", col.Name)
			}
			// has a default value ?
			if col.DefaultText != nil {
				if existingColumn.DefaultValue == nil {
					return eris.Errorf("column %v type longtext default value does not match", col.Name)
				}
				if *col.DefaultText != *existingColumn.DefaultValue {
					return eris.Errorf("column %v type longtext default value does not match", col.Name)
				}
			}

		case entity.ColumnTypeJSON:
			if existingColumn.DataType != "json" {
				return eris.Errorf("column %v type json does not match existing type json", col.Name)
			}
			// has a default value ?
			if col.DefaultJSON != nil {
				if existingColumn.DefaultValue == nil {
					return eris.Errorf("column %v type json default value does not match", col.Name)
				}
				// requires a json comparison
				// if *col.DefaultJSON != *existingColumn.DefaultValue {
				// 	return eris.Errorf("column %v type json default value does not match", col.Name)
				// }
			}
		default:
		}

		if existingColumn.Nullable != "NO" && col.IsRequired {
			return eris.Errorf("column %v is_required %v does not match existing is_required %v", col.Name, col.IsRequired, existingColumn.Nullable)
		}

		if col.ExtraDefinition != nil && *col.ExtraDefinition != existingColumn.Extra {
			return eris.Errorf("column %v extra_definition %v does not match existing extra_definition %v", col.Name, col.ExtraDefinition, existingColumn.Extra)
		}
	}

	return nil
}

// func (repo *RepositoryImpl) DeleteAppTable(ctx context.Context, workspace *entity.Workspace, customTable *entity.AppTable) error {

// 	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

// 	if err != nil {
// 		return err
// 	}

// 	defer conn.Close()

// 	tx, err := conn.BeginTx(ctx, nil)
// 	if err != nil {
// 		return eris.Wrap(err, "DeleteAppTable")
// 	}

// 	query := fmt.Sprintf("ALTER TABLE `%v` DROP COLUMN %v", customTable.Name, customTable.Name)

// 	_, err = tx.ExecContext(ctx, query)

// 	// ignore if table doesnt exist
// 	if err != nil && !(strings.Contains(err.Error(), "Error 1146: Table") && strings.Contains(err.Error(), "doesn't exist")) {
// 		tx.Rollback()
// 		return eris.Wrapf(err, "DeleteAppTable, query: %v", query)
// 	}

// 	// update workspace
// 	if err := repo.UpdateWorkspace(ctx, workspace, tx); err != nil {
// 		tx.Rollback()
// 		return eris.Wrap(err, "DeleteAppTable")
// 	}

// 	if err := tx.Commit(); err != nil {
// 		return eris.Wrap(err, "DeleteAppTable")
// 	}

// 	return nil
// }
