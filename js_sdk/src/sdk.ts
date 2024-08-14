// import Ajv from 'ajv'
import md5 from 'blueimp-md5'
import lifecycle from 'page-lifecycle/dist/lifecycle.mjs'

// const ajv = new Ajv()

// params used to decorate URLs for cross-device / cross-domain linking
const URLParams = {
  device_external_id: '_did',
  user_authenticated_external_id: '_authuid',
  user_external_id: '_uid',
  user_is_authenticated: '_auth',
  user_external_id_hmac: '_uidh'
}

const OneYearInSeconds = 31536000

const PageStates = {
  active: 'active',
  passive: 'passive',
  hidden: 'hidden',
  frozen: 'frozen'
}

type IRimdian = {
  config: IConfig

  isReady: boolean

  // current context
  dispatchConsent: boolean
  currentUser?: IUser
  currentDevice?: IDevice
  currentSession?: ISession
  currentPageview?: IPageview
  currentPageviewVisibleSince?: Date
  currentPageviewDuration: number // secs of time spent on the page
  currentCart?: ICart

  itemsQueue: IItemsQueue // current items in the queue
  dispatchQueue: IDataImport[] // batches ready to be dispatched

  isDispatching: boolean // true when dispatching

  // public methods
  log: (level: string, ...args) => void
  init: (cfg: IConfig) => void
  trackPageview: (data: IPageview) => void
  trackCustomEvent: (data: ICustomEvent) => void
  trackCart: (data: ICart) => void
  trackOrder: (data: IOrder) => void
  getCurrentUser: (callback: (user: IUser) => void) => void
  setDeviceContext: (data: IDevice) => void
  setSessionContext: (data: ISession) => void
  setUserContext: (data: IUser) => void
  saveUserProfile: () => void
  dispatch: (useBeacon: boolean) => void
  setDispatchConsent: (consent: boolean) => void
  isBrowserLegit: () => boolean
  uuidv4: () => string
  md5: (str: string) => string
  getReferrer: () => string | undefined
  getTimezone: () => string | undefined
  getQueryParam: (url: string, name: string) => string | undefined
  getHashParam: (hash: string, name: string) => string | undefined
  updateURLParam: (url: string, name: string, value: string) => string
  hasAdBlocker: () => boolean
  isPageVisible: () => boolean
  onReady: (fn: () => void) => void
  onReadyQueue: Array<Function>
  getCookie: (name: string) => string
  setCookie: (name: string, value: string, seconds: number) => void
  deleteCookie: (name: string) => void

  // private methods
  _onReady: (cfg: IConfig) => void
  _execWhenReady: (fn: () => void) => void
  _initDispatchLoop: (useBeacon: boolean) => void
  _postPayload: (dataImport: IDataImport, retryCount: number, useBeacon: boolean) => void
  _post: (data: IDataImport, useBeacon: boolean, callback: (error: string) => void) => void
  _handleUser: () => void
  _createUser: (userExternalId: string, isAuthenticated: boolean, createdAt: string) => void
  _enrichUserContext: () => void
  _handleDevice: () => void
  _createDevice: () => void
  _handleSession: () => void
  _startNewSession: (params: ISessionDTO) => void
  _onPagePassive: () => void
  _onPageActive: () => void
  _cartHash: (data: ICart) => string
  _localStorage: ILocalStorage
  _addEventListener: (
    element: any,
    eventType: string,
    eventHandler: Function,
    useCapture: boolean
  ) => void
  _normalizeUTMSource: (source: string) => string
  _decorateURL: (e: MouseEvent) => void
  _wipeAll: () => void
}

type IConfig = {
  workspace_id: string
  host: string // hostname of the collector server
  session_timeout: number // in minutes
  namespace: string // prefix for cookies and localStorage
  cross_domains: string[] // list of domains that will get URL decoration (userId, sessionId, etc.)
  ignored_origins: IOrigin[] // list of origins that won't trigger a new session (used to avoid bank redirections to ure the session)
  version: string
  log_level: 'error' | 'warn' | 'info' | 'debug' | 'trace'
  max_retry: number // max number of retries when posting data to the collector
  from_cm: boolean
}

type IOrigin = {
  utm_source: string
  utm_medium: string
  utm_campaign?: string
}

type ISessionDTO = {
  utm_source?: string
  utm_medium?: string
  utm_campaign?: string
  utm_content?: string
  utm_term?: string
  utm_id?: string
  utm_id_from?: string
}

// const OriginSchema = {
//   type: 'object',
//   properties: {
//     utm_source: { type: 'string' },
//     utm_medium: { type: 'string' },
//     utm_campaign: { type: 'string' }
//   },
//   required: ['utm_source', 'utm_medium']
// }

// const ValidateConfig = ajv.compile({
//   type: 'object',
//   properties: {
//     workspace_id: { type: 'string', required: true },
//     host: { type: 'string', required: true },
//     sessionTimeout: { type: 'number', required: true },
//     namespace: { type: 'string', required: true },
//     dispatchConsent: { type: 'boolean', required: true },
//     crossDomains: { type: 'array', items: { type: 'string' }, required: true },
//     ignoredOrigins: {
//       type: 'array',
//       items: {
//         type: 'object',
//         properties: {
//           utm_source: { type: 'string', required: true },
//           utm_medium: { type: 'string', required: true },
//           utm_campaign: { type: 'string', required: false }
//         }
//       },
//       required: true
//     },
//     version: { type: 'string', required: true },
//     logLevel: { type: 'string', required: true, enum: ['error', 'warn', 'info', 'debug', 'trace'] },
//     maxRetry: { type: 'number', required: true }
//   },
//   required: true
// })

type ILocalStorage = {
  get: (key: string) => string | null
  set: (key: string, value: string) => void
  remove: (key: string) => void
}

// keep track of current user
type IUser = {
  external_id: string
  is_authenticated: boolean
  created_at: string
  updated_at?: string
  last_interaction_at?: string
  user_centric_consent?: boolean // deprecated, replaced by consent_personalization
  consent_all?: boolean
  consent_personalization?: boolean
  consent_marketing?: boolean
  signed_up_at?: string
  hmac?: string
  [key: string]: any // custom dimensions
}

// const ValidateUser = ajv.compile({
//   type: 'object',
//   properties: {
//     external_id: { type: 'string', required: true },
//     is_authenticated: { type: 'boolean', required: true },
//     created_at: { type: 'string', required: true },
//     updated_at: { type: 'string', required: false },
//     user_centric_consent: { type: 'boolean', required: false },
//     signed_up_at: { type: 'string', required: false },
//     hmac: { type: 'string', required: false }
//   },
//   required: true,
//   additionalProperties: true
// })

type IDevice = {
  external_id: string
  created_at: string
  updated_at?: string
  user_agent?: string
  resolution?: string
  language?: string
  ad_blocker?: boolean
  [key: string]: any // custom dimensions
}

// const ValidateDevice = ajv.compile({
//   type: 'object',
//   properties: {
//     external_id: { type: 'string', required: true },
//     created_at: { type: 'string', required: true },
//     updated_at: { type: 'string', required: false },
//     user_agent: { type: 'string', required: false },
//     resolution: { type: 'string', required: false },
//     language: { type: 'string', required: false },
//     ad_blocker: { type: 'boolean', required: false }
//   },
//   required: true,
//   additionalProperties: true
// })

type ISession = {
  external_id: string
  created_at: string
  landing_page: string
  device_external_id?: string
  referrer?: string
  timezone?: string

  utm_source?: string
  utm_medium?: string
  utm_campaign?: string
  utm_content?: string
  utm_term?: string
  utm_id?: string
  utm_id_from?: string

  duration?: number
  pageviews_count?: number
  interactions_count?: number

  [key: string]: any // custom dimensions
}

// const ValidateSession = ajv.compile({
//   type: 'object',
//   properties: {
//     external_id: { type: 'string', required: true },
//     created_at: { type: 'string', required: true },
//     landing_page: { type: 'string', required: true },
//     referrer: { type: 'string', required: false },
//     timezone: { type: 'string', required: false },
//     utm_source: { type: 'string', required: false },
//     utm_medium: { type: 'string', required: false },
//     utm_campaign: { type: 'string', required: false },
//     utm_content: { type: 'string', required: false },
//     utm_term: { type: 'string', required: false },
//     utm_id: { type: 'string', required: false },
//     utm_id_from: { type: 'string', required: false }
//   },
//   required: true,
//   additionalProperties: true
// })

type IPageview = {
  external_id: string
  page_id: string // page URL by default
  title: string // page title by default
  created_at: string
  updated_at?: string
  referrer?: string
  duration?: number
  image_url?: string
  // ecommerce product details
  product_external_id?: string
  product_sku?: string
  product_name?: string
  product_brand?: string
  product_category?: string
  product_variant_external_id?: string
  product_variant_title?: string
  product_price?: number
  product_currency?: string
  [key: string]: any // custom dimensions
}

// const ValidatePageview = ajv.compile({
//   type: 'object',
//   properties: {
//     external_id: { type: 'string', required: true },
//     page_id: { type: 'string', required: true },
//     title: { type: 'string', required: true },
//     created_at: { type: 'string', required: true },
//     updated_at: { type: 'string', required: false },
//     referrer: { type: 'string', required: false },
//     duration: { type: 'number', required: false },
//     image_url: { type: 'string', required: false },
//     product_external_id: { type: 'string', required: false },
//     product_sku: { type: 'string', required: false },
//     product_name: { type: 'string', required: false },
//     product_brand: { type: 'string', required: false },
//     product_category: { type: 'string', required: false },
//     product_variant_external_id: { type: 'string', required: false },
//     product_variant_title: { type: 'string', required: false },
//     product_price: { type: 'number', required: false },
//     product_currency: { type: 'string', required: false }
//   },
//   required: true,
//   additionalProperties: true
// })

type ICart = {
  external_id: string
  created_at: string
  updated_at: string
  items: ICartItem[]
  currency?: string
  public_url?: string
  hash?: string // hash of the cart, used to avoid sending duplicate carts
  [key: string]: any // custom dimensions
}

// const CartItemSchema = {
//   type: 'object',
//   properties: {
//     external_id: { type: 'string', required: true },
//     name: { type: 'string', required: true },
//     sku: { type: 'string', required: true },
//     price: { type: 'number', required: true },
//     quantity: { type: 'number', required: true },
//     brand: { type: 'string', required: false },
//     category: { type: 'string', required: false },
//     variant_external_id: { type: 'string', required: false },
//     variant_title: { type: 'string', required: false },
//     image_url: { type: 'string', required: false },
//     currency: { type: 'string', required: false },
//     discount_codes: { type: 'array', required: false }
//   },
//   required: true,
//   additionalProperties: true
// }

// const ValidateCartItem = ajv.compile(CartItemSchema)

// const ValidateCart = ajv.compile({
//   type: 'object',
//   properties: {
//     external_id: { type: 'string', required: true },
//     created_at: { type: 'string', required: true },
//     updated_at: { type: 'string', required: true },
//     items: { type: 'array', required: true },
//     currency: { type: 'string', required: false },
//     public_url: { type: 'string', required: false },
//     hash: { type: 'string', required: false }
//   },
//   required: true,
//   additionalProperties: true
// })

type ICartItem = {
  cart_external_id: string
  external_id: string
  product_external_id: string
  name: string
  sku: string
  price?: number
  quantity: number
  brand?: string
  category?: string
  variant_external_id?: string
  variant_title?: string
  image_url?: string
  currency?: string
  discount_codes?: string[]
  [key: string]: any // custom dimensions
}

type IOrderItem = {
  order_external_id: string
  external_id: string
  name: string
  product_external_id: string
  price?: number
  quantity?: number
  sku?: string
  brand?: string
  category?: string
  variant_external_id?: string
  variant_title?: string
  image_url?: string
}

type IOrder = {
  external_id: string
  // session_external_id: string
  created_at: string
  updated_at?: string
  currency?: string
  items: IOrderItem[]
  discount_codes?: string[]
  subtotal_price?: number
  total_price?: number
  cancelled_at?: string
  cancel_reason?: string
  [key: string]: any // custom dimensions
}

// const ValidateOrder = ajv.compile({
//   type: 'object',
//   properties: {
//     external_id: { type: 'string', required: true },
//     // session_external_id: { type: 'string', required: true },
//     created_at: { type: 'string', required: true },
//     updated_at: { type: 'string', required: false },
//     currency: { type: 'string', required: false },
//     items: { type: 'array', required: true, items: CartItemSchema },
//     discount_codes: { type: 'array', required: false },
//     subtotal_price: { type: 'number', required: false },
//     total_price: { type: 'number', required: false },
//     cancelled_at: { type: 'string', required: false },
//     cancel_reason: { type: 'string', required: false }
//   },
//   required: true,
//   additionalProperties: true
// })

type ICustomEvent = {
  external_id: string
  label: string
  created_at: string
  session_external_id: string
  updated_at?: string
  value?: number
  non_interactive?: boolean
  [key: string]: any // custom dimensions
}

type ItemData = IUser | IUserAlias | IDevice | IPageview | ICart | IOrder | ICustomEvent
type IItem = {
  kind: ItemKind
  context?: IContext
  // items
  user?: IUser
  user_alias?: IUserAlias
  device?: IDevice
  session?: ISession
  pageview?: IPageview
  cart?: ICart
  order?: IOrder
  customEvent?: ICustomEvent
}

type ItemKind =
  | 'user'
  | 'user_alias'
  | 'device'
  | 'session'
  | 'pageview'
  | 'cart'
  | 'order'
  | 'custom_event'

type IItemsQueue = {
  items: IItem[]
  add: (kind: ItemKind, data: ItemData) => void
  addPageviewDuration: () => void
}

type IContext = {
  // used at dataImport level
  data_sent_at?: string
}

type IUserAlias = {
  from_user_external_id: string
  to_user_external_id: string
  to_user_is_authenticated?: boolean
  to_user_created_at?: string
}

type IDataImport = {
  id: string
  workspace_id: string
  items: IItem[]
  context: IContext
  created_at: string
}

const Rimdian: IRimdian = {
  config: {
    workspace_id: '',
    host: 'https://collector-eu.rimdian.com',
    session_timeout: 60 * 30, // 30 minutes
    namespace: '_rmd_',
    cross_domains: [],
    ignored_origins: [],
    version: '2.10.0',
    log_level: 'error',
    max_retry: 10,
    from_cm: false
  },

  isReady: false,

  dispatchConsent: false,
  currentUser: undefined,
  currentDevice: undefined,
  currentSession: undefined,
  currentPageview: undefined,
  currentPageviewVisibleSince: undefined,
  currentPageviewDuration: 0,
  currentCart: undefined,

  itemsQueue: {
    items: [],
    add: (kind: ItemKind, data: ItemData) => {
      // if current session expired, and received an interaction event, create a new session
      const sessionCookie = Rimdian.getCookie(Rimdian.config.namespace + 'session')
      if (
        !sessionCookie &&
        (['pageview', 'cart', 'order'].includes(kind) ||
          (kind === 'custom_event' && (data as ICustomEvent).non_interactive === true))
      ) {
        Rimdian._startNewSession({})
      }

      const item: IItem = { kind: kind }
      item[kind] = data

      // interations have a session + user + device context
      if (['pageview', 'custom_event', 'cart', 'order'].includes(kind)) {
        item.user = { ...Rimdian.currentUser }
        item.session = { ...Rimdian.currentSession }
      }

      Rimdian.itemsQueue.items.push(item)
      // persist items in local storage
      Rimdian._localStorage.set('items', JSON.stringify(Rimdian.itemsQueue.items))
    },
    // addPageviewDuration() is called when page becomes passive/not focused
    // pageview duration should not trigger a new session if the session expired
    // that's why we handle it separately
    addPageviewDuration: () => {
      // abort if we are not tracking the current pageview
      if (!Rimdian.currentPageview || !Rimdian.currentPageviewVisibleSince) {
        return
      }

      // increment the time spent
      const increment = Math.round(
        (new Date().getTime() - Rimdian.currentPageviewVisibleSince.getTime()) / 1000
      )
      Rimdian.currentPageviewDuration += increment
      Rimdian.log('info', 'time spent on page is now', Rimdian.currentPageviewDuration)

      // add pageview to items queue
      const pageview = { ...Rimdian.currentPageview }
      const now = new Date().toISOString()

      pageview.duration = Rimdian.currentPageviewDuration
      pageview.updated_at = now

      // increment session duration too
      if (!Rimdian.currentSession.duration) Rimdian.currentSession.duration = 0
      Rimdian.currentSession.duration += increment
      Rimdian.currentSession.updated_at = now
      Rimdian.setSessionContext(Rimdian.currentSession) // cookie update

      // error on invalid pageview
      // if (!ValidatePageview(pageview)) {
      //   Rimdian.log('error', 'invalid pageview', ValidatePageview.errors)
      // }

      const item = {
        kind: 'pageview' as ItemKind,
        pageview: pageview,
        user: { ...Rimdian.currentUser },
        session: { ...Rimdian.currentSession },
        device: { ...Rimdian.currentDevice }
      }
      Rimdian.itemsQueue.items.push(item)
      // persist items in local storage
      Rimdian._localStorage.set('items', JSON.stringify(Rimdian.itemsQueue.items))
    }
  },
  dispatchQueue: [],
  isDispatching: false,
  onReadyQueue: [],

  log: (level: string, ...args) => {
    switch (level) {
      case 'warn':
        if (['warn', 'info', 'debug', 'trace'].includes(Rimdian.config.log_level)) {
          console.warn(...args)
        }
        break
      case 'info':
        if (['info', 'debug', 'trace'].includes(Rimdian.config.log_level)) {
          console.info(...args)
        }
        break
      case 'debug':
        if (['debug', 'trace'].includes(Rimdian.config.log_level)) {
          console.debug(...args)
        }
        break
      case 'trace':
        if (Rimdian.config.log_level === 'trace') {
          console.trace(...args)
        }
        break
      // print errors by default
      default:
        console.error(...args)
    }
  },

  // watch for DOM to be ready and call onReady()
  init: (cfg: IConfig) => {
    // continue if DOM is ready
    if (document.readyState === 'complete' || document.readyState === 'interactive') {
      Rimdian._onReady(cfg)
    }

    // watch for DOM readiness
    document.onreadystatechange = () => {
      if (document.readyState === 'interactive' || document.readyState === 'complete') {
        Rimdian.log('info', 'document is now', document.readyState)
        Rimdian._onReady(cfg)
      }
    }

    const logLevel = Rimdian.getCookie(Rimdian.config.namespace + 'debug')
    if (logLevel) {
      Rimdian.config.log_level = 'info'
    }
  },

  setDispatchConsent: (consent: boolean) => {
    Rimdian.log('info', 'RMD dispatch consent is now', consent)
    Rimdian.dispatchConsent = consent
  },

  // return callback when the user is ready
  getCurrentUser: (callback: (user: IUser) => void) => {
    // return callback when the user is ready
    if (Rimdian.currentUser) {
      callback(Rimdian.currentUser)
      return
    }

    Rimdian._execWhenReady(() => {
      callback(Rimdian.currentUser)
    })
  },

  onReady: (fn: () => void) => {
    if (Rimdian.isReady) {
      fn()
    } else {
      Rimdian.onReadyQueue.push(fn)
    }
  },

  // tracks the current pageview
  trackPageview: (data: any) => {
    if (!Rimdian.isReady) {
      Rimdian.log('debug', 'RMD is not yet ready for trackPageview, queuing function...')
      Rimdian._execWhenReady(() => Rimdian.trackPageview(data))
      return
    }

    // persist existing pageview duration if any, or discard it if not active
    // because the pageview duration is already persisted when pages become passive
    if (Rimdian.currentPageview && !Rimdian.currentPageviewVisibleSince) {
      Rimdian.log('info', 'previous pageview has been discarded, was not active')
      Rimdian.currentPageview = undefined
    }

    const now = new Date().toISOString()

    // check if a pageview is already tracked (Single Page App)
    if (Rimdian.currentPageview) {
      Rimdian.log('info', 'pageview already tracked')

      // compute pageview duration
      const increment = Math.round(
        (new Date().getTime() - Rimdian.currentPageviewVisibleSince.getTime()) / 1000
      )
      Rimdian.currentPageviewDuration += increment
      Rimdian.currentPageview.duration = Rimdian.currentPageviewDuration
      Rimdian.currentPageview.updated_at = now

      // increment session duration too
      if (!Rimdian.currentSession.duration) Rimdian.currentSession.duration = 0
      Rimdian.currentSession.duration += increment
      Rimdian.currentSession.updated_at = now
      // session is persisted in cookie below

      // add current pageview to items queue
      Rimdian.itemsQueue.add('pageview', { ...Rimdian.currentPageview })
      // reset timer
      Rimdian.currentPageviewVisibleSince = undefined
      Rimdian.currentPageviewDuration = 0
    }

    // deep clone pageview
    const pageview = JSON.parse(JSON.stringify(data || {}))

    pageview.external_id = Rimdian.uuidv4()
    pageview.created_at = now

    const referrer = Rimdian.getReferrer()
    if (referrer) {
      pageview.referrer = referrer
    }

    // set defaults
    if (!pageview.title) {
      pageview.title = document.title
    }
    if (!pageview.page_id) {
      pageview.page_id = window.location.href
    }

    // escape title for JSON
    pageview.title = pageview.title.replace(/\\"/, '"')

    // amount in cents
    if (pageview.product_price && pageview.product_price > 0) {
      pageview.product_price = Math.round(pageview.product_price * 100)
    }

    // init visibility tracking
    if (Rimdian.isPageVisible()) {
      Rimdian.currentPageviewVisibleSince = new Date()
    }

    // // error on invalid pageview
    // if (!ValidatePageview(pageview)) {
    //   Rimdian.log('error', 'invalid pageview', ValidatePageview.errors)
    // }

    // set current pageview
    Rimdian.currentPageview = pageview

    // update user last interaction
    Rimdian.currentUser.last_interaction_at = now
    Rimdian.setUserContext(Rimdian.currentUser) // cookie update

    // increment pageview count + interaction count of session
    if (!Rimdian.currentSession.pageviews_count) Rimdian.currentSession.pageviews_count = 0
    if (!Rimdian.currentSession.interactions_count) Rimdian.currentSession.interactions_count = 0
    Rimdian.currentSession.pageviews_count++
    Rimdian.currentSession.interactions_count++
    Rimdian.currentSession.updated_at = now
    Rimdian.setSessionContext(Rimdian.currentSession) // cookie update

    // enqueue pageview
    Rimdian.itemsQueue.add('pageview', { ...pageview })
  },

  // tracks the current customEvent
  trackCustomEvent: (data: any) => {
    if (!Rimdian.isReady) {
      Rimdian.log('debug', 'RMD is not yet ready for trackCustomEvent, queuing function...')
      Rimdian._execWhenReady(() => Rimdian.trackCustomEvent(data))
      return
    }

    if (data && !data.label) {
      Rimdian.log('error', 'customEvent label is required')
      return
    }

    // deep clone customEvent
    const customEvent = JSON.parse(JSON.stringify(data || {}))
    const now = new Date().toISOString()

    customEvent.external_id = Rimdian.uuidv4()
    customEvent.created_at = now

    // escape label for JSON
    customEvent.label = customEvent.label.replace(/\\"/, '"')

    if (customEvent.string_value) {
      // escape string_value for JSON
      customEvent.string_value = customEvent.string_value.replace(/\\"/, '"')
    }

    // // error on invalid customEvent
    // if (!ValidatePageview(customEvent)) {
    //   Rimdian.log('error', 'invalid customEvent', ValidatePageview.errors)
    // }

    // update user last interaction
    if (!customEvent.non_interactive) {
      Rimdian.currentUser.last_interaction_at = now
      Rimdian.setUserContext(Rimdian.currentUser) // cookie update

      // increment customEvent count + interaction count of session
      if (!Rimdian.currentSession.interactions_count) Rimdian.currentSession.interactions_count = 0
      Rimdian.currentSession.interactions_count++
      Rimdian.currentSession.updated_at = now
      Rimdian.setSessionContext(Rimdian.currentSession) // cookie update
    }

    // enqueue customEvent
    Rimdian.itemsQueue.add('custom_event', { ...customEvent })
  },

  trackCart: (data: ICart) => {
    if (!Rimdian.isReady) {
      Rimdian.log('debug', 'RMD is not yet ready for trackCart, queuing function...')
      Rimdian._execWhenReady(() => Rimdian.trackCart(data))
      return
    }

    // check if data exists and is an object
    if (!data || typeof data !== 'object') {
      Rimdian.log('error', 'invalid cart data')
      return
    }
    // deep clone cart
    const cart = JSON.parse(JSON.stringify(data)) as ICart

    Rimdian.log('info', 'cart is', cart)

    if (!cart.external_id) {
      cart.external_id = Rimdian.uuidv4()
    }

    cart.session_external_id = Rimdian.currentSession.external_id

    if (cart.items) {
      // convert item prices to cents
      cart.items.forEach((item: ICartItem) => {
        item.cart_external_id = cart.external_id
        if (item.price && item.price > 0) {
          item.price = Math.round(item.price * 100)
        }
      })
    }

    // // error on invalid cart
    // if (!ValidateCart(cart)) {
    //   Rimdian.log('error', 'invalid cart', ValidateCart.errors)
    // }

    // compute a cart hash if not provided
    if (!cart.hash) {
      cart.hash = Rimdian._cartHash(cart)
    }

    // check if a cart is already tracked (Single Page App)
    if (Rimdian.currentCart) {
      Rimdian.log('info', 'cart already tracked', { ...Rimdian.currentCart })

      // skip if cart hash is the same
      if (Rimdian.currentCart.hash === cart.hash) {
        return
      }
    }

    const now = new Date().toISOString()

    if (!cart.updated_at) {
      cart.updated_at = new Date().toISOString()
    }

    Rimdian.currentCart = cart

    // increment interaction count of session
    // update user last interaction
    Rimdian.currentUser.last_interaction_at = now

    if (!Rimdian.currentSession.interactions_count) {
      Rimdian.currentSession.interactions_count = 0
    }
    Rimdian.currentSession.interactions_count++
    Rimdian.currentSession.updated_at = now
    Rimdian.setSessionContext(Rimdian.currentSession) // cookie update

    Rimdian.itemsQueue.add('cart', cart)
  },

  trackOrder: (data: IOrder) => {
    if (!Rimdian.isReady) {
      Rimdian.log('debug', 'RMD is not yet ready for trackOrder, queuing function...')
      Rimdian._execWhenReady(() => Rimdian.trackOrder(data))
      return
    }

    // check if data exists and is an object
    if (!data || typeof data !== 'object') {
      Rimdian.log('error', 'invalid order data')
      return
    }

    // deep clone order
    const order = JSON.parse(JSON.stringify(data)) as IOrder

    // convert item prices to cents
    if (order.items) {
      order.items.forEach((item: IOrderItem) => {
        if (item.price && item.price > 0) {
          item.price = Math.round(item.price * 100)
        }
        item.order_external_id = order.external_id
      })
    }

    // // error on invalid order, but don't abort
    // if (!ValidateOrder(order)) {
    //   Rimdian.log('error', 'invalid order', ValidateOrder.errors)
    // }

    const now = new Date().toISOString()

    // update user last interaction
    Rimdian.currentUser.last_interaction_at = now
    Rimdian.setUserContext(Rimdian.currentUser) // cookie update

    // increment interaction count of session
    if (!Rimdian.currentSession.interactions_count) Rimdian.currentSession.interactions_count = 0
    Rimdian.currentSession.interactions_count++
    Rimdian.currentSession.updated_at = now
    Rimdian.setSessionContext(Rimdian.currentSession) // cookie update

    // enqueue order after user+session update
    Rimdian.itemsQueue.add('order', order)
  },

  setDeviceContext: (data: IDevice) => {
    if (!Rimdian.isReady) {
      Rimdian.log('debug', 'RMD is not yet ready for setDeviceContext, queuing function...')
      Rimdian._execWhenReady(() => Rimdian.setDeviceContext(data))
      return
    }

    // merge currentDevice with data
    const newDevice = { ...Rimdian.currentDevice, ...data }
    // enriching the device data should trigger a DB update with updated_at
    newDevice.updated_at = new Date().toISOString()

    // // validate current device
    // if (!ValidateDevice(Rimdian.currentDevice)) {
    //   Rimdian.log('error', 'invalid device', ValidateDevice.errors)
    //   return
    // }

    Rimdian.currentDevice = newDevice
    // persist updated device
    Rimdian.setCookie(
      Rimdian.config.namespace + 'device',
      JSON.stringify(Rimdian.currentDevice),
      OneYearInSeconds
    )
  },

  setSessionContext: (data: ISession) => {
    if (!Rimdian.isReady) {
      Rimdian.log('debug', 'RMD is not yet ready for setSessionContext, queuing function...')
      Rimdian._execWhenReady(() => Rimdian.setSessionContext(data))
      return
    }

    // merge currentSession with data
    const newSession = { ...Rimdian.currentSession, ...data }
    // enriching the device data should trigger a DB update with updated_at
    newSession.updated_at = new Date().toISOString()

    Rimdian.currentSession = newSession

    Rimdian.log('info', 'RMD updated session is', newSession)

    // persist updated session
    Rimdian.setCookie(
      Rimdian.config.namespace + 'session',
      JSON.stringify(Rimdian.currentSession),
      Rimdian.config.session_timeout
    )
  },

  // check if user ID has changed
  // reset context if new user is also authenticated
  // or user_alias if previous was anonymous
  // set new user fields and eventually enqueue user
  setUserContext: (data: IUser) => {
    if (!Rimdian.isReady) {
      Rimdian.log('debug', 'RMD is not yet ready for setUserContext, queuing function...')
      Rimdian._execWhenReady(() => Rimdian.setUserContext(data))
      return
    }

    // replace user_centric_consent by consent_all
    if (data && data.user_centric_consent !== undefined) {
      data.consent_all = data.user_centric_consent
      delete data.user_centric_consent
    }

    if (
      data &&
      data.external_id &&
      data.external_id !== '' &&
      data.external_id !== Rimdian.currentUser.external_id
    ) {
      // if previous user was authenticated, reset new session / device, as we can't alias 2 authenticated users
      if (Rimdian.currentUser.is_authenticated && data.is_authenticated === true) {
        Rimdian.log('info', 'new authenticated user detected, reset context', data)
        // reset device
        Rimdian.currentDevice = undefined
        Rimdian._localStorage.remove('device')
        Rimdian._createDevice()

        // reset current session after device
        Rimdian.currentSession = undefined
        Rimdian.deleteCookie(Rimdian.config.namespace + 'session')
        Rimdian._handleSession()

        // loop over keys and ignore empty strings
        Object.keys(data).forEach((key) => {
          if (typeof data[key] === 'string' && data[key] === '') {
            delete data[key]
          }
        })

        Rimdian.currentUser = { ...data }

        // set defaults
        if (Rimdian.currentUser.created_at === undefined) {
          Rimdian.currentUser.created_at = new Date().toISOString()
        }

        // // validate current user
        // if (!ValidateUser(Rimdian.currentUser)) {
        //   Rimdian.log('error', 'invalid user', ValidateUser.errors)
        // }

        return
      }

      // create a user_alias if previous user was anonymous
      // and merge new user data & enqueue user
      if (!Rimdian.currentUser.is_authenticated) {
        Rimdian.log(
          'info',
          'alias previous user',
          Rimdian.currentUser.external_id,
          'with user',
          data
        )

        Rimdian.itemsQueue.add('user_alias', {
          from_user_external_id: Rimdian.currentUser.external_id,
          to_user_external_id: data.external_id,
          to_user_is_authenticated: data.is_authenticated === true,
          to_user_created_at: data.created_at ? data.created_at : new Date().toISOString()
        })
      }
    }

    // update user context with new data
    const newUser = { ...Rimdian.currentUser, ...data }

    // // validate current user
    // if (!ValidateUser(Rimdian.currentUser)) {
    //   // abort on error
    //   Rimdian.log('error', 'invalid user', ValidateUser.errors)
    //   return
    // }

    Rimdian.currentUser = newUser

    // persist updated user
    Rimdian.setCookie(
      Rimdian.config.namespace + 'user',
      JSON.stringify(Rimdian.currentUser),
      OneYearInSeconds
    )
  },

  // enqueue the current user
  saveUserProfile: () => {
    if (!Rimdian.isReady) {
      Rimdian.log('debug', 'RMD is not yet ready for saveUserProfile, queuing function...')
      Rimdian._execWhenReady(() => Rimdian.saveUserProfile())
      return
    }

    Rimdian.itemsQueue.add('user', { ...Rimdian.currentUser })
  },

  isPageVisible: () => {
    return lifecycle.state === PageStates.active
  },

  // dispatch items to the collector
  dispatch: (useBeacon: boolean) => {
    // wait for tracker to be ready
    if (Rimdian.isReady === false) {
      Rimdian.log('info', 'RMD is not ready, retrying dispatch soon...')

      window.setTimeout(function () {
        Rimdian.dispatch(useBeacon)
      }, 50)
      return
    }

    // abort if we don't have user consent
    if (!Rimdian.dispatchConsent) {
      Rimdian.log('info', 'RMD abort dispatch, no dispatch consent')
      return
    }

    // abort if we don't have items to dispatch
    if (Rimdian.itemsQueue.items.length === 0 && Rimdian.dispatchQueue.length === 0) {
      Rimdian.log('info', 'RMD abort dispatch, no items to dispatch')
      return
    }

    const deviceCtx = { ...Rimdian.currentDevice }
    // set updated_at only if user_agent changed, it will trigger a DB update
    if (deviceCtx.user_agent !== navigator.userAgent) {
      deviceCtx.updated_at = new Date().toISOString()
    }
    deviceCtx.user_agent = navigator.userAgent
    deviceCtx.language = navigator.language
    deviceCtx.ad_blocker = Rimdian.hasAdBlocker()
    deviceCtx.resolution =
      window.screen && window.screen.width && window.screen.height
        ? window.screen.height > window.screen.width
          ? window.screen.height + 'x' + window.screen.width
          : window.screen.width + 'x' + window.screen.height
        : undefined

    // create dataImport batches of 20 items maximum
    const batches = []
    let itemsBatch: IItem[] = []

    while (Rimdian.itemsQueue.items.length > 0) {
      itemsBatch.push(Rimdian.itemsQueue.items.shift())
      if (itemsBatch.length >= 20) {
        batches.push(itemsBatch)
        itemsBatch = []
      }
    }

    // add remaining items to last batch
    if (itemsBatch.length > 0) {
      batches.push(itemsBatch)
    }

    // convert batches into IDataImport objects
    batches.forEach((batch: IItem[]) => {
      // add device to items
      batch.forEach((item: IItem) => {
        item.device = deviceCtx
      })

      Rimdian.dispatchQueue.push({
        id: Rimdian.uuidv4(),
        workspace_id: Rimdian.config.workspace_id,
        items: batch,
        context: {
          // set this field right before sending data
          // data_sent_at: new Date().toISOString()
        },
        created_at: new Date().toISOString()
      })
    })

    // abort if we don't have data imports to dispatch
    if (Rimdian.dispatchQueue.length === 0) {
      Rimdian.log('info', 'RMD abort dispatch, no data imports to dispatch')
      return
    }

    // persist dispatch queue in local storage
    Rimdian._localStorage.set('dispatchQueue', JSON.stringify(Rimdian.dispatchQueue))

    // send data to collector
    Rimdian.log('info', 'RMD sending data to collector')

    Rimdian._initDispatchLoop(useBeacon)
  },

  _initDispatchLoop: (useBeacon: boolean) => {
    // abort if we are already dispatching
    if (Rimdian.isDispatching) {
      Rimdian.log('info', 'RMD abort dispatch, already dispatching')
      return
    }

    // dispatch items
    Rimdian.isDispatching = true

    // post the hits, starting with the oldest batch first if it exists
    Rimdian.dispatchQueue.sort((a, b) => {
      if (a.created_at < b.created_at) {
        return -1 // a listed before b
      }
      if (a.created_at > b.created_at) {
        return 1
      }
      return 0
    })

    const currentBatch = Rimdian.dispatchQueue[0]

    // post payload
    Rimdian._postPayload(currentBatch, 0, useBeacon)
  },

  _postPayload: (dataImport: IDataImport, retryCount: number, useBeacon: boolean) => {
    Rimdian._post(dataImport, useBeacon, (error) => {
      var success = true

      if (error) {
        // retry or requeue
        Rimdian.log('error', 'RMD post payload error', error)

        // parse error, ignore error 400
        try {
          var jsonError = JSON.parse(error) || {}
          // error is not a bad hit
          if (jsonError.code > 400) {
            success = false
          }
        } catch (err) {
          success = false
        }

        Rimdian.log('info', 'sucess', success, 'retry count', retryCount)

        if (success === false) {
          // stop after 10 retry
          if (retryCount >= Rimdian.config.max_retry) {
            Rimdian.isDispatching = false
            Rimdian.log('info', 'max retry reached, aborting')
            return
          }

          // exponential backoff
          var delay = 100 // ms
          var toWait = delay

          if (retryCount > 1) {
            for (var i = 0; i < retryCount; i++) {
              toWait = toWait * 2
            }
          }

          Rimdian.log('debug', 'retry in ' + toWait + 'ms')

          // wait 500 ms
          window.setTimeout(() => {
            retryCount++
            Rimdian._postPayload(dataImport, retryCount, useBeacon)
          }, toWait)

          return
        }
      }

      if (success === true) {
        // remove from the dispatch queue

        const remainingBatches = []

        // keep other batches
        Rimdian.dispatchQueue.forEach((di) => {
          if (di.id !== dataImport.id) {
            remainingBatches.push(di)
          }
        })

        Rimdian.dispatchQueue = remainingBatches

        // persist dispatch queue in local storage
        Rimdian._localStorage.set('dispatchQueue', JSON.stringify(Rimdian.dispatchQueue))

        // continue the loop
        Rimdian.isDispatching = false

        if (Rimdian.dispatchQueue.length > 0) {
          Rimdian._initDispatchLoop(useBeacon)
        }
      }
    })
  },

  _post: (dataImport: IDataImport, useBeacon: boolean, callback: (error: string) => void) => {
    // set client clock right before sending data
    dataImport.context.data_sent_at = new Date().toISOString()
    const data = JSON.stringify(dataImport)

    // log info data
    Rimdian.log('info', 'RMD sending data', data)

    // send dataImport to collector using beacon
    if (useBeacon && navigator.sendBeacon) {
      const queued = navigator.sendBeacon(
        Rimdian.config.host + '/live',
        new Blob([data], { type: 'application/json' })
      )
      return callback(queued ? null : 'sendBeacon failed')
    }

    const xhr: XMLHttpRequest = new XMLHttpRequest()

    xhr.onload = function () {
      const body = 'response' in xhr ? xhr.response : xhr['responseText']
      if (xhr.status >= 300) {
        return callback(body)
      }
      return callback(null)
    }

    xhr.onerror = function () {
      return callback('network request failed')
    }

    xhr.ontimeout = function () {
      return callback('network request timeout')
    }

    xhr.open('POST', Rimdian.config.host + '/live', true)
    xhr.setRequestHeader('Content-Type', 'application/json')
    xhr.withCredentials = true
    xhr.send(data)
  },

  // retrieve existing user ID or create a new one
  // 1. id found in URL
  // 2. OR id exists in a cookie
  // 3. OR id is created
  _handleUser: () => {
    // sometimes user IDs are injected by template engines and are not parsed - ie: {{ user_id }}
    // to avoid using/merging invalid user IDs, we forbid some patterns
    const forbiddenPatterns = ['{{', '}}', '{%', '%}', '{#', '#}', '*|', '|*']
    let userCookie = Rimdian.getCookie(Rimdian.config.namespace + 'user')
    let previousUser: any = {}

    if (userCookie && userCookie !== '') {
      previousUser = JSON.parse(userCookie)

      // ignore cookies where the user ID was a forbidden pattern
      // this is used to erase legacy cookies that allowed fordidden patterns in the past
      if (
        forbiddenPatterns.some(
          (pattern) => previousUser.external_id && previousUser.external_id.indexOf(pattern) !== -1
        )
      ) {
        previousUser = {}
        userCookie = ''
      }

      // we have a legacy cookie, migrate it
      if (
        Rimdian.config.from_cm === true &&
        previousUser.id !== undefined &&
        previousUser.external_id === undefined
      ) {
        previousUser = {
          external_id: previousUser.id,
          is_authenticated: previousUser.is_authenticated || false,
          created_at: new Date().toISOString(), // this date will be ignored if the user already existed in the DB
          hmac: previousUser.hmac || undefined
        }
      }
    }

    // an authenticated user id can be the param "_authuid" or the combination of "_uid" + "_auth=true"
    const authenticatedUserId = Rimdian.getQueryParam(
      document.URL,
      URLParams.user_authenticated_external_id
    )
    let userId =
      authenticatedUserId || Rimdian.getQueryParam(document.URL, URLParams.user_external_id)

    // 1. ID found in URL
    // check if the user ID does not contain forbidden patterns
    if (
      userId &&
      userId !== '' &&
      forbiddenPatterns.every((pattern) => userId.indexOf(pattern) === -1)
    ) {
      Rimdian.log('info', 'found user ID in URL', userId)

      const isAuthenticated = authenticatedUserId
        ? 'true'
        : Rimdian.getQueryParam(document.URL, URLParams.user_is_authenticated)
      Rimdian.log('info', 'user authenticated URL value', isAuthenticated)

      Rimdian.currentUser = {
        external_id: userId,
        is_authenticated: isAuthenticated === 'true' || isAuthenticated === '1',
        created_at: new Date().toISOString(),
        hmac: Rimdian.getQueryParam(document.URL, URLParams.user_external_id_hmac)
      }

      // alias previous user if was unknown

      if (userCookie && userCookie !== '') {
        Rimdian.log('info', 'found another user ID in cookie', userCookie)

        if (
          previousUser.is_authenticated === false &&
          previousUser.external_id !== Rimdian.currentUser.external_id
        ) {
          Rimdian.log(
            'info',
            'alias previous user',
            previousUser.external_id,
            'with user',
            Rimdian.currentUser.external_id
          )

          Rimdian.itemsQueue.add('user_alias', {
            from_user_external_id: previousUser.external_id,
            to_user_external_id: Rimdian.currentUser.external_id,
            to_user_is_authenticated: Rimdian.currentUser.is_authenticated,
            to_user_created_at: Rimdian.currentUser.created_at
          })
        }
      }

      Rimdian.setCookie(
        Rimdian.config.namespace + 'user',
        JSON.stringify(Rimdian.currentUser),
        OneYearInSeconds
      )
      return
    }

    // 2. ID found in cookie

    if (userCookie && userCookie !== '') {
      Rimdian.log('info', 'found user ID in cookie', userCookie)
      Rimdian.currentUser = previousUser
      Rimdian.setCookie(
        Rimdian.config.namespace + 'user',
        JSON.stringify(Rimdian.currentUser),
        OneYearInSeconds
      )
      return
    }

    // 3. ID is created, with anonymous user
    Rimdian._createUser(Rimdian.uuidv4(), false, new Date().toISOString())
  },

  _createUser: (userExternalId: string, isAuthenticated: boolean, createdAt: string) => {
    Rimdian.currentUser = {
      external_id: userExternalId,
      is_authenticated: isAuthenticated,
      created_at: createdAt
    }
    Rimdian.log('info', 'creating new user', { ...Rimdian.currentUser })
    Rimdian.setCookie(
      Rimdian.config.namespace + 'user',
      JSON.stringify(Rimdian.currentUser),
      OneYearInSeconds
    )
  },

  // retrieve eventual user details provided in URL
  _enrichUserContext: () => {
    // update current user profile with other parameters
    const otherUserIds = [
      { key: 'email', value: Rimdian.getQueryParam(document.URL, '_email') },
      { key: 'email_md5', value: Rimdian.getQueryParam(document.URL, '_emailmd5') },
      { key: 'email_sha1', value: Rimdian.getQueryParam(document.URL, '_emailsha1') },
      { key: 'email_sha256', value: Rimdian.getQueryParam(document.URL, '_emailsha256') },
      { key: 'telephone', value: Rimdian.getQueryParam(document.URL, '_telephone') }
    ]

    // enrich user context with other parameters
    otherUserIds.forEach((x, i) => {
      if (x.value && x.value !== '') {
        Rimdian.currentUser[x.key] = x.value
      }
    })

    // update user cookie
    Rimdian.setCookie(
      Rimdian.config.namespace + 'user',
      JSON.stringify(Rimdian.currentUser),
      OneYearInSeconds
    )
  },

  // retrieve existing device ID or create a new one
  // 1. id found in URL
  // 2. id exists in a cookie
  // 3. id is created
  // to save cookie space, device context is enriched while dispatching data
  _handleDevice: () => {
    let deviceId = Rimdian.getQueryParam(document.URL, URLParams.device_external_id)

    // 1. ID found in URL
    if (deviceId && deviceId !== '') {
      Rimdian.log('info', 'found device ID in URL', deviceId)
      Rimdian.currentDevice = {
        external_id: deviceId,
        created_at: new Date().toISOString(),
        user_agent: navigator.userAgent
      }
      Rimdian.setCookie(
        Rimdian.config.namespace + 'device',
        JSON.stringify(Rimdian.currentDevice),
        OneYearInSeconds
      )
      return
    }

    // 2. ID found in cookie
    const deviceCookie = Rimdian.getCookie(Rimdian.config.namespace + 'device')

    if (deviceCookie && deviceCookie !== '') {
      Rimdian.log('info', 'found device ID in cookie', deviceCookie)
      Rimdian.currentDevice = JSON.parse(deviceCookie)

      // check if we already had a legacy client ID in a cookie
      if (Rimdian.config.from_cm === true) {
        const legacyClientID = Rimdian.getCookie('_cm_cid')

        if (legacyClientID && legacyClientID !== '') {
          Rimdian.currentDevice.external_id = legacyClientID
          // extract its creation timestamp
          const legacyClientTimestamp = Rimdian.getCookie('_cm_cidat')
          if (legacyClientTimestamp && legacyClientTimestamp !== '') {
            Rimdian.currentDevice.created_at = new Date(
              parseInt(legacyClientTimestamp, 10)
            ).toISOString()
          }
        }
      }

      Rimdian.setCookie(
        Rimdian.config.namespace + 'device',
        JSON.stringify(Rimdian.currentDevice),
        OneYearInSeconds
      )
      return
    }

    // 3. ID is created
    Rimdian._createDevice()
  },

  _createDevice: () => {
    Rimdian.log('info', 'creating new device ID')
    Rimdian.currentDevice = {
      external_id: Rimdian.uuidv4(),
      created_at: new Date().toISOString(),
      user_agent: navigator.userAgent
    }
    Rimdian.setCookie(
      Rimdian.config.namespace + 'device',
      JSON.stringify(Rimdian.currentDevice),
      OneYearInSeconds
    )
  },

  // polyfill
  _addEventListener: (element, eventType, eventHandler, useCapture) => {
    if (element.addEventListener) {
      element.addEventListener(eventType, eventHandler, useCapture)
      return true
    }
    if (element.attachEvent) {
      return element.attachEvent('on' + eventType, eventHandler)
    }
    element['on' + eventType] = eventHandler
  },

  // load config, existing data and start tracking
  _onReady: (cfg: IConfig) => {
    // avoid calling onReady twice
    if (Rimdian.isReady) return
    Rimdian.isReady = true

    Rimdian.log('info', 'onReady() called')

    // check if browser is legit
    if (!Rimdian.isBrowserLegit()) {
      Rimdian.log('warn', 'Browser is not legit')
      return
    }

    // merge cfg with default config & validate
    const config = { ...Rimdian.config, ...cfg }

    // if (!ValidateConfig(config)) {
    //   Rimdian.log('error', 'RMD Config error:', ValidateConfig.errors)
    //   return
    // }

    // save config
    Rimdian.config = config
    Rimdian.log('info', 'RMD Config is:', Rimdian.config)

    // load items and batches from localStorage
    if (window.localStorage) {
      Rimdian.itemsQueue.items = JSON.parse(
        localStorage.getItem(Rimdian.config.namespace + 'items') || '[]'
      )
      Rimdian.dispatchQueue = JSON.parse(
        localStorage.getItem(Rimdian.config.namespace + 'dispatchQueue') || '[]'
      )
    }

    // init user context
    Rimdian._handleUser()
    Rimdian._enrichUserContext()
    // init device context
    Rimdian._handleDevice()
    // init session context after device
    Rimdian._handleSession()

    // execute queued functions
    if (Rimdian.onReadyQueue.length > 0) {
      Rimdian.onReadyQueue.forEach((x, i) => {
        Rimdian.log('debug', 'executing queued function', x)
        x()
      })

      Rimdian.onReadyQueue = []
    }

    // validate user / device / session
    // if (!ValidateUser(Rimdian.currentUser)) {
    //   Rimdian.log('error', 'invalid user', ValidateUser.errors)
    // }

    // if (!ValidateDevice(Rimdian.currentDevice)) {
    //   Rimdian.log('error', 'invalid device', ValidateDevice.errors)
    // }

    // if (!ValidateSession(Rimdian.currentSession)) {
    //   Rimdian.log('error', 'invalid session', ValidateSession.errors)
    // }

    // extend session expiration time every minute while pageview is visible
    window.setInterval(function () {
      if (Rimdian.currentPageview && Rimdian.isPageVisible()) {
        // get current session from cookie to see if it expired
        const cookieSession = Rimdian.getCookie(Rimdian.config.namespace + 'session')
        if (cookieSession) {
          // extend lifetime of session
          Rimdian.setCookie(
            Rimdian.config.namespace + 'session',
            cookieSession,
            Rimdian.config.session_timeout
          )
        } else {
          // don't start new session here, the user might want to close the tab or navigate away
          // that's why we have to check if the session is expired on .pageview() / .customEvent() calls
        }
      }
    }, 60000) // every minute

    // decorate cross domains links
    if (Rimdian.config.cross_domains.length > 0) {
      // loop over every <a> tag and add the decorateURL function
      for (var i = 0; i < document.links.length; i++) {
        var elt = document.links[i]

        Rimdian.config.cross_domains.forEach((d) => {
          // only decorate links to the matching domain
          if (elt.href.indexOf(d) !== -1) {
            Rimdian._addEventListener(elt, 'click', Rimdian._decorateURL, true)
            Rimdian._addEventListener(elt, 'mousedown', Rimdian._decorateURL, true)
          }
        })
      }
    }

    // use visibilitychange event to send pageview secs in beacon
    // use https://github.com/GoogleChromeLabs/page-lifecycle
    // https://developer.chrome.com/blog/page-lifecycle-api/#advice-hidden

    // watch page state changes
    lifecycle.addEventListener('statechange', (event) => {
      Rimdian.log('info', 'page state changed from', event.oldState, 'to', event.newState)

      if (event.oldState === PageStates.active && event.newState === PageStates.passive) {
        Rimdian._onPagePassive()
      } else if (event.oldState === PageStates.passive && event.newState === PageStates.active) {
        Rimdian._onPageActive()
      }
    })

    // dispatch items  automatically every 5 secs
    // window.setInterval(function () {
    //   Rimdian.dispatch(false)
    // }, 5000)
  },

  _execWhenReady: (fn: Function) => {
    if (Rimdian.isReady) {
      fn()
    } else {
      Rimdian.onReadyQueue.push(fn)
    }
  },

  _normalizeUTMSource: (source: string) => {
    // replace 98ad0bb6e8ada73c81aab4e8c2637e7f.safeframe.googlesyndication.com by safeframe.googlesyndication.com
    if (source.indexOf('safeframe.googlesyndication.com') !== -1) {
      return 'safeframe.googlesyndication.com'
    }

    return source
  },

  // session is stored in a cookie and will expire with its cookie
  // scenarii:
  // 1. no existing session -> create new session
  // 2. existing session with same referrer origin -> continue session
  // 3. existing session with different referrer origin -> create new session
  _handleSession: () => {
    // extract current utm params from url
    let utm_source =
      Rimdian.getQueryParam(document.URL, 'utm_source') ||
      Rimdian.getHashParam(window.location.hash, 'utm_source')
    let utm_medium =
      Rimdian.getQueryParam(document.URL, 'utm_medium') ||
      Rimdian.getHashParam(window.location.hash, 'utm_medium')
    let utm_campaign =
      Rimdian.getQueryParam(document.URL, 'utm_campaign') ||
      Rimdian.getHashParam(window.location.hash, 'utm_campaign')
    let utm_content =
      Rimdian.getQueryParam(document.URL, 'utm_content') ||
      Rimdian.getHashParam(window.location.hash, 'utm_content')
    let utm_term =
      Rimdian.getQueryParam(document.URL, 'utm_term') ||
      Rimdian.getHashParam(window.location.hash, 'utm_term')
    let utm_id =
      Rimdian.getQueryParam(document.URL, 'utm_id') ||
      Rimdian.getHashParam(window.location.hash, 'utm_id')
    let utm_id_from =
      Rimdian.getQueryParam(document.URL, 'utm_id_from') ||
      Rimdian.getHashParam(window.location.hash, 'utm_id_from')

    // extract referrer from url
    const referrer = Rimdian.getReferrer()

    if (referrer) {
      // parse referrer URL with <a> tag
      var referrerURL = document.createElement('a')
      referrerURL.href = referrer

      // check if referrer came from another domain
      let fromAnotherDomain = false

      // check if different domain is not is the cross_domain config
      if (referrerURL.hostname && referrerURL.hostname !== window.location.hostname) {
        let isCrossDomain = false

        if (Rimdian.config.cross_domains && Rimdian.config.cross_domains.length) {
          Rimdian.config.cross_domains.forEach((dom) => {
            if (referrerURL.href.indexOf(dom) !== -1) {
              isCrossDomain = true
            }
          })
        }

        if (isCrossDomain === false) {
          fromAnotherDomain = true
        }
      }

      if (fromAnotherDomain) {
        utm_source = referrerURL.hostname

        // utm_medium is referral by default
        if (!utm_medium || utm_medium === '') {
          utm_medium = 'referral'
        }

        // check if it comes from known search engines and extract search query if possible
        if (referrer.search('https?://(.*)google.([^/?]*)') === 0) {
          utm_term = Rimdian.getQueryParam(referrer, 'q')
        } else if (referrer.search('https?://(.*)bing.com') === 0) {
          utm_term = Rimdian.getQueryParam(referrer, 'q')
        } else if (referrer.search('https?://(.*)search.yahoo.com') === 0) {
          utm_term = Rimdian.getQueryParam(referrer, 'p')
        } else if (referrer.search('https?://(.*)ask.com') === 0) {
          utm_term = Rimdian.getQueryParam(referrer, 'q')
        } else if (referrer.search('https?://(.*)search.aol.com') === 0) {
          utm_term = Rimdian.getQueryParam(referrer, 'q')
        } else if (referrer.search('https?://(.*)duckduckgo.com') === 0) {
          utm_term = Rimdian.getQueryParam(referrer, 'q')
        }
      }
    }

    // extract gclid+fbclid+MSCLKID from url into utm_id + utm_id_from
    const ids = ['gclid', 'fbclid', 'msclkid', 'aecid']
    ids.forEach((param) => {
      const value =
        Rimdian.getQueryParam(document.URL, param) ||
        Rimdian.getHashParam(window.location.hash, param)
      if (value) {
        utm_id = value
        utm_id_from = param
      }
    })

    // google SEO & google Ads might both use the same source/medium
    // we detect Ads with gclid, and change its medium to
    // eventually trigger a new session
    if (utm_id_from === 'gclid' && utm_medium === 'referral') {
      utm_medium = 'ads'
    }

    utm_source = Rimdian._normalizeUTMSource(utm_source)

    Rimdian.log('info', 'RMD utm_source is:', utm_source)
    Rimdian.log('info', 'RMD utm_medium is:', utm_medium)
    Rimdian.log('info', 'RMD utm_campaign is:', utm_campaign)
    Rimdian.log('info', 'RMD utm_content is:', utm_content)
    Rimdian.log('info', 'RMD utm_term is:', utm_term)
    Rimdian.log('info', 'RMD utm_id is:', utm_id)
    Rimdian.log('info', 'RMD utm_id_from is:', utm_id_from)

    // read session cookie
    const sessionCookie = Rimdian.getCookie(Rimdian.config.namespace + 'session')

    // 1. no existing session -> create new session
    if (!sessionCookie || sessionCookie === '') {
      Rimdian.log('info', 'RMD session cookie not found')

      Rimdian._startNewSession({
        utm_source,
        utm_medium,
        utm_campaign,
        utm_content,
        utm_term,
        utm_id,
        utm_id_from
      })
      return
    }

    // check if this origin should be ignored
    let ignoredOrigin

    if (utm_source && utm_source !== '' && Rimdian.config.ignored_origins.length > 0) {
      // find a matching origin
      ignoredOrigin = Rimdian.config.ignored_origins.find((origin) => {
        // source medium matches
        if (origin.utm_source === utm_source && origin.utm_medium === utm_medium) {
          // if origin requires a campaign, check if it matches
          if (origin.utm_campaign && origin.utm_campaign !== '') {
            if (utm_campaign && origin.utm_campaign === utm_campaign) {
              return true
            }
            // origin is not matching, continue
            return false
          }
          // if origin does not require a campaign, its a match
          return true
        }
        return false
      })
    }

    // process existing session
    let existingSession = JSON.parse(sessionCookie)
    Rimdian.log('info', 'RMD existing session is:', existingSession)

    // check if session origin has changed from previous page
    let isEqual = true
    if (utm_source && utm_source !== '' && existingSession.utm_source !== utm_source)
      isEqual = false
    if (utm_medium && utm_medium !== '' && existingSession.utm_medium !== utm_medium)
      isEqual = false
    if (utm_campaign && utm_campaign !== '' && existingSession.utm_campaign !== utm_campaign)
      isEqual = false
    if (utm_content && utm_content !== '' && existingSession.utm_content !== utm_content)
      isEqual = false
    if (utm_term && utm_term !== '' && existingSession.utm_term !== utm_term) isEqual = false
    if (utm_id && utm_id !== '' && existingSession.utm_id !== utm_id) isEqual = false

    // 2. if this origin is ignored, or same origin, or empty origin, resume session
    if (ignoredOrigin || isEqual || !utm_source || utm_source === '') {
      Rimdian.log(
        'info',
        'RMD resume session (ignored:' +
          (ignoredOrigin ? 'yes' : 'no') +
          ', isEqual:' +
          isEqual +
          ', utm_source:' +
          utm_source +
          ')'
      )
      Rimdian.currentSession = existingSession
      Rimdian.setCookie(
        Rimdian.config.namespace + 'session',
        JSON.stringify(Rimdian.currentSession),
        Rimdian.config.session_timeout
      )
      return
    }

    // 3. origin has changed, start new session
    Rimdian._startNewSession({
      utm_source,
      utm_medium,
      utm_campaign,
      utm_content,
      utm_term,
      utm_id,
      utm_id_from
    })
  },

  _onPagePassive: () => {
    Rimdian.log('info', 'page is passive state')
    Rimdian.itemsQueue.addPageviewDuration()
    Rimdian.dispatch(true) // use beacon as the window might be closing
  },

  _onPageActive: () => {
    Rimdian.log('info', 'page is active state')

    // abort if we are not tracking the current pageview
    if (!Rimdian.currentPageview) {
      return
    }

    Rimdian.currentPageviewVisibleSince = new Date() // reset the timer
  },

  getTimezone: () => {
    const DateTimeFormat = window.Intl?.DateTimeFormat
    if (DateTimeFormat) {
      const timezone = new DateTimeFormat().resolvedOptions().timeZone
      if (timezone) {
        return timezone
      }
    }
    return undefined
  },

  getQueryParam: (url: string, name: string) => {
    try {
      var urlObject = new URL(url)
      var params = new URLSearchParams(urlObject.search)
      return params.get(name) || undefined
    } catch (e) {
      return undefined
    }
  },

  getHashParam: (hash: string, name: string) => {
    var matches = hash.match(new RegExp(name + '=([^&]*)'))
    return matches ? matches[1] : undefined
  },

  updateURLParam: (url: string, name: string, value: string) => {
    var urlObject = new URL(url)
    var params = new URLSearchParams(urlObject.search)
    params.set(name, value)
    urlObject.search = params.toString()
    return urlObject.toString()
  },

  hasAdBlocker: () => {
    const ads = document.createElement('div')
    ads.innerHTML = '&nbsp;'
    ads.className = 'adsbox'
    let blocked = false
    try {
      // body may not exist, that's why we need try/catch
      document.body.appendChild(ads)
      blocked = (document.getElementsByClassName('adsbox')[0] as HTMLElement).offsetHeight === 0
      document.body.removeChild(ads)
    } catch (_e) {
      blocked = false
    }
    return blocked
  },

  isBrowserLegit: () => {
    // detect IE 9
    var ua = navigator.userAgent.toLowerCase()
    if (ua.indexOf('msie') !== -1) {
      if (parseInt(ua.split('msie')[1], 10) <= 9) {
        return false
      }
    }

    // detect known bot
    if (
      /(google web preview|baiduspider|yandexbot|bingbot|googlebot|yahoo! slurp|nuhk|yammybot|openbot|slurp|msnBot|ask jeeves\/teoma|ia_archiver)/i.test(
        navigator.userAgent
      )
    ) {
      return false
    }

    // detect headless chrome
    if (navigator.webdriver) {
      return false
    }
    return true
  },

  uuidv4: () => {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
      var r = (Math.random() * 16) | 0,
        v = c == 'x' ? r : (r & 0x3) | 0x8
      return v.toString(16)
    })
  },

  md5: (str: string) => {
    return md5(str)
  },

  getReferrer: () => {
    var referrer = undefined
    try {
      referrer = window.top.document.referrer !== '' ? window.top.document.referrer : undefined
    } catch (e) {
      if (window.parent) {
        try {
          referrer =
            window.parent.document.referrer !== '' ? window.parent.document.referrer : undefined
        } catch (_e) {
          referrer = undefined
        }
      }
    }
    if (!referrer) {
      referrer = document.referrer !== '' ? document.referrer : undefined
    }
    return referrer
  },

  _startNewSession: (params: ISessionDTO) => {
    // default to direct / none on empty origin
    if (!params.utm_source || params.utm_source === '') {
      params.utm_source = 'direct'
    }
    if (!params.utm_medium || params.utm_medium === '') {
      params.utm_medium = 'none'
    }

    Rimdian.currentSession = {
      external_id: Rimdian.uuidv4(),
      created_at: new Date().toISOString(),
      device_external_id: Rimdian.currentDevice.external_id,
      landing_page: window.location.href,
      referrer: Rimdian.getReferrer(),
      timezone: Rimdian.getTimezone(),

      utm_source: params.utm_source,
      utm_medium: params.utm_medium,
      utm_campaign: params.utm_campaign,
      utm_content: params.utm_content,
      utm_term: params.utm_term,
      utm_id: params.utm_id,
      utm_id_from: params.utm_id_from,

      duration: 0,
      pageviews_count: 0,
      interactions_count: 0
    }

    Rimdian.log('info', 'RMD new session is:', Rimdian.currentSession)

    // persist session to cookie
    Rimdian.setCookie(
      Rimdian.config.namespace + 'session',
      JSON.stringify(Rimdian.currentSession),
      Rimdian.config.session_timeout
    )
  },

  getCookie: (name: string) => {
    return (
      decodeURIComponent(
        document.cookie.replace(
          new RegExp(
            '(?:(?:^|.*;)\\s*' +
              encodeURIComponent(name).replace(/[-.+*]/g, '\\$&') +
              '\\s*\\=\\s*([^;]*).*$)|^.*$'
          ),
          '$1'
        )
      ) || null
    )
  },

  // cookies are secured and cross-domain by default
  setCookie: (name: string, value: string, seconds: number) => {
    // cross_domain
    const matches = window.location.hostname.match(/[a-z0-9][a-z0-9\-]+\.[a-z\.]{2,6}$/i)
    const domain = matches ? matches[0] : ''
    const xdomain = domain ? '; domain=.' + domain : ''

    const now = new Date()
    now.setTime(now.getTime() + seconds * 1000)
    const expires = '; expires=' + now.toUTCString()

    const cookie_value =
      name + '=' + encodeURIComponent(value) + expires + '; path=/' + xdomain + '; secure'
    document.cookie = cookie_value
    return
  },

  deleteCookie: (name: string) => {
    Rimdian.setCookie(name, '', -1)
  },

  _localStorage: {
    get: (key: string) => {
      return localStorage.getItem(Rimdian.config.namespace + key)
    },
    set: (key: string, value: string) => {
      try {
        localStorage.setItem(Rimdian.config.namespace + key, value)
      } catch (e) {
        Rimdian.log('error', 'localStorage error:', e)
      }
    },
    remove: (key: string) => {
      localStorage.removeItem(Rimdian.config.namespace + key)
    }
  },

  // inject the device + user ids on the fly
  _decorateURL: (e: MouseEvent) => {
    const target = e.target as HTMLAnchorElement
    target.href = Rimdian.updateURLParam(
      target.href,
      URLParams.device_external_id,
      Rimdian.currentDevice.external_id
    )
    target.href = Rimdian.updateURLParam(
      target.href,
      URLParams.user_external_id,
      Rimdian.currentUser.external_id
    )
    target.href = Rimdian.updateURLParam(
      target.href,
      URLParams.user_is_authenticated,
      Rimdian.currentUser.is_authenticated.toString()
    )
    if (Rimdian.currentUser.hmac) {
      target.href = Rimdian.updateURLParam(
        target.href,
        URLParams.user_external_id_hmac,
        Rimdian.currentUser.hmac
      )
    }
  },

  // the cart hash is a combination of public_url + products id + items variant id + items quantity
  _cartHash: (data: ICart) => {
    let cartHash = data.public_url ? data.public_url : ''

    if (data.items && data.items.length > 0) {
      data.items.forEach((item) => {
        cartHash =
          cartHash +
          item.product_external_id +
          (item.variant_external_id ? item.variant_external_id : '') +
          (item.quantity || '0')
      })
    }

    return md5(cartHash)
  },

  _wipeAll: () => {
    // create an alert and clear cookies and localtorage on confirmation
    if (window.confirm('Do you know what you are doing?')) {
      // clear cookies
      Rimdian.deleteCookie(Rimdian.config.namespace + 'device')
      Rimdian.deleteCookie(Rimdian.config.namespace + 'user')
      Rimdian.deleteCookie(Rimdian.config.namespace + 'session')
      // clear localstorage
      Rimdian._localStorage.remove('items')
      Rimdian._localStorage.remove('dispatchQueue')
      // reinitialize
      Rimdian.currentUser = undefined
      Rimdian.currentDevice = undefined
      Rimdian.currentSession = undefined
      Rimdian.currentCart = undefined
      Rimdian.currentPageview = undefined
      Rimdian.isReady = false
      Rimdian._onReady(Rimdian.config)
    }
  }
}

export default Rimdian
