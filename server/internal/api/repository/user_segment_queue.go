package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// matches the users against the segment and insert the results in the user_segment_queue table
func (repo *RepositoryImpl) EnqueueMatchingSegmentUsers(ctx context.Context, workspaceID string, segment *entity.Segment) (entersCount int, exitCount int, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	// insert the matching users in the user_segment_queue table
	// conditions example:
	// u.is_authenticated = 1 AND ((u.id IN (SELECT id FROM `user` WHERE (country = ?))))

	// insert ignore to avoid duplicate entries errors in case of task retry
	// also include already matching users, as we want to compare the new results with the old ones
	// in order to find the exiting users
	// the worker_id is a random int between 1 and 10, so that we can have 10 workers running in parallel

	query := fmt.Sprintf(`
		INSERT IGNORE INTO user_segment_queue (user_id, segment_id, segment_version, enters, worker_id)
		SELECT u.id, '%v', %v, 1, FLOOR(RAND() * 10) + 1
		FROM user u
		WHERE u.is_merged = 0 AND %v;
	`, segment.ID, segment.Version, segment.GeneratedSQL)

	if _, err = conn.ExecContext(ctx, query, segment.GeneratedArgs...); err != nil {
		return 0, 0, eris.Wrapf(err, "MatchSegmentUsers query: %v", query)
	}

	// find the users from the user_segment table that are not in the user_segment_queue table
	query = `
		INSERT IGNORE INTO user_segment_queue (user_id, segment_id, segment_version, enters)
		SELECT user_segment.user_id, ?, ?, 0
		FROM user_segment
		LEFT JOIN user_segment_queue ON user_segment.user_id = user_segment_queue.user_id
		WHERE 
			user_segment.segment_id = ? AND 
			user_segment_queue.user_id IS NULL
		;`

	args := []interface{}{
		segment.ID,
		segment.Version,
		segment.ID,
	}

	// log.Printf("query: %v\n", query)

	if _, err = conn.ExecContext(ctx, query, args...); err != nil {
		return 0, 0, eris.Wrap(err, "MatchSegmentUsers")
	}

	// join the user_segment table with the user_segment_queue table
	// to find the users that already entered the segment in order to deleter them from the user_segment_queue table
	query = `
		DELETE user_segment_queue
		FROM user_segment_queue
		INNER JOIN user_segment ON user_segment_queue.user_id = user_segment.user_id
		WHERE
			user_segment.segment_id = ? AND
			user_segment_queue.enters = 1;
	`

	args = []interface{}{
		segment.ID,
	}

	// log.Printf("query: %v\n", query)

	if _, err = conn.ExecContext(ctx, query, args...); err != nil {
		return 0, 0, eris.Wrap(err, "MatchSegmentUsers")
	}

	// count the users that entered and exited the segment
	query = `
		SELECT
			COALESCE(SUM(CASE WHEN enters = 1 THEN 1 ELSE 0 END), 0) AS enters_count,
			COALESCE(SUM(CASE WHEN enters = 0 THEN 1 ELSE 0 END), 0) AS exit_count
		FROM user_segment_queue
		WHERE segment_id = ? AND segment_version = ?;
	`

	args = []interface{}{
		segment.ID,
		segment.Version,
	}

	// log.Printf("query: %v\n", query)

	result := struct {
		Enters int `db:"enters_count"`
		Exits  int `db:"exit_count"`
	}{}

	if err = sqlscan.Get(ctx, conn, &result, query, args...); err != nil {
		return 0, 0, eris.Wrap(err, "MatchSegmentUsers")
	}

	return result.Enters, result.Exits, nil
}

func (repo *RepositoryImpl) DeleteUserSegmentQueueRow(ctx context.Context, workspaceID string, segmentID string, segmentVersion int, userID string, tx *sql.Tx) (err error) {

	result, err := tx.ExecContext(ctx, "DELETE FROM user_segment_queue WHERE user_id = ? AND segment_id = ? AND segment_version = ?", userID, segmentID, segmentVersion)

	if err != nil {
		return eris.Wrap(err, "DeleteUserSegmentQueueRow")
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return eris.Wrap(err, "DeleteUserSegmentQueueRow")
	}

	if rowsAffected == 0 {
		log.Printf("DeleteUserSegmentQueueRow: no rows affected, user_id: %v, segment_id: %v, segment_version: %v", userID, segmentID, segmentVersion)
	}

	return
}

// cleans the user_segment_queue table before recomputing the segment
func (repo *RepositoryImpl) ClearUserSegmentQueue(ctx context.Context, workspaceID string, segmentID string, segmentVersion int) (err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	if _, err = conn.ExecContext(ctx, "DELETE FROM user_segment_queue WHERE segment_id = ? AND segment_version <= ?", segmentID, segmentVersion); err != nil {
		return eris.Wrap(err, "ClearUserSegmentQueue")
	}

	return
}

// func (repo *RepositoryImpl) GetUserSegmentQueueRowsForWorker(ctx context.Context, workspaceID string, segmentID string, segmentVersion int, workerID int, limit int) (rows []*entity.UserSegmentQueue, err error) {

// 	var conn *sql.Conn

// 	conn, err = repo.GetWorkspaceConnection(ctx, workspaceID)

// 	if err != nil {
// 		return nil, err
// 	}

// 	defer conn.Close()

// 	builder := sq.Select("*").From("user_segment_queue").Where(sq.Eq{
// 		"segment_id":      segmentID,
// 		"segment_version": segmentVersion,
// 	}).Limit(uint64(limit))

// 	// don't mind the worker if it's 0, its the main thread and we don't use workers yet
// 	if workerID != 0 {
// 		builder = builder.Where(sq.Eq{"worker_id": workerID})
// 	}

// 	query, args, err := builder.ToSql()

// 	if err != nil {
// 		return nil, eris.Wrap(err, "GetUserSegmentQueueRowsForWorker")
// 	}

// 	err = sqlscan.Select(ctx, conn, &rows, query, args...)

// 	if err != nil {
// 		return nil, eris.Wrap(err, "GetUserSegmentQueueRowsForWorker")
// 	}

// 	return
// }
