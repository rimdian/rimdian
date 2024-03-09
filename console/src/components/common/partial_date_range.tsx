import { faCalendar } from '@fortawesome/free-regular-svg-icons'
import { faRotateRight } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { Button, DatePicker, Dropdown } from 'antd'
import { useAccount } from 'components/login/context_account'
import { useState } from 'react'
import AntdCaretDown from 'utils/antd_caret_down'
import { useDateRangeCtx } from './context_date_range'
import dayjs from 'dayjs'
import CSS from 'utils/css'
const { RangePicker } = DatePicker

type DateRangeSelectorProps = {
  noCompare?: boolean
}

type Preset = {
  label: string
  key: string
  range: [dayjs.Dayjs, dayjs.Dayjs]
}

const DateRangeSelector = (props: DateRangeSelectorProps) => {
  const [visible, setVisible] = useState(false)
  const [vsVisible, setVsVisible] = useState(false)
  const [openCustom, setOpenCustom] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const accountCtx = useAccount()
  const dateRangeCtx = useDateRangeCtx()

  const timezone = accountCtx.account?.account.timezone || 'UTC'
  // export interface DateRangeCtxValue {
  //   dateRange: Dayjs.Dayjs[]
  //   dateRangePrevious: Dayjs.Dayjs[]
  //   setDateRange: (label: string, range: Dayjs.Dayjs[]) => void
  //   versus: string
  //   label: string
  // }

  const presets: Preset[] = [
    {
      key: 'last_30_days',
      range: [
        dayjs().tz(timezone).subtract(31, 'day').startOf('day'),
        dayjs().tz(timezone).subtract(1, 'day').endOf('day')
      ],
      label: 'Last 30 days'
    },
    {
      key: 'last_7_days',
      range: [
        dayjs().tz(timezone).subtract(8, 'day').startOf('day'),
        dayjs().tz(timezone).subtract(1, 'day').endOf('day')
      ],
      label: 'Last 7 days'
    },
    {
      key: 'last_14_days',
      range: [
        dayjs().tz(timezone).subtract(15, 'day').startOf('day'),
        dayjs().tz(timezone).subtract(1, 'day').endOf('day')
      ],
      label: 'Last 14 days'
    },
    {
      key: 'last_week',
      range: [
        dayjs().tz(timezone).startOf('week').subtract(1, 'week'),
        dayjs().tz(timezone).startOf('week').subtract(1, 'second')
      ],
      label: 'Last week'
    },
    {
      key: 'last_month',
      range: [
        dayjs().tz(timezone).startOf('month').subtract(1, 'month'),
        dayjs().tz(timezone).startOf('month').subtract(1, 'second')
      ],
      label: 'Last month'
    },
    {
      key: 'custom',
      range: [
        dayjs().tz(timezone).startOf('month').subtract(1, 'month'),
        dayjs().tz(timezone).startOf('month').subtract(1, 'second')
      ],
      label: 'Custom'
    }
  ]

  const handleClick = (e: any) => {
    const range = presets.find((x: any) => x.key === e.key)

    if (e.key === 'custom') {
      setOpenCustom(true)
    } else if (range) {
      dateRangeCtx.setDateRange(e.key, range.range)
    } else {
      console.error('date range not implemented', e.key)
    }

    setVisible(false)
  }

  const handleClickVs = (e: any) => {
    dateRangeCtx.setVersus(e.key)
    setVsVisible(false)
  }

  let label = 'custom'

  presets.forEach((x) => {
    if (x.key === dateRangeCtx.labelKey) {
      label = x.label
    }
  })

  if (dateRangeCtx.labelKey === 'custom') {
    label =
      dayjs(dateRangeCtx.dateRange[0]).format('ll') +
      ' ~ ' +
      dayjs(dateRangeCtx.dateRange[1]).format('ll')
  }

  return (
    <div>
      <Dropdown
        placement="bottomRight"
        trigger={['click']}
        menu={{
          onClick: handleClick,
          items: presets.map((x: any) => ({
            key: x.key,
            label: (dateRangeCtx.labelKey === x.key ? '• ' : '') + x.label
          }))
        }}
        open={visible}
        onOpenChange={setVisible}
      >
        <Button size="small">
          <FontAwesomeIcon icon={faCalendar} style={{ opacity: 0.7 }} />
          &nbsp;&nbsp;
          <span className={CSS.font_size_xs}>{label}</span>
          <AntdCaretDown />
          {openCustom === true && (
            <RangePicker
              open={true}
              onOpenChange={setOpenCustom}
              value={dateRangeCtx.dateRange}
              style={{ position: 'absolute', opacity: '0' }}
              onCalendarChange={(values, _formatString, info) => {
                if (info.range === 'end' && values) {
                  dateRangeCtx.setDateRange('custom', values as [dayjs.Dayjs, dayjs.Dayjs])
                  setVisible(false)
                  setOpenCustom(false)
                }
              }}
              disabledDate={(date: any) =>
                dayjs.tz(date, timezone).endOf('day').isAfter(dayjs().tz(timezone).endOf('day'))
              }
            />
          )}
        </Button>
      </Dropdown>
      {!props.noCompare && (
        <>
          <span className={CSS.padding_h_s}>vs</span>

          <Dropdown
            placement="bottomRight"
            trigger={['click']}
            open={vsVisible}
            onOpenChange={(value) => setVsVisible(value)}
            menu={{
              onClick: handleClickVs,
              items: [
                {
                  key: 'previous_period',
                  label: (dateRangeCtx.versus === 'previous_period' ? '• ' : '') + 'Previous period'
                },
                {
                  key: 'previous_year',
                  label: (dateRangeCtx.versus === 'previous_year' ? '• ' : '') + 'Previous year'
                }
              ]
            }}
          >
            <Button size="small">
              <span className={CSS.font_size_xs}>
                {dateRangeCtx.versus === 'previous_period' ? 'Previous period' : 'Previous year'}
              </span>
              <AntdCaretDown />
            </Button>
          </Dropdown>
        </>
      )}
      <Button
        size="small"
        type="default"
        shape="circle"
        className={CSS.margin_l_s}
        style={{ verticalAlign: 0 }}
        icon={<FontAwesomeIcon icon={faRotateRight} />}
        loading={isRefreshing}
        onClick={() => {
          dateRangeCtx.refresh()
          setIsRefreshing(true)
          window.setTimeout(() => {
            setIsRefreshing(false)
          }, 1000)
        }}
      />
    </div>
  )
}

export default DateRangeSelector
