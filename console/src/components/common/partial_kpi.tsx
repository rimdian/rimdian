import { ReactNode, useEffect, useMemo, useState } from 'react'
import { Popover, Spin, Tooltip } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import FormatCurrency from 'utils/format_currency'
import FormatDuration from 'utils/format_duration'
import ReactECharts from 'echarts-for-react'
import { SeriesOption } from 'echarts'
import FormatGrowth from 'utils/format_growth'
import FormatPercent from 'utils/format_percent'
import FormatNumber from 'utils/format_number'
import dayjs from 'dayjs'
import CSS, { colorLabel, colorPrimary } from 'utils/css'
import { css } from '@emotion/css'
import { faCircleInfo, faDatabase } from '@fortawesome/free-solid-svg-icons'
import { TimeDimension } from '@cubejs-client/core'
import { useRimdianCube } from 'components/workspace/context_cube'

export interface ExecutedSQL {
  name: string
  sql: string
  args: any[]
}

export type KPIProps = {
  title: string
  valueType: 'number' | 'percent' | 'currency' | 'duration'
  currency?: string
  goodIsBad?: boolean // when a growing kpi is bad
  measure: string
  dimensions?: string[]
  filters?: any[]
  timeDimension: string
  color: string
  workspaceId: string
  timezone: string
  refreshAt: number
  dateFrom: string
  dateTo: string
  dateFromPrevious: string
  dateToPrevious: string
  tooltip?: ReactNode | string
}

export const KPI = (props: KPIProps) => {
  const { cubeApi } = useRimdianCube()
  const [isLoading, setIsLoading] = useState(true)
  const [value, setValue] = useState<number | undefined>(undefined)
  const [previousValue, setPreviousValue] = useState<number | undefined>(undefined)
  const [error, setError] = useState<string | undefined>(undefined)
  // graph
  const [isLoadingGraph, setIsLoadingGraph] = useState(true)
  const [valueGraph, setValueGraph] = useState<any[] | undefined>(undefined)
  const [previousValueGraph, setPreviousValueGraph] = useState<any[] | undefined>(undefined)
  const [executedSQL, setExecutedSQL] = useState<ExecutedSQL[]>([])

  const totalQuery = useMemo(() => {
    return {
      measures: [props.measure],
      dimensions: props.dimensions,
      filters: props.filters,
      timeDimensions: [
        {
          dimension: props.timeDimension,
          granularity: undefined, // total count
          compareDateRange: [
            [props.dateFrom, props.dateTo],
            [props.dateFromPrevious, props.dateToPrevious]
          ]
        }
      ],
      timezone: props.timezone
    }
  }, [
    props.measure,
    props.dimensions,
    props.filters,
    props.timeDimension,
    props.timezone,
    props.dateFrom,
    props.dateTo,
    props.dateFromPrevious,
    props.dateToPrevious
  ])

  const graphQuery = useMemo(() => {
    return {
      measures: [props.measure],
      dimensions: props.dimensions,
      filters: props.filters,
      timeDimensions: [
        {
          dimension: props.timeDimension,
          granularity: 'day', // by day
          compareDateRange: [
            [props.dateFrom, props.dateTo],
            [props.dateFromPrevious, props.dateToPrevious]
          ]
        }
      ] as TimeDimension[],
      timezone: props.timezone
    }
  }, [
    props.measure,
    props.dimensions,
    props.filters,
    props.timeDimension,
    props.timezone,
    props.dateFrom,
    props.dateTo,
    props.dateFromPrevious,
    props.dateToPrevious
  ])

  // fetch
  useEffect(() => {
    // console.log('KPI: fetch', totalQuery, graphQuery)
    setIsLoading(true)
    setIsLoadingGraph(true)

    Promise.all([
      cubeApi.sql(totalQuery),
      cubeApi.load(totalQuery),
      cubeApi.sql(graphQuery),
      cubeApi.load(graphQuery)
      // cubeApi.sql(totalQuery, { mutexObj: window.MutexCubeJS, mutexKey: 'sql' }),
      // cubeApi.load(totalQuery, { mutexObj: window.MutexCubeJS, mutexKey: 'load' }),
      // cubeApi.sql(graphQuery, { mutexObj: window.MutexCubeJS, mutexKey: 'sql' }),
      // cubeApi.load(graphQuery, { mutexObj: window.MutexCubeJS, mutexKey: 'load' })
    ])
      .then(([sqlQuery, resultSet, sqlGraphQuery, graphResultSet]: any[]) => {
        const [currentTotal, previousTotal] = resultSet.decompose()
        setValue(currentTotal.series()[0]?.series[0].value)
        setPreviousValue(previousTotal.series()[0]?.series[0].value)
        setIsLoading(false)
        setError(undefined)

        const [currentDays, previousDays] = graphResultSet.decompose()
        setValueGraph(currentDays.series()[0]?.series || [])
        setPreviousValueGraph(previousDays.series()[0]?.series || [])
        setIsLoadingGraph(false)

        setExecutedSQL([
          {
            name: 'Total',
            sql: sqlQuery[0].sqlQuery.sql.sql[0],
            args: sqlQuery[0].sqlQuery.sql.sql[1]
          },
          {
            name: 'Graph',
            sql: sqlGraphQuery[0].sqlQuery.sql.sql[0],
            args: sqlGraphQuery[0].sqlQuery.sql.sql[1]
          }
        ])
      })
      .catch((error) => {
        setError(error.toString())
      })
  }, [totalQuery, graphQuery, props.refreshAt, cubeApi])

  let title = props.title

  if (error) {
    title = `${title} - ${error}`
  }

  const kpiProps = {
    valueIsLoading: isLoading,
    title: title,
    tooltip: props.tooltip,
    value: value,
    valueType: props.valueType,
    currency: props.currency,
    goodIsBad: props.goodIsBad,
    previousValue: previousValue,
    graphIsLoading: isLoadingGraph,
    graphData: valueGraph,
    graphDataPrevious: previousValueGraph,
    executedSQL: executedSQL
  } as RenderKPIProps

  return <RenderKPI {...kpiProps} />
}

export type RenderKPIProps = {
  valueIsLoading?: boolean
  title: ReactNode | string
  tooltip?: ReactNode | string
  valueType: 'number' | 'percent' | 'currency' | 'duration' | 'relativeDate'
  currency?: string
  value: number
  previousValue?: number
  goodIsBad?: boolean // when a growing kpi is bad
  graphIsLoading?: boolean
  graphData?: any[]
  graphDataPrevious?: any[]
  executedSQL?: ExecutedSQL[]
}

export const kpiCss = {
  self: css(
    {
      overflow: 'hidden',
      // '& :not(:last-child)': {
      //   borderRight: '1px solid ' + borderColorSecondary
      // },
      '& .echart > div': {
        cursor: 'default'
      }
    },
    CSS.padding_a_m
  ),
  title: css(
    {
      fontSize: 12,
      fontWeight: 500,
      color: colorLabel,
      textTransform: 'capitalize'
    },
    CSS.margin_b_m
  ),
  value: css({
    lineHeight: '24px',
    minHeight: '24px',
    fontSize: 18,
    fontWeight: 600,
    color: '#011832'
  }),

  graphLoader: css({
    height: 30,
    backgroundColor: '#E8EAF6',
    borderBottom: '1px solid ' + colorPrimary
  })
}

export const RenderKPI = (props: RenderKPIProps) => {
  let value: any = props.value

  if (!props.valueIsLoading) {
    if (props.valueType === 'number') {
      value = FormatNumber(value)
    }

    if (props.valueType === 'percent') {
      value = FormatPercent(value)
    }

    if (props.valueType === 'currency') {
      value = FormatCurrency(value, props.currency as string)
    }

    if (props.valueType === 'duration') {
      value = FormatDuration(value)
    }

    if (props.valueType === 'relativeDate') {
      if (!value) value = '-'
      else value = dayjs.unix(props.value).fromNow()
    }
  }

  const categories: string[] = []
  const lineData: number[] = []
  const previousLineData: number[] = []
  const series = [] as SeriesOption[]

  if (props.graphDataPrevious) {
    props.graphDataPrevious.forEach((item: any, i: number) => {
      previousLineData.push(item.value as number)
    })
    series.push({
      silent: true, // disable hover and cursor:pointer
      type: 'line',
      symbol: 'none',
      cursor: 'default',
      z: 2,
      lineStyle: { type: 'solid', color: '#3D5AFE', width: 1, opacity: 0.3 },
      data: previousLineData
    } as SeriesOption)
  }

  if (props.graphData) {
    props.graphData.forEach((item: any, i: number) => {
      lineData.push(item.value as number)
      categories.push(i + '')
    })
    series.push({
      silent: true, // disable hover and cursor:pointer
      type: 'line',
      symbol: 'none',
      cursor: 'default',
      z: 3,
      // smooth: true,
      // areaStyle: {
      //   opacity: 0.5,
      //   color: new graphic.LinearGradient(0, 0, 0, 1, [
      //     {
      //       offset: 0,
      //       color: '#3D5AFE'
      //     },
      //     {
      //       offset: 1,
      //       color: '#3D5AFE'
      //     }
      //   ])
      // },
      //   areaStyle: {
      //     color: '#00E5FF',
      //     opacity: 0.05
      //   },
      lineStyle: { type: 'solid', color: '#3D5AFE', width: 1 },
      data: lineData
    } as SeriesOption)
  }

  return (
    <div className={kpiCss.self}>
      <div className={kpiCss.title}>
        {props.tooltip && (
          <Tooltip title={props.tooltip}>
            {props.title}{' '}
            <FontAwesomeIcon
              icon={faCircleInfo}
              className={css([CSS.margin_l_xs, CSS.opacity_30])}
            />
          </Tooltip>
        )}
        {!props.tooltip && props.title}

        {props.executedSQL && (
          <span style={{ float: 'right' }}>
            <Popover
              title="SQL queries"
              placement="bottom"
              content={
                <div style={{ width: 500 }}>
                  {props.executedSQL.map((q) => {
                    return (
                      <div key={q.name} style={{ marginBottom: 24 }}>
                        <b>{q.name}</b>
                        <div className={CSS.font_size_xs}>
                          <code>{q.sql}</code>
                        </div>
                        {q.args.length > 0 && (
                          <div className={CSS.margin_t_s}>
                            <b>Arguments:</b>
                            {q.args.map((arg, index) => {
                              return <div key={index}>{arg}</div>
                            })}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>
              }
            >
              {props.executedSQL.length > 0 && (
                <span className={css([CSS.margin_l_xs, CSS.opacity_30])}>
                  <FontAwesomeIcon icon={faDatabase} />
                  {/* <span
                    style={{
                      backgroundColor: colorLabel,
                      color: 'white',
                      fontSize: 8,
                      fontWeight: 'bold',
                      padding: '1px 2px',
                      borderRadius: 2,
                      position: 'relative',
                      top: -1
                    }}
                  >
                    SQL
                  </span> */}
                </span>
              )}
            </Popover>
          </span>
        )}
      </div>
      <span className={kpiCss.value}>
        {props.valueIsLoading && <Spin size="small" />}
        {!props.valueIsLoading && value}
      </span>
      {!props.valueIsLoading && FormatGrowth(props.value, props.previousValue, props.goodIsBad)}
      <div className={CSS.margin_t_m}>
        {props.graphIsLoading && (
          <Spin size="small">
            <div className={kpiCss.graphLoader}></div>
          </Spin>
        )}
        {!props.graphIsLoading && props.graphData && (
          <ReactECharts
            option={{
              animationDuration: 150,
              renderer: 'svg',
              grid: {
                top: 0,
                left: 0,
                right: 0,
                bottom: '1px', // required to see the bottom line
                show: false,
                //   borderWidth: 1,
                containLabel: false
              },
              xAxis: [
                {
                  type: 'category',
                  boundaryGap: false,
                  data: categories,
                  axisLabel: { show: false },
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
              series: series
            }}
            className="echart"
            style={{ height: 30, cursor: 'default !important' }}
          />
        )}
      </div>
    </div>
  )
}
