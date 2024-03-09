package entity

import "time"

// UserSegmentQueue is a row in the user_segment_queue table
// it is used to queue the users that need to be recomputed for a segment

type UserSegmentQueue struct {
	UserID         string    `db:"user_id" json:"user_id"`
	SegmentID      string    `db:"segment_id" json:"segment_id"`
	SegmentVersion int       `db:"segment_version" json:"segment_version"`
	Enters         int       `db:"enters" json:"enters"`       // 1 = enter, 0 = exit
	WorkerID       int       `db:"worker_id" json:"worker_id"` // random int between 1-10 for worker atomicity
	DBCreatedAt    time.Time `db:"db_created_at" json:"db_created_at"`
}

var UserSegmentQueueSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS user_segment_queue (
	user_id VARCHAR(64) NOT NULL,
	segment_id VARCHAR(64) NOT NULL,
	segment_version INT NOT NULL,
	enters INT NOT NULL,
	-- specify a default = 1 that is useless... as singlestore bugs with Error 1364 (HY000): Field 'worker_id' doesn't have a default value
	worker_id INT NOT NULL DEFAULT 1,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (segment_id, segment_version, user_id),
	KEY(worker_id)
) COLLATE utf8mb4_general_ci;`

var UserSegmentQueueSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS user_segment_queue (
	user_id VARCHAR(64) NOT NULL,
	segment_id VARCHAR(64) NOT NULL,
	segment_version INT NOT NULL,
	enters INT NOT NULL,
	worker_id INT NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (segment_id, segment_version, user_id),
	KEY(worker_id)
) COLLATE utf8mb4_general_ci;`
