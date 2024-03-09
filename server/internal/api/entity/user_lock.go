package entity

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// the user lock is used to prevent multiple users from accessing the same resource at the same time
// it leverages a rowstore table to store the locks

type UsersLock struct {
	Mu       sync.Mutex
	Acquired map[string]bool
	UserIDs  []string
	LockedAt int64
	Token    string
}

func (lock *UsersLock) AddUser(userID string) {
	// app items who don't have a speicifc user_id (=none) are not locked
	if userID == "" || userID == None {
		return
	}

	lock.Mu.Lock()
	defer lock.Mu.Unlock()

	if _, ok := lock.Acquired[userID]; ok {
		return
	}

	lock.Acquired[userID] = false
	lock.UserIDs = append(lock.UserIDs, userID)
}

func (lock *UsersLock) RemoveUser(userID string) {
	lock.Mu.Lock()
	defer lock.Mu.Unlock()

	for i, uID := range lock.UserIDs {
		if uID == userID {
			lock.UserIDs = append(lock.UserIDs[:i], lock.UserIDs[i+1:]...)
			delete(lock.Acquired, userID)
			break
		}
	}
}

func (lock *UsersLock) SetAcquired(userID string, value bool) {
	lock.Mu.Lock()
	defer lock.Mu.Unlock()

	lock.Acquired[userID] = value
}

func (lock *UsersLock) ListNonAcquiredUserIDs() []string {
	lock.Mu.Lock()
	defer lock.Mu.Unlock()

	var userIDs []string
	for userID, acquired := range lock.Acquired {
		if !acquired {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs
}

func (lock *UsersLock) ListAcquiredUserIDs() []string {
	lock.Mu.Lock()
	defer lock.Mu.Unlock()

	var userIDs []string
	for userID, acquired := range lock.Acquired {
		if acquired {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs
}

func NewUsersLock() *UsersLock {
	// generate a uuidv4
	id, _ := uuid.NewRandom()

	return &UsersLock{
		Acquired: make(map[string]bool),
		UserIDs:  []string{},
		Token:    id.String(),
	}
}

// lock a user for 60secs (=task/data-import deadline)
// to ensure serialization of users data processing
type UserLock struct {
	UserID   string    `db:"user_id"`
	LockedAt time.Time `db:"locked_at"`
	Token    string    `db:"token"` // used to make sure the lock is not released by another worker
}

const UserLockSchema = `
CREATE ROWSTORE TABLE IF NOT EXISTS user_lock (
	user_id CHAR(40) NOT NULL,
	locked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	token CHAR(36) NOT NULL,
	PRIMARY KEY (user_id),
	SHARD KEY (user_id)
  ) COLLATE utf8mb4_general_ci;`

const UserLockSchemaMYSQL = `
CREATE TABLE IF NOT EXISTS user_lock (
	user_id CHAR(40) NOT NULL,
	locked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	token CHAR(36) NOT NULL,
	PRIMARY KEY (user_id)
	-- SHARD KEY (user_id)
  ) COLLATE utf8mb4_general_ci;`
