import { Tooltip, Tag } from 'antd'
import dayjs from 'dayjs'
import CSS from 'utils/css'

const PartialCustomColumn = (value: any, timezone: string): any => {
  const valueType = toString.call(value)

  if (valueType === '[object Boolean]') return value.toString()
  // object with {$date, $timezone}
  if (valueType === '[object Object]' && value['$date']) {
    // todo
    // return (
    //   <Tooltip
    //     title={
    //       Moment(value['$date'])
    //         .tz(value['$timezone'] || 'UTC')
    //         .format('lll') +
    //       ' (' +
    //       (value['$timezone'] || 'UTC') +
    //       ')'
    //     }
    //   >
    //     {Moment(value['$date']).fromNow()}
    //   </Tooltip>
    // )
  }
  if (valueType === '[object String]') {
    // if is a date
    if (
      /-?([1-9][0-9]{3,}|0[0-9]{3})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T(([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?|(24:00:00(\.0+)?))(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))?/.test(
        value
      )
    ) {
      return (
        <Tooltip
          title={
            dayjs(value)
              .tz(timezone || 'UTC')
              .format('lll') +
            ' (' +
            (timezone || 'UTC') +
            ')'
          }
        >
          {dayjs(value).fromNow()}
        </Tooltip>
      )
    }

    // is url
    if (/https?:\/\/([\da-zA-Z.-]+)\.([a-zA-Z.]{2,6})([/\w.-]*)*\/?/.test(value)) {
      return (
        <a href={value} target="_blank" rel="noreferrer noopener">
          {value}
        </a>
      )
    }

    // default return value
    return value
  }

  if (valueType === '[object Number]') {
    // todo
    //   return Numeral(value).format('0,0[.]00')
  }

  if (valueType === '[object Array]') {
    return value.map((item: any) => (
      <span key={item}>
        <Tag className={CSS.margin_r_s}>{item}</Tag>
      </span>
    ))
  }

  value.toString()
}

export default PartialCustomColumn
