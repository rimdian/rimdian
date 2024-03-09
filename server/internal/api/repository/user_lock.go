package repository

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// a user lock expires after 30 secs == the tasks/data import deadline
// the locked_at field it as the microsecond

func (repo *RepositoryImpl) EnsureUsersLock(ctx context.Context, workspaceID string, lock *entity.UsersLock, withRetry bool) (mainErr error) {

	// acquire users locks in goroutine

	ids := lock.ListNonAcquiredUserIDs()

	if len(ids) == 0 {
		return nil
	}

	retryCount := 10
	// we can disable retries in dev mode
	if !withRetry {
		retryCount = 0
	}

	remaining := retryCount

	// acquire a user lock or retry every 500ms for 5 seconds
	retry := retrier.New(retrier.ConstantBackoff(retryCount, 500*time.Millisecond), nil)

	errRetry := retry.Run(func() error {

		code, err := repo.RunInTransactionForWorkspace(ctx, workspaceID, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

			// try to insert the lock with provided token
			squirrelQuery := squirrel.Insert("user_lock").Columns("user_id", "token")

			for _, userID := range ids {
				squirrelQuery = squirrelQuery.Values(userID, lock.Token)
			}

			query, args, err := squirrelQuery.ToSql()

			if err != nil {
				return 500, eris.Wrap(err, "AcquireUserLock")
			}

			var insertResult sql.Result
			insertResult, insertErr := tx.ExecContext(ctx, query, args...)

			if insertErr != nil {

				// a lock exists or is expired
				if repo.IsDuplicateEntry(insertErr) {

					// try to remove the expired lock at the last retry
					if remaining == 1 {
						squirrelQuery := squirrel.Delete("user_lock").Where("TIMESTAMPADD(SECOND, 30, locked_at) < NOW()").Where(squirrel.Eq{"user_id": ids})

						query, args, err := squirrelQuery.ToSql()

						if err != nil {
							return 500, eris.Wrapf(err, "AcquireUserLock query %v", query)
						}

						var delResult sql.Result
						delResult, delErr := tx.ExecContext(ctx, query, args...)

						if delErr != nil {
							return 500, eris.Wrap(delErr, "AcquireUserLock delete expired lock")
						}

						delRowsAffected, delCounterr := delResult.RowsAffected()

						if delCounterr != nil {
							return 500, eris.Wrap(delCounterr, "AcquireUserLock")
						}

						if delRowsAffected > 0 {
							log.Println("EXPIRED lock deleted")
							// return nil error to commit the transaction and remove the lock
							return http.StatusConflict, nil
						}
					}

					return 500, ErrAcquireUserLockFailed
				}

				return 500, eris.Wrap(insertErr, "AcquireUserLock")
			}

			// check if lock was inserted
			rowsAffected, countErr := insertResult.RowsAffected()

			if countErr != nil {
				return 500, eris.Wrap(countErr, "AcquireUserLock")
			}

			if rowsAffected == 0 {
				return 500, eris.Wrap(ErrAcquireUserLockFailed, "no rows could be inserted")
			}

			return 200, nil
		})

		remaining--

		// lock has expired and been removed, we need to retry acquiring the lock
		if code == http.StatusConflict {
			return eris.Wrapf(ErrAcquireUserLockFailed, "status code %v", code)
		}

		return err
	})

	// failed to lock them all, cancel operation
	if errRetry != nil {
		mainErr = eris.Wrap(errRetry, "EnsureUsersLock error")
		repo.ReleaseUsersLock(workspaceID, lock)
		return
	}

	for _, userID := range ids {
		lock.SetAcquired(userID, true)
	}

	return nil
}

func (repo *RepositoryImpl) ReleaseUsersLock(workspaceID string, lock *entity.UsersLock) error {

	userIds := lock.ListAcquiredUserIDs()

	if len(userIds) == 0 {
		return nil
	}

	ctx := context.Background()
	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return err
	}

	defer conn.Close()

	for _, userID := range userIds {

		// delete the user lock for the given token
		query := `DELETE FROM user_lock WHERE user_id = ? AND token = ?`

		result, errDel := conn.ExecContext(ctx, query, userID, lock.Token)

		if errDel != nil {
			return eris.Wrap(errDel, "ReleaseUserLock")
		}

		// check if lock was deleted
		rowsAffected, err := result.RowsAffected()

		if err != nil {
			return eris.Wrap(err, "ReleaseUserLock")
		}

		lock.SetAcquired(userID, false)

		if rowsAffected == 0 {
			// return eris.Errorf("no lock found for user %v at %v", userID, lock.Token)
			return eris.New("no lock found")
			// log.Println("NOT FOUND")
		}
	}

	return nil
}
