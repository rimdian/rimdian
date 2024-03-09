package integrationtests_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/package/api"
	"github.com/sirupsen/logrus"
)

func TestIntegration_UserLock(t *testing.T) {

	// instantiate a repo
	repo, _, err := api.NewRepository(context.Background(), logrus.New())

	if err != nil {
		t.Fatalf("cannot create api server: %v", err)
	}

	t.Run("should acquire a user lock once", func(t *testing.T) {

		ctx := context.Background()
		userID := "test"
		workspaceID := "acme_testing"

		_, err := repo.RunInTransactionForWorkspace(ctx, workspaceID, func(ctx context.Context, tx *sql.Tx) (code int, err error) {
			// clean all user locks
			query := `DELETE FROM user_lock`
			_, err = tx.ExecContext(ctx, query)

			if err != nil {
				return 500, err
			}

			return 200, nil
		})

		if err != nil {
			t.Fatalf("cannot clean all user locks: %v", err)
		}

		lock := entity.NewUsersLock()
		lock.AddUser(userID)
		lock.AddUser(userID)
		lock.AddUser(userID)

		defer func() {
			if err := repo.ReleaseUsersLock(workspaceID, lock); err != nil {
				t.Fatalf("cannot release user lock: %v", err)
			}
		}()

		err = repo.EnsureUsersLock(ctx, workspaceID, lock, true)

		if err != nil {
			t.Errorf("cannot create users lock: %v", err)
		}

		// pretend we removed the lock, and try to acquire it again
		newLock := entity.NewUsersLock()
		newLock.AddUser(userID)

		err = repo.EnsureUsersLock(ctx, workspaceID, newLock, false)

		if err == nil {
			t.Errorf("expected error when acquiring a user lock twice, second token is %v", newLock.Token)
		}
	})
}
