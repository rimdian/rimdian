import dayjs from 'dayjs'
import { Query, TimeDimensionGranularity } from '@cubejs-client/core'
import { useEffect, useMemo, useRef, useState } from 'react'
import ReactECharts from 'echarts-for-react'
import { EChartsOption, SeriesOption } from 'echarts'
import { Alert, Divider, Spin, Button } from 'antd'
import { css } from '@emotion/css'
import CSS from 'utils/css'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faRefresh } from '@fortawesome/free-solid-svg-icons'
import { useRimdianCube } from 'components/workspace/context_cube'

export interface PartialHistogramProps {
  timezone: string
  measure: string
  timeDimension: string
  dateFrom: string
  dateTo: string
  granularity: TimeDimensionGranularity
}

const PartialHistogram = (props: PartialHistogramProps) => {
  //   const now = dayjs().format('YYYY-MM-DD')
  //   const last2days = dayjs().subtract(2, 'days').format('YYYY-MM-DD')
  //   const accountCtx = useAccount()
  const { cubeApi } = useRimdianCube()
  const refreshAtRef = useRef(0)
  const [refreshAt, setRefreshAt] = useState(dayjs().unix())
  const [loading, setLoading] = useState<boolean>(true)
  const [data, setData] = useState<any[]>([])
  const [categories, setCategories] = useState<any[]>([])
  const [error, setError] = useState<string | undefined>(undefined)

  const query: Query = useMemo(() => {
    return {
      measures: [props.measure],
      timeDimensions: [
        {
          dimension: props.timeDimension,
          granularity: props.granularity,
          dateRange: [props.dateFrom, props.dateTo]
        }
      ],
      timezone: props.timezone,
      limit: 1000
    }
  }, [
    props.dateFrom,
    props.dateTo,
    props.measure,
    props.timeDimension,
    props.timezone,
    props.granularity
  ])

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
        const colors: any = {}
        colors[props.measure] = '#69F0AE'

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
              name: s.title.replace('Data Imports ', ''),
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
          series[0].series.map((s) => dayjs.tz(s.x, props.timezone).format('dddd Do, HH:mm'))
        )

        setLoading(false)
        setError(undefined)
      })
      .catch((error) => {
        setError(error.toString())
        setLoading(false)
      })
  }, [query, refreshAt, cubeApi, props.timezone, props.measure])

  // console.log('categories', categories)
  // console.log('series', series)

  return (
    <span>
      {loading && (
        <div className={css(CSS.text_center, CSS.padding_v_l)} style={{ height: 100 }}>
          <Spin size="small" />
        </div>
      )}
      {!loading && (
        <span>
          <Button
            style={{ position: 'absolute', right: 32, opacity: 0.5, zIndex: 1000 }}
            size="small"
            type="text"
            onClick={() => setRefreshAt(dayjs().unix())}
          >
            <FontAwesomeIcon spin={loading} icon={faRefresh} />
          </Button>
          {error && (
            <div style={{ height: 100 }}>
              <Alert type="error" message={error} />
            </div>
          )}
          {!error && data.length === 0 && (
            <Divider style={{ padding: '32px 0' }} plain>
              No data
            </Divider>
          )}
          {!error && data.length > 0 && (
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
                  legend: {
                    show: false
                  },
                  // https://echarts.apache.org/en/option.html#xAxis
                  xAxis: [
                    {
                      type: 'category',
                      boundaryGap: true,
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
        </span>
      )}
    </span>
  )
}

export default PartialHistogram
