import { useAccount } from 'components/login/context_account'
import { createContext, useContext, useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import dayjs from 'dayjs'

// IMPORTANT: the date range context is placed before the routes
// to make sure we preserve the date range when navigating and the refresh key

export interface DateRangeCtxValue {
  dateRange: [dayjs.Dayjs, dayjs.Dayjs]
  dateRangePrevious: [dayjs.Dayjs, dayjs.Dayjs]
  setDateRange: (label: string, range: [dayjs.Dayjs, dayjs.Dayjs]) => void
  setVersus: (label: string) => void
  versus: string
  labelKey: string
  refresh: () => void
  refreshAt: number
}

const DateRangeContext = createContext<DateRangeCtxValue | null>(null)

export function useDateRangeCtx(): DateRangeCtxValue {
  const contextValue = useContext(DateRangeContext)
  if (!contextValue) {
    throw new Error('Missing DateRangeContextProvider in its parent.')
  }
  return contextValue
}

type DateRangeProviderProps = {
  children?: React.ReactNode
}

export const DateRangeProvider = (props: DateRangeProviderProps) => {
  const accountCtx = useAccount()
  const timezone = accountCtx.account?.account.timezone || 'UTC'
  const [searchParams, setSearchParams] = useSearchParams()
  const refreshKeyParam = searchParams.get('refresh_key')
  const refreshKeyRef = useRef(refreshKeyParam ? Number(refreshKeyParam) : dayjs().unix())
  const [refreshAt, setRefreshAt] = useState(refreshKeyRef.current)

  // update the refresh key when the search params change
  useEffect(() => {
    // console.log('effect', refreshKeyParam, refreshKeyRef.current)
    if (refreshKeyParam !== null && Number(refreshKeyParam) !== refreshKeyRef.current) {
      refreshKeyRef.current = Number(refreshKeyParam)
      setRefreshAt(refreshKeyRef.current)
    }
  }, [refreshKeyParam])

  const defaultVersus = searchParams.get('versus') || 'previous_period'
  const defaultLabelKey = searchParams.get('labelKey') || 'last_30_days'

  const defaultDateFrom = searchParams.get('date_from')
    ? dayjs.tz(searchParams.get('date_from'), timezone)
    : dayjs().tz(timezone).subtract(30, 'days').startOf('day')

  const defaultDateTo = searchParams.get('date_to')
    ? dayjs(searchParams.get('date_to'), timezone)
    : dayjs().tz(timezone).subtract(1, 'day').endOf('day')

  const [versus, setVersusValue] = useState(defaultVersus)
  const [labelKey, setLabelKey] = useState(defaultLabelKey)
  const [dateRange, setDateRangeValue] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([
    defaultDateFrom,
    defaultDateTo
  ])

  const setDateRange = (label: string, range: [dayjs.Dayjs, dayjs.Dayjs]) => {
    setDateRangeValue(range)
    setLabelKey(label)
    setRefreshAt(dayjs().unix())
  }

  const setVersus = (value: string) => {
    setVersusValue(value)
  }

  // set refresh_key to the searchs params
  const refresh = () => {
    const params: any = {}
    searchParams.forEach((value, key) => {
      params[key] = value
    })
    params.refresh_key = dayjs().unix()
    setSearchParams({ ...params })
  }

  // console.log('state', this.state);

  // clone dates to prevent mutating the original dates
  const dateFrom = dayjs(dateRange[0])
  const dateTo = dayjs(dateRange[1])

  let dateFromPrevious: dayjs.Dayjs = dayjs(dateFrom)
  let dateToPrevious: dayjs.Dayjs = dayjs(dateTo)

  if (versus === 'previous_period') {
    // add one day to the diff to avoid days overlaping between the to ranges
    const diff = dateTo.diff(dateFrom, 'days') + 1
    // console.log('diff is', diff);

    dateToPrevious = dateToPrevious.subtract(diff, 'days')
    dateFromPrevious = dateFromPrevious.subtract(diff, 'days')
  }

  if (versus === 'previous_year') {
    dateToPrevious = dateToPrevious.subtract(1, 'years')
    dateFromPrevious = dateFromPrevious.subtract(1, 'years')
  }

  return (
    <DateRangeContext.Provider
      value={{
        dateRange,
        dateRangePrevious: [dateFromPrevious, dateToPrevious],
        setDateRange,
        setVersus,
        versus,
        labelKey,
        refresh,
        refreshAt
      }}
    >
      {props.children}
    </DateRangeContext.Provider>
  )
}
