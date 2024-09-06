import { faCalendar } from '@fortawesome/free-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { DatePicker, Radio } from 'antd'
import { useState } from 'react'
import dayjs from 'dayjs'
import CSS from 'utils/css'
import { SetURLSearchParams } from 'react-router-dom'

const { RangePicker } = DatePicker

type DateRangeSelectorProps = {
  preset: '7D' | '14D' | '30D' | 'custom'
  value?: DateRangeValue
  onChange?: (preset: DateRangePreset, value: DateRangeValue) => void
  small?: boolean
}

export type DateRangeValue = [string, string]
export type DateRangePreset = '7D' | '14D' | '30D' | 'custom'
export const presets = {
  '7D': (): DateRangeValue => [
    dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
    dayjs().format('YYYY-MM-DD')
  ],
  '14D': (): DateRangeValue => [
    dayjs().subtract(14, 'day').format('YYYY-MM-DD'),
    dayjs().format('YYYY-MM-DD')
  ],
  '30D': (): DateRangeValue => [
    dayjs().subtract(30, 'day').format('YYYY-MM-DD'),
    dayjs().format('YYYY-MM-DD')
  ],
  // default to 30 days if no value is provided for custom
  custom: (value?: DateRangeValue): DateRangeValue =>
    value || [dayjs().subtract(30, 'day').format('YYYY-MM-DD'), dayjs().format('YYYY-MM-DD')]
}

const DateRangeSelector = (props: DateRangeSelectorProps) => {
  const [preset, setPreset] = useState(props.preset)
  const [value, setValue] = useState<DateRangeValue>(
    presets[props.preset](props.value) as DateRangeValue
  )
  const [openCustom, setOpenCustom] = useState(false)
  const today = dayjs()

  const onChange = (preset: DateRangePreset, value: DateRangeValue) => {
    setPreset(preset)
    setValue(value)
    if (props.onChange) {
      props.onChange(preset, value)
    }
  }

  return (
    <div>
      <Radio.Group value={preset}>
        <Radio.Button
          value="custom"
          onClick={() => {
            setOpenCustom(true)
            setPreset('custom')
          }}
        >
          {preset === 'custom' && (
            <span className={CSS.padding_r_s}>
              {dayjs(value[0]).format('ll')} - {dayjs(value[1]).format('ll')}
            </span>
          )}
          <FontAwesomeIcon icon={faCalendar} />
          {openCustom === true && (
            <RangePicker
              open={true}
              onOpenChange={setOpenCustom}
              defaultValue={[dayjs(value[0]), dayjs(value[1])]}
              format="YYYY-MM-DD"
              placement="topRight"
              style={{ position: 'absolute', opacity: '0' }}
              onChange={(values) => {
                console.log('values', values)
                // onChange('custom', values as [dayjs.Dayjs, dayjs.Dayjs])
              }}
              disabledDate={(date: any) => {
                return date.endOf('day').isAfter(today.endOf('day'))
              }}
            />
          )}
        </Radio.Button>
        <Radio.Button
          value="7D"
          onClick={() => {
            onChange('7D', [
              dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
              today.format('YYYY-MM-DD')
            ])
          }}
        >
          7D
        </Radio.Button>
        <Radio.Button
          value="14D"
          onClick={() => {
            onChange('14D', [
              dayjs().subtract(14, 'day').format('YYYY-MM-DD'),
              today.format('YYYY-MM-DD')
            ])
          }}
        >
          14D
        </Radio.Button>
        <Radio.Button
          value="30D"
          onClick={() => {
            onChange('30D', [
              dayjs().subtract(30, 'day').format('YYYY-MM-DD'),
              today.format('YYYY-MM-DD')
            ])
          }}
        >
          30D
        </Radio.Button>
      </Radio.Group>
      {/* <Switch checkedChildren={<>vs</>} unCheckedChildren={<>vs</>} defaultChecked /> */}
    </div>
  )
}

export const dateRangeValuesFromSearchParams = (query: URLSearchParams): DateRangeValue => {
  const dateRange = query.get('date_range') as DateRangePreset
  if (dateRange) {
    if (dateRange === 'custom') {
      const dateFrom = query.get('date_from')
      const dateTo = query.get('date_to')
      if (dateFrom && dateTo) {
        return [dateFrom, dateTo]
      }
    }
    return presets[dateRange] ? presets[dateRange]() : presets['30D']()
  }

  return presets['30D']()
}

export const toStartOfDay = (date: string): string => {
  return dayjs(date).startOf('day').format('YYYY-MM-DD HH:mm:ss')
}

export const toEndOfDay = (date: string): string => {
  return dayjs(date).endOf('day').format('YYYY-MM-DD HH:mm:ss')
}

export const vsDateRangeValues = (dateFrom: string, dateTo: string): DateRangeValue => {
  // count the days between the two dates
  // return a date range of the previous period based on the number of days
  const diff = dayjs(dateTo).diff(dayjs(dateFrom), 'days') + 1
  return [
    dayjs(dateFrom).subtract(diff, 'days').format('YYYY-MM-DD'),
    dayjs(dateFrom).subtract(1, 'days').format('YYYY-MM-DD')
  ]
}

export const updateSearchParams = (
  searchParams: URLSearchParams,
  setSearchParams: SetURLSearchParams,
  preset: DateRangePreset,
  range: DateRangeValue
) => {
  if (preset !== 'custom') {
    searchParams.set('date_range', preset)
    searchParams.set('refresh_key', new Date().getTime().toString())
    searchParams.delete('date_from')
    searchParams.delete('date_to')
    setSearchParams(searchParams)
  } else {
    searchParams.set('date_range', preset)
    searchParams.set('date_from', range[0])
    searchParams.set('date_to', range[1])
    searchParams.set('refresh_key', new Date().getTime().toString())
    setSearchParams(searchParams)
  }
}
export default DateRangeSelector
