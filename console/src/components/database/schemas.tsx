export interface TableInformationSchema {
  name: string
  type: string // BASE TABLE | VIEW
  created_at: string
  storage_type: string // COLUMNSTORE | INMEMORY_ROWSTORE
  rows: number
  memory_use: number
  columns: ColumnInformationSchema[]
}

export interface ColumnInformationSchema {
  name: string
  position: number
  default_value?: string
  nullable: string
  column_type: string
  data_type: string
  character_max_length?: number
  numeric_precision?: number
  character_set?: string
  collation?: string
  column_key: string // "" | "PRI" | "MUL"
  extra: string // "" | ON UPDATE CURRENT_TIMESTAMP
}

export const TablesWithCustomColumns = [
  'users'
  // TODO: add table that allow custom dimensions
]

export const TablesDescriptions = {
  users: {
    description: '...',
    columns: {
      id: 'id used internally',
      external_id: 'id used by an external data source',
      is_merged: 'is this user merged into another?',
      merged_to: 'user id of merged user',
      is_authenticated: 'is this user id coming from your primary data source?'
      // TODO
    }
  }
}

// export const Tables: DBTableSchema[] = [
//     {
//         name: 'users',
//         description: <>Users table</>,
//         columns: [
//             {
//                 name: 'id',
//                 description: <>user internal ID</>,
//                 type: 'VARCHAR(60)',
//                 nullable: false,
//             }
//         ],
//     }
// ]

// id VARCHAR(60) NOT NULL,
// external_id VARCHAR(255) NOT NULL,
// is_merged BOOLEAN DEFAULT FALSE,
// merged_to VARCHAR(60),
// merged_at DATETIME,
// is_authenticated BOOLEAN DEFAULT FALSE,
// created_at DATETIME NOT NULL,
// updated_at DATETIME NOT NULL,
// first_seen_at DATETIME NOT NULL,
// last_interaction_at DATETIME NOT NULL,
// user_centric_consent BOOLEAN DEFAULT FALSE,
// timezone VARCHAR(50) NOT NULL,
// language CHAR(2) NOT NULL,
// country VARCHAR(50) NOT NULL,
// db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
// db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

// last_ip VARCHAR(45),
// latitude DECIMAL(9,6),
// longitude DECIMAL(9,6),
// geo AS GEOGRAPHY_POINT(longitude, latitude) PERSISTED geographypoint,
// signed_up_at DATETIME,
// first_name NVARCHAR(50),
// last_name NVARCHAR(50),
// gender CHAR(1),
// birthday DATE,
// photo_url VARCHAR(2083),
// email VARCHAR(255),
// email_md5 AS MD5(email) PERSISTED NVARCHAR(512),
// email_sha1 AS SHA1(email) PERSISTED NVARCHAR(512),
// email_sha256 AS SHA2(email, 256) PERSISTED NVARCHAR(512),
// telephone VARCHAR(22),
// remarketing_consent BOOLEAN DEFAULT FALSE,
// address_line_1 NVARCHAR(255),
// address_line_2 NVARCHAR(255),
// city NVARCHAR(50),
// region NVARCHAR(50),
// postal_code VARCHAR(50),
// state NVARCHAR(50),
// cart JSON NOT NULL,
// cart_items_count AS JSON_LENGTH(cart::items) PERSISTED INT UNSIGNED,
// cart_updated_at AS TO_DATE(cart::$updatedAt, 'YYYY-MM-DDTHH24:MI:SS') PERSISTED DATETIME,
// cart_abandoned BOOLEAN DEFAULT FALSE,
// wishlist JSON NOT NULL,
// wishlist_items_count AS JSON_LENGTH(wishlist::items) PERSISTED INT UNSIGNED,
// wishlist_updated_at AS TO_DATE(wishlist::$updatedAt, 'YYYY-MM-DDTHH24:MI:SS') PERSISTED DATETIME,

// KEY (created_at) USING CLUSTERED COLUMNSTORE,
// PRIMARY KEY (id),
// KEY (is_authenticated),
// KEY (external_id),
// SHARD KEY (id)
