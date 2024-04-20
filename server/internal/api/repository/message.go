package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) FindMessageByID(ctx context.Context, workspace *entity.Workspace, messageID string, userID string, tx *sql.Tx) (messageFound *entity.Message, err error) {

	messageFound = &entity.Message{}

	sql, args, err := sq.Select("*").From("message").Where(sq.Eq{"id": messageID, "user_id": userID}).ToSql()

	if err != nil {
		return nil, eris.Wrap(err, "FindMessageByID")
	}

	if tx == nil {

		conn, errConn := repo.GetWorkspaceConnection(ctx, workspace.ID)

		if errConn != nil {
			return nil, errConn
		}

		defer conn.Close()

		err = sqlscan.Get(ctx, conn, messageFound, sql, args...)
	} else {
		err = sqlscan.Get(ctx, tx, messageFound, sql, args...)
	}

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, err
		}
		return nil, eris.Wrap(err, "FindMessageByID")
	}

	return messageFound, nil
}

func (repo *RepositoryImpl) InsertMessage(ctx context.Context, message *entity.Message, tx *sql.Tx) (err error) {

	if message == nil {
		err = eris.New("message is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// INSERT
	q := sq.Insert("message").Columns(
		"id",
		"external_id",
		"user_id",
		"domain_id",
		"session_id",
		"created_at",
		"fields_timestamp",
		"channel",
		"is_inbound",
		"is_transactional",
		"message_template_id",
		"message_template_version",
		"subscription_list_id",
		"data",
		"status",
		"status_at",
		"comment",
		"retry_count",
		"is_sent",
		"sent_at",
		"scheduled_at",
		"delivered_at",
		"first_open_at",
		"first_click_at",
	).Values(
		message.ID,
		message.ExternalID,
		message.UserID,
		message.DomainID,
		message.SessionID,
		message.CreatedAt,
		message.FieldsTimestamp,
		message.Channel,
		message.IsInbound,
		message.IsTransactional,
		message.MessageTemplateID,
		message.MessageTemplateVersion,
		message.SubscriptionListID,
		message.Data,
		message.Status,
		message.StatusAt,
		message.Comment,
		message.RetryCount,
		message.IsSent,
		message.SentAt,
		message.ScheduledAt,
		message.DeliveredAt,
		message.FirstOpenAt,
		message.FirstClickAt,
	)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query insert message: %v\n", message)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		if repo.IsDuplicateEntry(errExec) {
			return eris.Wrap(ErrRowAlreadyExists, "InsertMessage")
		}

		err = eris.Wrap(errExec, "InsertMessage")
		return
	}

	message.DBCreatedAt = now
	message.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) UpdateMessage(ctx context.Context, message *entity.Message, tx *sql.Tx) (err error) {

	if message == nil {
		err = eris.New("message is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// UPDATE
	// specify sharding key to avoid deadlocks
	q := sq.Update("message").Where(sq.Eq{"user_id": message.UserID}).Where(sq.Eq{"id": message.ID}).
		Set("created_at", message.CreatedAt).
		Set("fields_timestamp", message.FieldsTimestamp).
		Set("channel", message.Channel).
		Set("is_inbound", message.IsInbound).
		Set("is_transactional", message.IsTransactional).
		Set("message_template_id", message.MessageTemplateID).
		Set("message_template_version", message.MessageTemplateVersion).
		Set("subscription_list_id", message.SubscriptionListID).
		Set("data", message.Data).
		Set("status", message.Status).
		Set("status_at", message.StatusAt).
		Set("comment", message.Comment).
		Set("retry_count", message.RetryCount).
		Set("is_sent", message.IsSent).
		Set("sent_at", message.SentAt).
		Set("scheduled_at", message.ScheduledAt).
		Set("delivered_at", message.DeliveredAt).
		Set("first_open_at", message.FirstOpenAt).
		Set("first_click_at", message.FirstClickAt)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update message: %v\n", message)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		err = eris.Wrap(errExec, "UpdateMessage")
		return
	}

	message.DBUpdatedAt = now

	return
}

// clones messages from a user to another user with
// because the shard key "user_id" is immutable, we can't use an UPDATE
// we have to INSERT FROM SELECT + DELETE
func (repo *RepositoryImpl) MergeUserMessages(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error) {

	// find eventual extra columns for the message table
	messageCustomColumns := workspace.FindExtraColumnsForItemKind(entity.ItemKindMessage)

	messageStruct := entity.Message{}
	columns := entity.GetNotComputedDBColumnsForObject(messageStruct, entity.MessageComputedFields, messageCustomColumns)

	// replace dynamically the user_id+merged_from_user_id on the SELECT statement
	selectedColumns := []string{}
	for _, col := range columns {

		if col == "user_id" {
			selectedColumns = append(selectedColumns, fmt.Sprintf("'%v' as user_id", toUserID))
		} else if col == "merged_from_user_id" {
			selectedColumns = append(selectedColumns, fmt.Sprintf("'%v' as merged_from_user_id", fromUserID))
		} else {
			selectedColumns = append(selectedColumns, col)
		}
	}

	query := fmt.Sprintf(`
		INSERT IGNORE INTO message (%v) 
		SELECT %v FROM message 
		WHERE user_id = ?
	`, strings.Join(columns, ", "), strings.Join(selectedColumns, ", "))

	// log.Println(query)

	if _, err := tx.ExecContext(ctx, query, fromUserID); err != nil {
		return eris.Wrap(err, "MergeUserMessages")
	}

	return nil
}
