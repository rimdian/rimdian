import dayjs from 'dayjs'
import { useAccount } from 'components/login/context_account'
import { Query } from '@cubejs-client/core'
import { useEffect, useMemo, useRef, useState } from 'react'
import ReactECharts from 'echarts-for-react'
import { EChartsOption, SeriesOption } from 'echarts'
import { Button, Spin } from 'antd'
import Block from 'components/common/block'
import { css } from '@emotion/css'
import CSS from 'utils/css'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faRefresh } from '@fortawesome/free-solid-svg-icons'
import { useRimdianCube } from 'components/workspace/context_cube'

const GraphActivity = () => {
  const now = dayjs().format('YYYY-MM-DD')
  const last2days = dayjs().subtract(2, 'days').format('YYYY-MM-DD')
  const accountCtx = useAccount()
  const { cubeApi } = useRimdianCube()
  const refreshAtRef = useRef(0)
  const [refreshAt, setRefreshAt] = useState(dayjs().unix())
  const [loading, setLoading] = useState<boolean>(true)
  const [data, setData] = useState<any[]>([])
  const [categories, setCategories] = useState<any[]>([])
  const [error, setError] = useState<string | undefined>(undefined)

  const query: Query = useMemo(() => {
    return {
      measures: [
        'Data_log.successful_count',
        'Data_log.not_done_count',
        'Data_log.error_non_retryable_count',
        'Data_log.error_retryable_count'
      ],
      timeDimensions: [
        {
          dimension: 'Data_log.event_at_trunc',
          granularity: 'hour',
          dateRange: [last2days, now]
        }
      ],
      timezone: accountCtx.account?.account.timezone,
      // order: {
      //   ['Data_log.event_at_trunc']: 'asc'
      // },
      limit: 1000
    }
  }, [last2days, now, accountCtx.account?.account.timezone])

  useEffect(() => {
    if (refreshAtRef.current === refreshAt) {
      return
    }

    refreshAtRef.current = refreshAt

    setLoading(true)

    cubeApi
      .load(query)
      .then((resultSet) => {
        // colors are defined by the series keys
        const colors: any = {
          'Data_log.successful_count': '#00E676',
          'Data_log.not_done_count': '#00B0FF',
          'Data_log.error_non_retryable_count': '#FF1744',
          'Data_log.error_retryable_count': '#FFEA00'
        }

        const seriesNames = resultSet.seriesNames()

        if (seriesNames.length === 0) {
          setCategories([])
          setData([])
          setLoading(false)
          setError(undefined)
          return
        }

        setData(
          resultSet.series().map((s, i) => {
            return {
              name: s.title.replace('Data logs ', ''),
              type: 'bar',
              stack: 'all',
              color: colors[seriesNames[i].key] as string,
              emphasis: { focus: 'series' }, // blur non focused series
              data: s.series.map((v) => v.value)
            } as SeriesOption
          })
        )

        const series = resultSet.series()

        setCategories(
          series[0].series.map((s) =>
            dayjs
              .tz(s.x, 'UTC')
              .tz(accountCtx.account?.account.timezone || 'UTC')
              .format('dddd Do, HH:mm')
          )
        )

        setLoading(false)
        setError(undefined)
      })
      .catch((error) => {
        setError(error.toString())
      })
  }, [query, refreshAt, cubeApi, accountCtx.account?.account.timezone])

  // console.log('categories', categories)
  // console.log('series', series)

  return (
    <Block
      small
      title="Logs over the last 48h"
      extra={
        <Button size="small" onClick={() => setRefreshAt(dayjs().unix())}>
          <FontAwesomeIcon spin={loading} icon={faRefresh} />
        </Button>
      }
    >
      {loading && (
        <div className={css(CSS.text_center, CSS.padding_v_l)} style={{ height: 100 }}>
          <Spin size="small" />
        </div>
      )}
      {!loading && error}
      {!loading && (
        <ReactECharts
          option={
            {
              animationDuration: 150,
              renderer: 'svg',
              grid: {
                // top: 0,
                left: 0,
                right: 0,
                bottom: '1px', // required to see the bottom line
                // height: 80,
                // show: false,
                //   borderWidth: 1,
                containLabel: false
              },
              tooltip: {
                trigger: 'axis',
                axisPointer: {
                  type: 'shadow'
                }
              },
              legend: {},
              // https://echarts.apache.org/en/option.html#xAxis
              xAxis: [
                {
                  type: 'category',
                  boundaryGap: false,
                  data: categories,
                  // axisLabel: { show: false },
                  axisTick: { show: false },
                  axisLine: {
                    show: true,
                    lineStyle: {
                      color: 'rgba(0,0,0,0.1)',
                      width: 1,
                      type: 'solid'
                    }
                  }
                  // axisTick: { show: false }
                }
              ],
              yAxis: [
                {
                  type: 'value',
                  axisLabel: { show: false },
                  axisLine: { show: false },
                  axisTick: { show: false },
                  splitLine: { show: false }
                }
              ],
              series: data
            } as EChartsOption
          }
          className="echart"
          style={{ height: 100, cursor: 'default !important' }}
        />
      )}
    </Block>
  )
}

export default GraphActivity
