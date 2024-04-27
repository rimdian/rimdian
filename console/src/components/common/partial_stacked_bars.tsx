import dayjs from 'dayjs'
import { Query, TimeDimensionGranularity } from '@cubejs-client/core'
import { CubeContext } from '@cubejs-client/react'
import { useContext, useEffect, useMemo, useRef, useState } from 'react'
import ReactECharts from 'echarts-for-react'
import { EChartsOption, SeriesOption } from 'echarts'
import { Alert, Button, Divider, Spin } from 'antd'
import { css } from '@emotion/css'
import CSS from 'utils/css'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faRefresh } from '@fortawesome/free-solid-svg-icons'

export interface PartialStackedBarsSerie {
  name: string
  measure: string
  timeDimension: string
  color?: string
}
export interface PartialStackedBarsProps {
  timezone: string
  series: PartialStackedBarsSerie[]
  dateFrom: string
  dateTo: string
  granularity: TimeDimensionGranularity
}

const PartialStackedBars = (props: PartialStackedBarsProps) => {
  //   const now = dayjs().format('YYYY-MM-DD')
  //   const last2days = dayjs().subtract(2, 'days').format('YYYY-MM-DD')
  //   const accountCtx = useAccount()
  const { cubeApi } = useContext(CubeContext)
  const refreshAtRef = useRef(0)
  const [refreshAt, setRefreshAt] = useState(dayjs().unix())
  const [loading, setLoading] = useState<boolean>(true)
  const [data, setData] = useState<any[]>([])
  const [categories, setCategories] = useState<any[]>([])
  const [error, setError] = useState<string | undefined>(undefined)

  const queries: Query[] = useMemo(() => {
    return props.series.map((serie) => {
      return {
        measures: [serie.measure],
        timeDimensions: [
          {
            dimension: serie.timeDimension,
            granularity: props.granularity,
            dateRange: [props.dateFrom, props.dateTo]
          }
        ],
        timezone: props.timezone,
        limit: 1000
      }
    })
  }, [props.dateFrom, props.dateTo, props.series, props.timezone, props.granularity])

  useEffect(() => {
    if (refreshAtRef.current === refreshAt) {
      return
    }

    refreshAtRef.current = refreshAt

    setLoading(true)

    cubeApi
      .load(queries)
      .then((resultSet) => {
        // console.log('resultSet', resultSet)

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
            const serie = props.series.find((serie) => serie.measure === s.key)
            return {
              name: serie?.name,
              type: 'bar',
              stack: 'all',
              color: serie?.color,
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
  }, [queries, refreshAt, cubeApi, props.timezone, props.series])

  // console.log('categories', categories)
  // console.log('series', series)
  return (
    <div>
      {loading && (
        <div className={css(CSS.text_center, CSS.padding_v_l)} style={{ height: 100 }}>
          <Spin size="small" />
        </div>
      )}
      {!loading && (
        <>
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
            <Divider style={{ padding: '24px 0' }} plain>
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
        </>
      )}
    </div>
  )
}

export default PartialStackedBars
