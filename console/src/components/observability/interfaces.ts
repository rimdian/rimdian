import { BinaryOperator } from '@cubejs-client/core'
import { TableColumnDataType } from 'interfaces'

// export type TableColumnOperator =
//   | '='
//   | '!='
//   | '>'
//   | '<'
//   | 'IS NULL'
//   | 'IS NOT NULL'
//   | 'BETWEEN'
//   | 'IN'
//   | 'IS TRUE'
//   | 'IS FALSE'
// export type TableColumnOperatorString = '=' | '!=' | 'IS NULL' | 'IS NOT NULL' | 'IN'
// export type TableColumnOperatorNumber =
//   | '='
//   | '!='
//   | '>'
//   | '<'
//   | 'IS NULL'
//   | 'IS NOT NULL'
//   | 'BETWEEN'
//   | 'IN'
// export type TableColumnOperatorDate =
//   | '='
//   | '!='
//   | '>'
//   | '<'
//   | 'IS NULL'
//   | 'IS NOT NULL'
//   | 'BETWEEN'
//   | 'IN'

export interface ObservabilityCheck {
  id: string
  name: string
  measure: string
  time_dimension: string
  filters: ObservabilityCheckFilter[]
  rolling_window_value: number
  rolling_window_unit: string
  rolling_window_function: string
  condition_type: string
  threshold_position?: string
  threshold_value?: number
  is_active: boolean
  next_run_at?: Date
  emails: string[]
  db_created_at: Date
  db_updated_at: Date
}

export interface ObservabilityCheckFilter {
  dimension: string
  type: 'string' | 'number' | 'date' | 'boolean' // cubejs type
  operator: BinaryOperator
  // values are passed as SQL parameters
  values?: any[]
}

// when the check is triggered, an incident is created
export interface ObservabilityIncident {
  id: string
  check_id: string
  value: number
  comments?: string
  first_triggered_at: Date
  last_triggered_at: Date
  is_closed: boolean
  db_created_at: Date
  db_updated_at: Date
}

export interface ObservabilityColumn {
  name: string
  data_type: TableColumnDataType
  aggregation_functions: string[]
}

export interface ObservabilityTable {
  name: string
  columns: ObservabilityColumn[]
  timestamp_column: string
}
