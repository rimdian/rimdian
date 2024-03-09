package repository

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) MatchSegmentUsers(ctx context.Context, workspaceID string, segment *entity.Segment, userIDs []string) (matchingUserIDs []*string, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	// build the query from the filter tree
	conditions, conditionsArgs, err := segmentNodeAsSQL(segment.ParentSegmentID, &segment.Tree, segment.Timezone)

	if err != nil {
		return nil, err
	}

	parentSegmentFilter := ""
	if segment.ParentSegmentID != nil {
		if *segment.ParentSegmentID != entity.SegmentAllUsersID {
			parentSegmentFilter = "u.is_authenticated = 1 AND"
			if *segment.ParentSegmentID == "anonymous" {
				parentSegmentFilter = "u.is_authenticated = 0 AND"
			}
		}
	}

	returnedSQL := fmt.Sprintf("%v (%v)", parentSegmentFilter, conditions)

	args := []interface{}{}

	in := fmt.Sprintf("u.id IN (%v) AND", sq.Placeholders(len(userIDs)))

	for _, userID := range userIDs {
		args = append(args, userID)
	}

	args = append(args, conditionsArgs...)

	query := fmt.Sprintf("SELECT u.id FROM `user` u WHERE u.is_merged = 0 AND %v %v;", in, returnedSQL)

	// log.Printf("MatchSegmentUsers: query: %v, args: %v", query, args)

	err = sqlscan.Select(ctx, conn, &matchingUserIDs, query, args...)

	if err != nil {
		err = eris.Wrap(err, "MatchSegmentUsers")
	}

	return
}

// insert the users from the user_segment_queue table in the user_segment table
func (repo *RepositoryImpl) EnterUserSegmentFromQueue(ctx context.Context, workspaceID string, segmentID string, segmentVersion int) (err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	query := `
		INSERT IGNORE INTO user_segment (user_id, segment_id)
		SELECT user_id, segment_id
		FROM user_segment_queue
		WHERE segment_id = ? AND segment_version = ? AND enters = 1;
	`

	if _, err = conn.ExecContext(ctx, query, segmentID, segmentVersion); err != nil {
		return eris.Wrap(err, "MatchSegmentUsers")
	}

	return
}

// remove the users from the user_segment_queue table in the user_segment table
func (repo *RepositoryImpl) ExitUserSegmentFromQueue(ctx context.Context, workspaceID string, segmentID string, segmentVersion int) (err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	query := `
		DELETE FROM user_segment
		WHERE user_id IN (
			SELECT user_id
			FROM user_segment_queue
			WHERE segment_id = ? AND segment_version = ? AND enters = 0
		);
	`

	if _, err = conn.ExecContext(ctx, query, segmentID, segmentVersion); err != nil {
		return eris.Wrap(err, "MatchSegmentUsers")
	}

	return
}
func (repo *RepositoryImpl) ListUserSegments(ctx context.Context, workspaceID string, userIDs []string, tx *sql.Tx) (userSegments []*entity.UserSegment, err error) {

	// init
	userSegments = []*entity.UserSegment{}

	query, args, err := sq.Select("*").From("user_segment").Where(sq.Eq{"user_id": userIDs}).ToSql()

	if err != nil {
		err = eris.Wrap(err, "ListUserSegments")
		return
	}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspaceID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		err = sqlscan.Select(ctx, conn, &userSegments, query, args...)
	} else {
		err = sqlscan.Select(ctx, tx, &userSegments, query, args...)
	}

	if err != nil {
		err = eris.Wrap(err, "ListUserSegments")
	}

	return
}

func (repo *RepositoryImpl) DeleteUserSegment(ctx context.Context, userID string, segmentID string, tx *sql.Tx) (err error) {

	_, err = tx.ExecContext(ctx, "DELETE FROM user_segment WHERE user_id = ? AND segment_id = ?", userID, segmentID)

	if err != nil {
		err = eris.Wrap(err, "DeleterUserSegment")
	}
	return
}

func (repo *RepositoryImpl) InsertUserSegment(ctx context.Context, userSegment *entity.UserSegment, tx *sql.Tx) (err error) {

	query, args, err := sq.Insert("user_segment").Columns(
		"user_id",
		"segment_id",
		// "segment_version",
	).Values(
		userSegment.UserID,
		userSegment.SegmentID,
		// userSegment.SegmentVersion,
	).ToSql()

	if err != nil {
		return eris.Wrapf(err, "InsertUserSegment build query for userSegment %+v\n", *userSegment)
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		if repo.IsDuplicateEntry(err) {
			return entity.ErrUserSegmentAlreadyExists
		}
		return eris.Wrapf(err, "InsertUserSegment exec query %v", query)
	}

	return
}
