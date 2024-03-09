package entity

import (
	"time"

	"github.com/rotisserie/eris"
)

var (
	ErrUserSegmentAlreadyExists = eris.New("user segment already exists")
)

type UserSegment struct {
	UserID    string `db:"user_id" json:"user_id"`
	SegmentID string `db:"segment_id" json:"segment_id"`
	// SegmentVersion int    `db:"segment_version" json:"segment_version"`
	// EnterAt      time.Time  `db:"enter_at" json:"enter_at"`
	// EnterAtTrunc time.Time  `db:"enter_at_trunc" json:"enter_at_trunc"`
	// ExitAt       *time.Time `db:"exit_at" json:"exit_at,omitempty"`
	// OutdatedAt   *time.Time `db:"outdated_at" json:"outdated_at,omitempty"`
	DBCreatedAt time.Time `db:"db_created_at" json:"db_created_at"`
	// DBUpdatedAt time.Time `db:"db_updated_at" json:"db_updated_at"`
	// joined server-side
	// Segment *Segment `db:"-" json:"segment,omitempty"`
}

func NewUserSegment(userID string, segmentID string) *UserSegment {
	return &UserSegment{
		UserID:    userID,
		SegmentID: segmentID,
		// SegmentVersion: segmentVersion,
		// EnterAt:      enterAt,
		// EnterAtTrunc: enterAt.Truncate(time.Hour),
		DBCreatedAt: time.Now(),
		// DBUpdatedAt: time.Now(),
	}
}

var UserSegmentSchema string = `CREATE TABLE IF NOT EXISTS user_segment (
	user_id VARCHAR(64) NOT NULL,
	segment_id VARCHAR(64) NOT NULL,
	-- segment_version INT NOT NULL,
	-- enter_at DATETIME(6) NOT NULL,
	-- enter_at_trunc AS DATE_TRUNC('hour', enter_at) PERSISTED DATETIME,
	-- exit_at DATETIME,
	-- outdated_at DATETIME,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	-- db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

	SORT KEY (db_created_at),
	PRIMARY KEY (user_id, segment_id),
	-- KEY (outdated_at),
	SHARD KEY (user_id, segment_id)
  ) COLLATE utf8mb4_general_ci;`

var UserSegmentSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS user_segment (
	user_id VARCHAR(60) NOT NULL,
	segment_id VARCHAR(60) NOT NULL,
	-- segment_version INT NOT NULL,
	-- enter_at DATETIME(6) NOT NULL,
	-- enter_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(enter_at, '%Y-%m-%d %H:00:00')) STORED,
	-- exit_at DATETIME,
	-- outdated_at DATETIME,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	-- db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

	PRIMARY KEY (user_id, segment_id)
	-- SORT KEY (enter_at),
	-- KEY (outdated_at)
	-- SHARD KEY (user_id, segment_id)
  ) COLLATE utf8mb4_general_ci;`
