import { FilesSettings } from 'components/assets/files/interfaces'

export interface MapOfStrings {
  [key: string]: string
}

export interface MapOfInterfaces {
  [key: string]: any
}

export interface NullableString {
  exists: boolean
  is_null: boolean
  string: string
}

export interface NullableInt64 {
  exists: boolean
  is_null: boolean
  int64: number
}

export interface NullableFloat64 {
  exists: boolean
  is_null: boolean
  float64: number
}

export interface NullableBool {
  exists: boolean
  is_null: boolean
  bool: boolean
}

export interface NullableJSON {
  exists: boolean
  is_null: boolean
  json: any
}

export interface NullableTime {
  exists: boolean
  is_null: boolean
  time: Date
}

// map of fieldName index -> field updated date
export interface MergeableFields {
  [key: string]: Date
}

export interface Account {
  id: string
  full_name?: string
  timezone: string
  locale: string
  email: string
  is_service_account: boolean
  deactivated_at?: Date
  created_at: Date
  updated_at: Date

  // joined fields
  is_owner: boolean
  workspaces_scopes: WorkspaceScope[]
}

export interface AccountLoginResult {
  account: Account
  access_token: string
  access_token_expires_at: Date
  refresh_token: string
  refresh_token_expires_at: Date
}

export interface AccountRefreshAccessTokenResult {
  access_token: string
  access_token_expires_at: Date
}

export interface Organization {
  id: string
  name: string
  currency: string
  dpo_id: string
  created_at: Date
  updated_at: Date
  deleted_at?: Date
  im_owner: boolean
}

export interface WorkspaceScope {
  workspace_id: string // * for all
  scopes: string[]
}

export interface OrganizationAccount {
  account_id: string
  organization_id: string
  is_owner: boolean
  is_service_account: boolean
  from_account_id: string
  workspaces_scopes: WorkspaceScope[]
  created_at: Date
  updated_at: Date
  deleted_at?: Date
}

export interface OrganizationInvitation {
  email: string
  organization_id: string
  from_account_id: string
  workspaces_scopes: WorkspaceScope[]
  expires_at: Date
  consumed_at?: Date
  created_at: Date
  updated_at: Date
}

export interface SecretKey {}

export const WorkspaceUserIdSigningNone = 'none'
export const WorkspaceUserIdSigningAuthenticated = 'authenticated'
export const WorkspaceUserIdSigningAll = 'all'

export const WorkspaceKindReal = 'real'
export const WorkspaceKindDemoOrder = 'order'
export const WorkspaceKindDemoLead = 'lead'

export type OnMultipleExec = 'allow' | 'discard_new' | 'retry_later' | 'abort_existing'

export interface DomainHost {
  host: string
  path_prefix: string
}

export const DomainKind = {
  Web: 'web',
  App: 'app',
  Marketplace: 'marketplace',
  Retail: 'retail',
  Telephone: 'telephone'
}

export interface Domain {
  id: string
  type: 'web' | 'app' | 'marketplace' | 'retail' | 'telephone'
  name: string
  hosts: DomainHost[]
  params_whitelist: string[]
  // brand_keywords_as_direct: boolean
  // brand_keywords: string[]
  homepages_paths: string[]
  created_at: Date
  updated_at: Date
  deleted_at?: Date
}

export interface ChannelGroup {
  id: string
  name: string
  color: string
  created_at: Date
  updated_at: Date
}

export interface VoucherCode {
  code: string
  origin_id: string
  set_utm_campaign?: string
  set_utm_content?: string
  description?: string
}

export interface Origin {
  id: string // source / medium / campaign
  match_operator: 'equals'
  utm_source: string
  utm_medium: string
  utm_campaign?: string
}

export interface Channel {
  id: string
  name: string
  origins: Origin[]
  voucher_codes: VoucherCode[]
  group_id: string
  created_at: Date
  updated_at: Date
}

export const LeadStageStatus = {
  Open: 'open',
  Converted: 'converted',
  Lost: 'lost'
}

export interface LeadStage {
  id: string
  label: string
  color: string
  status: 'open' | 'converted' | 'lost'
  created_at: Date
  updated_at: Date
  deleted_at?: Date
  migrate_to_id?: string
}

export type ExtraColumnsTable =
  | 'user'
  | 'session'
  | 'pageview'
  | 'event'
  | 'order'
  | 'lead'
  | 'cart'
export type TableColumnType =
  | 'boolean'
  | 'number'
  | 'date'
  | 'datetime'
  | 'timestamp'
  | 'varchar'
  | 'longtext'
  | 'json'
export type TableColumnDataType =
  | 'varchar'
  | 'json'
  | 'tinyint'
  | 'timestamp'
  | 'double'
  | 'date'
  | 'text'
  | 'datetime'
  | 'float'
  | 'int'
  | 'smallint'
  | 'char'
  | 'decimal'
  | 'point' // mysql
  | 'geographypoint' // singlestore
  | 'longtext'
  | 'bigint'
  | 'enum'
  | 'mediumtext'
  | 'set'
  | 'binary'
  | 'varbinary'
  | 'blob'
  | 'mediumblob'
  | 'time'
  | 'longblob'

export const TableColumnStringTypes = ['varchar', 'longtext', 'text', 'enum', 'char', 'mediumtext']
export const TableColumnNumberTypes = [
  'tinyint',
  'double',
  'float',
  'int',
  'smallint',
  'decimal',
  'bigint'
]
export const TableColumnDateTypes = ['date', 'datetime', 'timestamp', 'time']

export interface TableColumn {
  name: string
  type: TableColumnType
  size?: number
  is_required: boolean
  description?: string
  default_boolean?: boolean
  default_number?: number
  default_date?: string
  default_datetime?: string
  default_timestamp?: string
  default_string?: string
  default_json?: object
  extra_definition?: string
  hide_in_analytics?: boolean
  // created_at: Date
  // updated_at: Date
  // deleted_at?: Date
}

export interface TableJoin {
  external_table: string
  external_column: string
  local_column: string
  relationship: 'one_to_one' | 'one_to_many' | 'many_to_one'
}

export interface TableIndex {
  name: string
  columns: string[]
}

export interface ExtraColumns extends TableColumn {
  table: ExtraColumnsTable
}

export interface AppTable {
  name: string
  app_id?: string
  storage_type?: 'columnstore' | 'rowstore'
  description?: string
  columns: TableColumn[]
  joins: TableJoin[]
  indexes?: TableIndex[]
  shard_key: string[]
  unique_key: string[]
  sort_key: string[]
  timeseries_column?: string
  created_at?: Date
  updated_at?: Date
  deleted_at?: Date
}

export interface TaskExec {
  id: string
  task_id: string
  name: string
  on_multiple_exec: OnMultipleExec
  multiple_exec_key?: string
  state: any
  status: number
  message?: string
  db_created_at: Date
  db_updated_at: Date
}

export interface TaskExecList {
  task_execs: TaskExec[]
}

// holds processing state of a data import item
export interface ItemState {
  is_done: boolean
  code?: number // 200: ok, 400: external error, 500: internal error
  error?: string
}

// map of item index -> item state
export interface ItemsState {
  [key: number]: ItemState
}

export interface TaskExecJob {
  id: string
  task_exec_id: string
  done_at?: string
  db_created_at: string
  db_updated_at: string
}

// export interface FieldUpdate {
//   field: string;
//   previous: any;
//   new: any;
// }

export type ItemKind =
  | 'user'
  | 'user_alias'
  | 'pageview'
  | 'device'
  | 'segment'
  | 'session'
  | 'cart'
  | 'cart_item'
  | 'order'
  | 'order_item'
  | 'custom_event'
  | 'app_observability_check'
  | 'app_observability_incident'

export interface DataLogItem {
  kind: ItemKind
  operation: 'upsert'
  context?: DataLogContext

  // each kind has its dedicated object
  // user?: DataLogItemUser
  // user_alias?: DataLogItemUserAlias
  // app_item?: MapOfInterfaces
}

export interface DataLogItemUser extends DataLogItem {
  user: {
    external_id: string
    is_authenticated?: boolean
    created_at: Date
    updated_at?: Date // for updates
    signed_up_at?: Date
    hmac?: string
    timezone?: string
    language?: string
    country?: string
    consent_all?: NullableBool
    consent_personalization?: NullableBool
    consent_marketing?: NullableBool
    last_ip: NullableString
    longitude: NullableFloat64
    latitude: NullableFloat64
    first_name: NullableString
    last_name: NullableString
    gender: NullableString
    birthday: NullableString
    photo_url: NullableString
    email: NullableString
    email_md5: NullableString
    email_sha1: NullableString
    email_sha_256: NullableString
    telephone: NullableString
    address_line_1: NullableString
    address_line_2: NullableString
    city: NullableString
    region: NullableString
    postal_code: NullableString
    state: NullableString
    extra_columns: MapOfInterfaces
  }
}

export interface DataLogItemUserAlias extends DataLogItem {
  user_alias: {
    from_user_external_id: string
    to_user_external_id: string
    to_user_is_authenticated?: boolean
    to_user_created_at?: Date
  }
}

export interface DataLogItemSession extends DataLogItem {
  session: {
    external_id: string
    created_at: Date
    updated_at?: Date // for updates
    timezone?: string // use default workspace timezone if not provided
    device_external_id?: string
    domain_id?: string // deduced from Host for web Origin, required for non-web Origin
    duration: NullableInt64 // computed automatically with pageviews "timeSpent", but you can force it here for server-side imports
    landing_page: NullableString
    referrer: NullableString

    // utm_ parameters
    utm_source: NullableString // utm_source
    utm_medium: NullableString // utm_medium
    utm_id: NullableString // utm_id
    utm_id_from: NullableString // gclid / fbclid / cm...
    utm_campaign: NullableString // utm_campaign
    utm_content: NullableString // utm_content
    utm_term: NullableString // utm_term

    via_utm_source: NullableString // keep previous utm_source if we overwrite it
    via_utm_medium: NullableString // keep previous utm_medium if we overwrite it
    via_utm_id: NullableString // keep previous utm_id if we overwrite it
    via_utm_id_from: NullableString // keep previous utm_id from if we overwrite it
    via_utm_campaign: NullableString // keep previous utm_campaign if we overwrite it
    via_utm_content: NullableString // keep previous utm_content if we overwrite it
    via_utm_term: NullableString // keep previous utm_term if we overwrite it

    extra_columns: MapOfInterfaces
  }
}

export type DeviceType =
  | 'desktop'
  | 'mobile'
  | 'tablet'
  | 'console'
  | 'smarttv'
  | 'wearable'
  | 'embedded'

export interface DataLogItemDevice extends DataLogItem {
  device: {
    external_id: string
    // optional fields
    created_at?: Date // use first session createdAt if missing
    updated_at?: Date // for udpates
    user_agent?: string // user agent
    device_type?: DeviceType
    browser: NullableString
    browser_version: NullableString
    browser_version_major: NullableString
    os: NullableString
    resolution: NullableString // resolution
    language: NullableString // language
    ad_blocker: NullableBool // has ad blocker
    extra_columns: MapOfInterfaces
  }
}

export interface DataLogContext {
  workspace_id: string
  received_at: Date
  headers_and_params: MapOfStrings
  ip?: string
  data_sent_at?: Date
}

export interface DataLogBatch {
  workspace_id: string
  items: DataLogItem[]
}

export interface DataLogInQueueResult {
  has_error: boolean
  error?: string
  queue_should_retry?: boolean
}

export interface DataHookState {
  done: boolean
  err: boolean // is_error
  msg?: string // message
}

export interface DataHooksState {
  [key: string]: string
}

export interface DataLog {
  id: string
  workspace_id: string
  origin: number // 0: client, 1: token, 2: internal from data_log, 3: internal from task, 4: internal from workflow
  origin_id: string
  context: DataLogContext
  item: string // JSON encoded DataLogItem
  checkpoint: number
  has_error: number
  errors: MapOfStrings
  hooks: DataHooksState
  user_id: string
  merged_from_user_external_id?: string
  kind: ItemKind
  action: string
  item_id: string
  item_external_id: string
  updated_fields: MapOfStrings[]
  event_at: string
  event_at_trunc: string
  db_created_at: string
  db_updated_at: string
}

export interface User {
  id: string
  external_id: string
  is_merged?: boolean
  merged_to?: string
  merged_at?: Date
  is_authenticated: boolean
  signed_up_at?: Date
  created_at: Date
  created_at_trunc: Date
  last_interaction_at: Date
  timezone: string
  language: string
  country: string
  db_created_at: string
  db_updated_at: string
  mergeable_fields: MergeableFields

  // optional fields:
  consent_all?: boolean
  consent_personalization?: boolean
  consent_marketing?: boolean
  last_ip?: string
  longitude?: number
  latitude?: number
  geo?: any
  first_name?: string
  last_name?: string
  gender?: string
  birthday?: string
  photo_url?: string
  email?: string
  email_md5?: string
  email_sha1?: string
  email_sha256?: string
  telephone?: string
  address_line_1?: string
  address_line_2?: string
  city?: string
  region?: string
  postal_code?: string
  state?: string
  cart: Cart
  cart_items_count: number
  cart_updated_at?: Date
  cart_abandoned: boolean
  wishList: Cart
  wishlist_items_count: number
  wishlist_updated_at?: Date
  // computed fields
  orders_count: number
  orders_ltv: number
  orders_avg_cart: number
  first_order_at?: Date
  first_order_subtotal: number
  first_order_ttc: number
  last_order_at?: Date
  avg_repeat_cart: number
  avg_repeat_order_ttc: number
  [key: string]: any // custom dimensions
}

// raw cubejs schema returned by the API
export interface CubeSchemaRaw {
  fileName: string
  content: string
}

export interface CubeSchema {
  sql: string
  title: string
  description: string
  rewriteQueries?: boolean
  shown?: boolean
  joins?: { [key: string]: CubeSchemaJoin }
  segments?: { [key: string]: CubeSchemaSegment }
  measures: { [key: string]: CubeSchemaMeasure }
  dimensions: { [key: string]: CubeSchemaDimension }
}

export interface CubeSchemaJoin {
  relationship: string
  sql: string
}

export interface CubeSchemaSegment {
  sql: string
}

export interface CubeSchemaMeasure {
  title: string
  description: string
  type: 'time' | 'string' | 'number' | 'boolean' | 'geo'
  sql: string
  drillMembers: string[]
  filters?: CubeSchemaMeasureFilter[]
  format?: string
  rollingWindow?: CubeSchemaRollingWindow
  meta?: MapOfInterfaces
  shown?: boolean
}

export interface CubeSchemaMeasureFilter {
  sql: string
}

export interface CubeSchemaRollingWindow {
  trailing?: string
  leading?: string
  offset?: string
}

export interface CubeSchemaDimension {
  title: string
  description: string
  type: string
  sql: string
  primaryKey?: boolean
  shown?: boolean
  case?: CubeSchemaCase
  subquery?: boolean
  propagateFiltersToSubQuery?: boolean
  format?: string
  meta?: MapOfInterfaces
}

export interface CubeSchemaCase {
  when: CubeSchemaCaseWhen[]
  else: CubeSchemaCaseElse
}
export interface CubeSchemaCaseWhen {
  sql: string
  label: string
}
export interface CubeSchemaCaseElse {
  label: string
}

export interface UserSegment {
  user_id: string
  segment_id: string
  enter_at: Date
  enter_at_trunc: Date
  exit_at?: Date
  outdated_at?: Date
  db_created_at: string
  db_updated_at: string
}

export interface UserAlias {
  from_user_external_id: string
  to_user_external_id: string
  to_user_is_authenticated: boolean
  db_created_at: string
}

export interface Workspace {
  id: string
  name: string
  created_at: Date
  updated_at: Date
  deleted_at?: Date
  is_demo: boolean
  demo_kind: string
  website_url: string
  privacy_policy_url: string
  industry: string
  currency: string
  organization_id: string
  dpo_id: string
  default_user_timezone: string
  default_user_country: string
  default_user_language: string
  user_id_signing: string
  session_timeout: number
  abandoned_carts_processed_until?: Date
  domains: Domain[]
  channels: Channel[]
  channel_groups: ChannelGroup[]
  has_orders: boolean
  has_leads: boolean
  lead_stages: LeadStage[]
  installed_apps: AppManifest[]
  outdated_conversions_attribution: boolean
  data_hooks: DataHook[]
  license_key?: string
  files_settings: FilesSettings
  emailBlocks: any // TODO

  // joined
  cubejs_token: string
  license_info: LicenseInfo
  apps: App[]
}

export interface LicenseInfo {
  usq: number // user segments quota
  dlo90: number // data logs over 90 days
  ar: boolean // has admin role
  uslq: number // user subscription lists quota
}

export interface Session {
  id: string
  external_id: string
  user_id: string
  domain_id: string
  device_id?: string
  created_at: string
  created_at_trunc: string
  db_created_at: string
  db_updated_at: string
  merged_from_user_id?: string
  mergeable_fields: string

  timezone: string
  year: number
  month: number
  month_day: number
  week_day: number
  hour: number

  duration?: number // computed automatically with pageviews "timeSpent", but you can force it here for server-side imports
  bounced: boolean | number
  pageviews_count: number
  interactions_count: number

  landing_page?: string
  landing_page_path?: string
  referrer?: string
  referrer_domain?: string
  referrer_path?: string

  channel_origin_id: string
  channel_id: string
  channel_group_id: string

  // utm_ parameters
  utm_source?: string // utm_source
  utm_medium?: string // utm_medium
  utm_id?: string // utm_id
  utm_id_from?: string // gclid / fbclid / cm...
  utm_campaign?: string // utm_campaign
  utm_content?: string // utm_content
  utm_term?: string // utm_term

  via_utm_source?: string // keep previous utm_source if we overwrite it
  via_utm_medium?: string // keep previous utm_medium if we overwrite it
  via_utm_id?: string // keep previous utm_id if we overwrite it
  via_utm_id_from?: string // keep previous utm_id from if we overwrite it
  via_utm_campaign?: string // keep previous utm_campaign if we overwrite it
  via_utm_content?: string // keep previous utm_content if we overwrite it
  via_utm_term?: string // keep previous utm_term if we overwrite it

  conversion_type?: string
  conversion_id?: string
  conversion_external_id?: string
  conversion_at?: string
  conversion_amount?: number
  linear_amount_attributed?: number
  linear_percentage_attributed?: number
  time_to_conversion?: number
  is_first_conversion?: number
  role?: number

  [key: string]: any // custom dimensions
}

export interface Device {
  id: string
  external_id: string
  user_id: string
  created_at: string
  created_at_trunc: string
  db_created_at: string
  db_updated_at: string
  merged_from_user_id?: string
  mergeable_fields: string

  user_agent?: string
  user_agent_hash?: string
  browser?: string
  browser_version?: string
  browser_version_major?: string
  os?: string
  os_version?: string
  device_type: DeviceType
  resolution?: string
  language?: string
  languages?: string
  ad_blocker?: boolean
  in_webview?: boolean
  [key: string]: any // custom dimensions
}

export interface Order {
  id: string
  external_id: string
  user_id: string
  domain_id: string
  session_id?: string
  created_at: string
  created_at_trunc: string
  db_created_at: string
  db_updated_at: string
  merged_from_user_id?: string
  mergeable_fields: string
  discount_codes?: string // JSON array of codes
  subtotal_price: number
  total_price: number
  currency: string
  fx_rate: number
  converted_subtotal_price: number
  converted_total_price: number
  cancelled_at?: string
  cancel_reason?: string
  items: string // JSON array of I
  is_first_conversion: boolean
  time_to_conversion: number
  devices_funnel: string
  devices_type_count: number
  domains_funnel: string
  domains_type_funnel: string
  domains_count: number
  funnel: any
  funnel_hash: string
  attribution_updated_at?: string
  [key: string]: any // custom dimensions
}

export interface ProductItem {
  id: string
  external_id: string
  user_id: string
  product_external_id: string
  name: string
  sku?: string
  brand?: string
  category?: string
  variant_external_id?: string
  variant_title?: string
  image_url?: string
  price: number
  currency: string
  fx_rate: number
  converted_price: number
  quantity: number
  created_at: string
  created_at_trunc: string
  db_created_at: string
  db_updated_at: string
  is_deleted: boolean
  // merged_from_user_id?: string
  // fields_timestamp: string
}

// orderitem extends productitem and adds a field order_id
export interface OrderItem extends ProductItem {
  order_id: string
}

export interface CartItem extends ProductItem {
  cart_id: string
}

export interface Task {
  id: string
  name: string
  on_multiple_exec: OnMultipleExec
  app_id: string
  is_active: boolean
  is_cron: boolean
  minutes_interval: number // in minutes
  next_run?: string
  last_run?: string
}

export interface TaskList {
  tasks: Task[]
}

export interface App {
  id: string
  name: string
  status: 'initializing' | 'active' | 'stopped'
  state: any
  manifest: AppManifest
  is_native: boolean
  created_at: string
  updated_at: string
  deleted_at?: string
  ui_token?: string // token used to authenticate to the UI for private apps
}

export interface AppManifest {
  id: string
  name: string
  homepage: string
  author: string
  icon_url: string
  short_description: string
  description: string
  version: string
  ui_endpoint: string
  webhook_endpoint: string
  tasks?: TaskManifest[]
  app_tables?: AppTable[]
  data_hooks?: DataHookManifest[]
  extra_columns?: ExtraColumnsManifest[]
  sql_queries?: SqlQuery[]
  is_native?: boolean
}

export interface SqlQuery {
  id: string
  type: 'select'
  name: string
  description: string
  query: string
  test_args: string[]
}

export interface ExtraColumnsManifest {
  kind: string
  columns: TableColumn[]
}

export interface TaskExecJobInfoInfo {
  id: string
  // name: string
  create_time: Date
  schedule_time?: Date
  // dispatch_deadline?: string
  dispatch_count: Number
  response_count: Number
  first_attempt?: TaskExecJobInfoAttempt
  last_attempt?: TaskExecJobInfoAttempt
}

export interface TaskExecJobInfoAttempt {
  schedule_time?: Date
  dispatch_time?: Date
  response_time?: Date
  response_code?: Number
  response_message?: string
}

export interface TaskManifest {
  id: string
  name: string
  is_cron: boolean
  on_multiple_exec: 'allow' | 'discard_new' | 'retry_later' | 'abort_existing'
  minutes_interval?: number
}

export interface AppItemField {
  name: string
  type: string
  bool_value?: boolean
  float64_value?: number
  string_value?: string
  time_value?: Date
  json_value?: any
}

export interface AppItem {
  id: string
  external_id: string
  kind: string
  user_id: string // 'none' if anonymous
  created_at: Date
  created_at_trunc: Date
  db_created_at: string
  db_updated_at: string
  merged_from_user_id?: string
  fields_timestamp: MapOfInterfaces
  // specific fields
  [key: string]: any
}

export interface Pageview {
  id: string
  external_id: string
  user_id: string
  domain_id: string
  session_id: string
  created_at: string
  created_at_trunc: string
  db_created_at: string
  db_updated_at: string
  is_deleted: boolean
  merged_from_user_id?: string
  fields_timestamp: MapOfInterfaces
  // specific fields
  page_id?: string
  title?: string
  referrer?: string
  referrer_domain?: string
  referrer_path?: string
  duration?: number
  image_url?: string
  product_external_id?: string
  product_sku?: string
  product_name?: string
  product_brand?: string
  product_category?: string
  product_variant_external_id?: string
  product_variant_title?: string
  product_price?: number
  product_currency?: string
  product_fx_rate?: number
  product_converted_price?: number
}

export interface CustomEvent {
  id: string
  external_id: string
  user_id: string
  domain_id: string
  session_id: string
  created_at: string
  created_at_trunc: string
  db_created_at: string
  db_updated_at: string
  is_deleted: boolean
  merged_from_user_id?: string
  fields_timestamp: MapOfInterfaces
  // specific fields
  label: string
  string_value?: string
  number_value?: number
  boolean_value?: boolean
  non_interactive: boolean
}

export interface Cart {
  id: string
  external_id: string
  user_id: string
  domain_id: string
  session_id?: string
  created_at: string
  created_at_trunc: string
  db_created_at: string
  db_updated_at: string
  // is_deleted: boolean
  // merged_from_user_id?: string
  // fields_timestamp: MapOfInterfaces
  // specific fields
  currency: string
  fx_rate: number
  public_url?: string
  items: string
  status: number // 0: abandoned, 1: converted, 2: recovered
}

export interface DataHookManifest {
  id: string
  name: string
  on: 'on_validation' | 'on_success'
  for: DataHookFor[]
  js?: string
}

export interface DataHookFor {
  kind: string
  action: string
}

export interface DataHook {
  id: string
  app_id: string
  name: string
  on: 'on_validation' | 'on_success'
  for: DataHookFor[]
  enabled: boolean
  db_created_at: string
  db_updated_at: string
}

export interface SubscriptionList {
  id: string
  name: string
  color: string
  channel: 'email' // only email for now
  double_opt_in: boolean
  email_template_id?: string
  email_template_version?: number
  db_created_at: string
  db_updated_at: string
  users_count?: number
}
