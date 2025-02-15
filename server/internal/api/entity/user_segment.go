package entity

import (
	"time"

	"github.com/rotisserie/eris"
)

var (
	ErrUserSegmentAlreadyExists = eris.New("user segment already exists")
)

type UserSegment struct {
	UserID      string    `db:"user_id" json:"user_id"`
	SegmentID   string    `db:"segment_id" json:"segment_id"`
	DBCreatedAt time.Time `db:"db_created_at" json:"db_created_at"`
}

func NewUserSegment(userID string, segmentID string) *UserSegment {
	return &UserSegment{
		UserID:      userID,
		SegmentID:   segmentID,
		DBCreatedAt: time.Now(),
	}
}

func NewUserSegmentCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "User segments",
		Description: "User segments",
		SQL:         "SELECT * FROM `user_segment`",
		Joins:       map[string]CubeJSSchemaJoin{},
		Segments:    map[string]CubeJSSchemaSegment{},
		Measures:    map[string]CubeJSSchemaMeasure{},
		Dimensions: map[string]CubeJSSchemaDimension{
			"user_id": {
				SQL:         "user_id",
				Type:        "string",
				Title:       "User ID",
				Description: "field: user_id",
			},
			"segment_id": {
				SQL:         "segment_id",
				Type:        "string",
				Title:       "Segment ID",
				Description: "field: segment_id",
			},
		},
	}
}

var UserSegmentSchema string = `CREATE TABLE IF NOT EXISTS user_segment (
	user_id VARCHAR(64) NOT NULL,
	segment_id VARCHAR(64) NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

	SORT KEY (db_created_at),
	PRIMARY KEY (user_id, segment_id),
	SHARD KEY (user_id, segment_id)
  ) COLLATE utf8mb4_general_ci;`

var UserSegmentSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS user_segment (
	user_id VARCHAR(60) NOT NULL,
	segment_id VARCHAR(60) NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

	PRIMARY KEY (user_id, segment_id)
  ) COLLATE utf8mb4_general_ci;`
