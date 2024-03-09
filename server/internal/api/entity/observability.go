package entity

// type App_Observability_Check struct {
// 	ID            string `db:"id" json:"id"`
// 	Name          string `db:"name" json:"name"`
// 	Measure       string `db:"measure" json:"measure"`
// 	TimeDimension string `db:"time_dimension" json:"time_dimension"`
// 	// Table               string `db:"table" json:"table"`                               // table name
// 	// Column              string `db:"column" json:"column"`                             // table column
// 	// AggregationFunction string `db:"aggregation_function" json:"aggregation_function"` // aggregation function
// 	// Metric                string                      `db:"measure" json:"measure"`
// 	Filters            App_Observability_Check_Filters `db:"filters" json:"filters"`
// 	RollingWindowValue int64                           `db:"rolling_window_value" json:"rolling_window_value"`
// 	RollingWindowUnit  string                          `db:"rolling_window_unit" json:"rolling_window_unit"`
// 	// RollingWindowFunction string                      `db:"rolling_window_function" json:"rolling_window_function"`
// 	ConditionType     string     `db:"condition_type" json:"condition_type"`
// 	ThresholdPosition *string    `db:"threshold_position" json:"threshold_position,omitempty"`
// 	ThresholdValue    *float64   `db:"threshold_value" json:"threshold_value,omitempty"`
// 	IsActive          bool       `db:"is_active" json:"is_active"`
// 	LastRunAt         *time.Time `db:"last_run_at" json:"last_run_at,omitempty"`
// 	NextRunAt         *time.Time `db:"next_run_at" json:"next_run_at,omitempty"`
// 	Emails            Emails     `db:"emails" json:"emails"`
// 	// fields computed by the server:
// 	// SQL         string    `db:"sql" json:"sql"` // generated SQL
// 	DBCreatedAt time.Time `db:"db_created_at" json:"db_created_at"`
// 	DBUpdatedAt time.Time `db:"db_updated_at" json:"db_updated_at"`
// }

// compute the next run time based on the rolling window
// func (c *App_Observability_Check) ComputeNextRun() {
// 	now := time.Now()
// 	var nextRunAt time.Time

// 	switch c.RollingWindowUnit {
// 	case Observability_Check_Unit_Minute:
// 		nextRunAt = now.Add(time.Duration(c.RollingWindowValue) * time.Minute)
// 	case Observability_Check_Unit_Hour:
// 		nextRunAt = now.Add(time.Duration(c.RollingWindowValue) * time.Hour)
// 	case Observability_Check_Unit_Day:
// 		nextRunAt = now.Add(time.Duration(c.RollingWindowValue) * time.Hour * 24)
// 	}

// 	c.NextRunAt = &nextRunAt
// }

// func (c *App_Observability_Check) Validate() error {
// 	if c.ID == "" {
// 		return eris.New("ID is required")
// 	}
// 	if c.Name == "" {
// 		return eris.New("Name is required")
// 	}
// 	if c.Measure == "" {
// 		return eris.New("Metric is required")
// 	}
// 	if c.TimeDimension == "" {
// 		return eris.New("TimeDimension is required")
// 	}
// 	// if c.AggregationFunction != Observability_Check_AggregationFunction_Count &&
// 	// 	c.AggregationFunction != Observability_Check_AggregationFunction_Sum &&
// 	// 	c.AggregationFunction != Observability_Check_AggregationFunction_Average &&
// 	// 	c.AggregationFunction != Observability_Check_AggregationFunction_Min &&
// 	// 	c.AggregationFunction != Observability_Check_AggregationFunction_Max {
// 	// 	return eris.New("AggregationFunction is invalid")
// 	// }
// 	if c.RollingWindowValue <= 0 {
// 		return eris.New("Rolling window value must be greater than 0")
// 	}
// 	// check unit is valid
// 	if c.RollingWindowUnit == "" {
// 		return eris.New("Rolling window unit is required")
// 	}
// 	if c.RollingWindowUnit != Observability_Check_Unit_Minute &&
// 		c.RollingWindowUnit != Observability_Check_Unit_Hour &&
// 		c.RollingWindowUnit != Observability_Check_Unit_Day {
// 		return eris.New("Rolling window unit is invalid")
// 	}
// 	if c.ConditionType == "" {
// 		return eris.New("Condition type is required")
// 	}
// 	if c.ConditionType == Observability_Check_ConditionType_Threshold {
// 		if c.ThresholdPosition == nil {
// 			return eris.New("Threshold position is required")
// 		}
// 		if c.ThresholdValue == nil {
// 			return eris.New("Threshold value is required")
// 		}
// 	}
// 	if c.Emails == nil {
// 		c.Emails = []string{}

// 		// check emails are valid
// 		for _, email := range c.Emails {
// 			if !govalidator.IsEmail(email) {
// 				return eris.Errorf("Email %v is invalid", email)
// 			}
// 		}
// 	}

// 	if c.Filters == nil {
// 		c.Filters = []*App_Observability_Check_Filter{}

// 		// check filters are valid
// 		for _, filter := range c.Filters {
// 			if err := filter.Validate(c); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// // Filter = SQL WHERE clause
// type App_Observability_Check_Filters []*App_Observability_Check_Filter

// func (x *App_Observability_Check_Filters) Scan(val interface{}) error {
// 	return json.Unmarshal(val.([]byte), &x)
// }

// func (x App_Observability_Check_Filters) Value() (driver.Value, error) {
// 	return json.Marshal(x)
// }

// type ArrayOfInterfaces []interface{}

// func (x *ArrayOfInterfaces) Scan(val interface{}) error {
// 	return json.Unmarshal(val.([]byte), &x)
// }

// func (x ArrayOfInterfaces) Value() (driver.Value, error) {
// 	return json.Marshal(x)
// }

// type App_Observability_Check_Filter struct {
// 	Dimension string            `json:"dimension"`
// 	Type      string            `json:"type"` // cubejs type: string | number | time | boolean
// 	Operator  string            `json:"operator"`
// 	Values    ArrayOfInterfaces `json:"values,omitempty"`
// }

// func (x *App_Observability_Check_Filter) Scan(val interface{}) error {
// 	return json.Unmarshal(val.([]byte), &x)
// }

// func (x App_Observability_Check_Filter) Value() (driver.Value, error) {
// 	return json.Marshal(x)
// }

// // validate filter
// func (f *App_Observability_Check_Filter) Validate(check *App_Observability_Check) error {
// 	if f.Dimension == "" {
// 		return eris.New("Dimension is required")
// 	}
// 	if f.Type == "" {
// 		return eris.New("Type is required")
// 	}
// 	if f.Operator == "" {
// 		return eris.New("Operator is required")
// 	}
// 	return nil
// }

// when the check is triggered, an incident is created
// type App_Observability_Incident struct {
// 	ID          string  `db:"id" json:"id"`
// 	CheckID     string  `db:"check_id" json:"check_id"`
// 	Value       float64 `db:"value" json:"value"`
// 	Comments    *string `db:"comments" json:"comments,omitempty"`
// 	FirstSeenAt string  `db:"first_seen_at" json:"first_seen_at"`
// 	LastSeenAt  string  `db:"last_seen_at" json:"last_seen_at"`
// 	IsClosed    bool    `db:"is_closed" json:"is_closed"`
// 	DBCreatedAt string  `db:"db_created_at" json:"db_created_at"`
// 	DBUpdatedAt string  `db:"db_updated_at" json:"db_updated_at"`
// }

// var Observability_CheckSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS observability_check (
// 	id VARCHAR(64) NOT NULL,
// 	name VARCHAR(255) NOT NULL,
// 	measure VARCHAR(64) NOT NULL,
// 	time_dimension VARCHAR(64) NOT NULL,
// 	filters JSON NOT NULL,
// 	rolling_window_value INT NOT NULL,
// 	rolling_window_unit VARCHAR(64) NOT NULL,
// 	-- rolling_window_function VARCHAR(64) NOT NULL,
// 	condition_type VARCHAR(64) NOT NULL,
// 	threshold_position VARCHAR(64),
// 	threshold_value FLOAT,
// 	is_active BOOLEAN NOT NULL DEFAULT FALSE,
// 	last_run_at DATETIME,
// 	next_run_at DATETIME,
// 	emails JSON NOT NULL,
// 	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
// 	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

// 	PRIMARY KEY (id),
// 	SHARD KEY (id)
// ) COLLATE utf8mb4_general_ci;`

// var Observability_CheckSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS observability_check (
// 	id VARCHAR(64) NOT NULL,
// 	name VARCHAR(255) NOT NULL,
// 	measure VARCHAR(64) NOT NULL,
// 	time_dimension VARCHAR(64) NOT NULL,
// 	filters JSON NOT NULL,
// 	rolling_window_value INT NOT NULL,
// 	rolling_window_unit VARCHAR(64) NOT NULL,
// 	-- rolling_window_function VARCHAR(64) NOT NULL,
// 	condition_type VARCHAR(64) NOT NULL,
// 	threshold_position VARCHAR(64),
// 	threshold_value FLOAT,
// 	is_active BOOLEAN NOT NULL DEFAULT FALSE,
// 	last_run_at DATETIME,
// 	next_run_at DATETIME,
// 	emails JSON NOT NULL,
// 	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
// 	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

// 	PRIMARY KEY (id)
// ) COLLATE utf8mb4_general_ci;`

// // INCIDENT

// var Observability_IncidentSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS observability_incident (
// 	id VARCHAR(64) NOT NULL,
// 	check_id VARCHAR(64) NOT NULL,
// 	value FLOAT NOT NULL,
// 	comments VARCHAR(512),
// 	first_seen_at DATETIME NOT NULL,
// 	last_seen_at DATETIME NOT NULL,
// 	is_closed BOOLEAN NOT NULL DEFAULT FALSE,
// 	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
// 	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

// 	PRIMARY KEY (id),
// 	SHARD KEY (check_id),
// 	KEY (db_created_at DESC)
// ) COLLATE utf8mb4_general_ci;`

// var Observability_IncidentSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS observability_incident (
// 	id VARCHAR(64) NOT NULL,
// 	check_id VARCHAR(64) NOT NULL,
// 	value FLOAT NOT NULL,
// 	comments VARCHAR(512),
// 	first_seen_at DATETIME NOT NULL,
// 	last_seen_at DATETIME NOT NULL,
// 	is_closed BOOLEAN NOT NULL DEFAULT FALSE,
// 	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
// 	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

// 	PRIMARY KEY (id),
// 	KEY (db_created_at DESC)
// ) COLLATE utf8mb4_general_ci;`
