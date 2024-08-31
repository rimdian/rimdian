import { useEffect, useMemo, useRef, useState } from 'react'
import { Popover, Spin, Tooltip } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import ReactECharts from 'echarts-for-react'
import { SeriesOption } from 'echarts'
import dayjs from 'dayjs'
import CSS, { colorLabel, colorPrimary } from 'utils/css'
import { css } from '@emotion/css'
import { faDatabase, faTriangleExclamation } from '@fortawesome/free-solid-svg-icons'
import { ResultSet, Series, TimeDimension } from '@cubejs-client/core'
import { useRimdianCube } from 'components/workspace/context_cube'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'

interface ExecutedSQL {
  name: string
  sql: string
  args: any[]
}

export type LineWidgetMeasure = {
  name: string
  measure: string
  color: string
  filters?: any[]
  dateFrom: string
  dateTo: string
}

export type LinesWidgetProps = {
  workspaceCtx: CurrentWorkspaceCtxValue
  schema: string
  line1: LineWidgetMeasure
  line2: LineWidgetMeasure
  timeDimension: string
  refreshKey: string
  title: string
}

type SeriesItem = {
  value: number
  x: string
}

const kpiCss = {
  self: css(
    {
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
      color: colorLabel
      //   textTransform: 'capitalize'
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

export const LinesWidget = (props: LinesWidgetProps) => {
  const { cubeApi } = useRimdianCube()
  const [error, setError] = useState<string | undefined>(undefined)
  const refreshKeyRef = useRef<string | undefined>(undefined)
  const [isLoading, setIsLoading] = useState(true)
  const [line1Graph, setLine1Graph] = useState<Series<SeriesItem> | undefined>(undefined)
  const [line2Graph, setLine2Graph] = useState<Series<SeriesItem> | undefined>(undefined)
  const [executedSQL, setExecutedSQL] = useState<ExecutedSQL[]>([])

  const line1Query = useMemo(() => {
    return {
      measures: [props.schema + '.' + props.line1.measure],
      filters: props.line1.filters || [],
      timeDimensions: [
        {
          dimension: props.timeDimension,
          granularity: 'day', // by day
          dateRange: [props.line1.dateFrom, props.line1.dateTo]
        }
      ] as TimeDimension[],
      timezone: props.workspaceCtx.accountCtx.account?.account.timezone || 'UTC',
      renewQuery: refreshKeyRef.current !== props.refreshKey
    }
  }, [
    props.line1,
    props.timeDimension,
    props.workspaceCtx,
    props.schema,
    props.refreshKey,
    refreshKeyRef
  ])

  const line2Query = useMemo(() => {
    return {
      measures: [props.schema + '.' + props.line2.measure],
      filters: props.line2.filters || [],
      timeDimensions: [
        {
          dimension: props.timeDimension,
          granularity: 'day', // by day
          dateRange: [props.line2.dateFrom, props.line2.dateTo]
        }
      ] as TimeDimension[],
      timezone: props.workspaceCtx.accountCtx.account?.account.timezone || 'UTC',
      renewQuery: refreshKeyRef.current !== props.refreshKey
    }
  }, [
    props.line2,
    props.timeDimension,
    props.workspaceCtx,
    props.refreshKey,
    refreshKeyRef,
    props.schema
  ])

  // fetch
  useEffect(() => {
    if (props.refreshKey === refreshKeyRef.current) {
      return
    } else {
      refreshKeyRef.current = props.refreshKey
    }

    setIsLoading(true)

    Promise.all([
      cubeApi.sql(line1Query),
      cubeApi.load(line1Query),
      cubeApi.sql(line2Query),
      cubeApi.load(line2Query)
    ])
      .then(([line1SQL, line1Result, line2SQL, line2Result]: any[]) => {
        line1Result = line1Result as ResultSet
        line2Result = line2Result as ResultSet

        // console.log('line 1 series', line1Result.series()[0])
        // console.log('line 2 series', line2Result.series())

        setLine1Graph(line1Result.series()[0])
        setLine2Graph(line2Result.series()[0])
        setError(undefined)
        setExecutedSQL([
          {
            name: props.line1.name,
            sql: line1SQL.sqlQuery.sql.sql[0],
            args: line1SQL.sqlQuery.sql.sql[1]
          },
          {
            name: props.line2.name,
            sql: line2SQL.sqlQuery.sql.sql[0],
            args: line2SQL.sqlQuery.sql.sql[1]
          }
        ])
      })
      .catch((error) => {
        setError(error.toString())
      })
      .finally(() => {
        setIsLoading(false)
      })
  }, [line1Query, line2Query, cubeApi, props.refreshKey, props.line1.name, props.line2.name])

  const series: SeriesOption[] = useMemo(() => {
    return [
      {
        name: props.line2.name,
        silent: true, // disable hover and cursor:pointer
        type: 'line',
        symbol: 'none',
        cursor: 'default',
        z: 1,
        lineStyle: { type: 'solid', color: props.line2.color || '#3D5AFE', width: 1 },
        data: line2Graph?.series.map((item: any) => item.value as number) || []
      },
      {
        name: props.line1.name,
        silent: true, // disable hover and cursor:pointer
        type: 'line',
        symbol: 'none',
        cursor: 'default',
        z: 1,
        lineStyle: { type: 'solid', color: props.line1.color || '#3D5AFE', width: 1 },
        data: line1Graph?.series.map((item: any) => item.value as number) || []
      }
    ] as SeriesOption[]
  }, [
    line1Graph,
    line2Graph,
    props.line1.color,
    props.line1.name,
    props.line2.color,
    props.line2.name
  ])

  const categories = useMemo(() => {
    if (!line1Graph) return []
    return line1Graph?.series.map((item: any) => dayjs(item.x).format('MMM D')) || []
  }, [line1Graph])

  return (
    <div className={kpiCss.self}>
      <div className={kpiCss.title}>
        {props.title}

        {executedSQL && (
          <span style={{ float: 'right' }}>
            <Popover
              title="SQL queries"
              placement="bottom"
              content={
                <div style={{ width: 500 }}>
                  {executedSQL.map((q) => {
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
              {executedSQL.length > 0 && (
                <span className={css([CSS.margin_l_xs, CSS.opacity_30])}>
                  <FontAwesomeIcon icon={faDatabase} />
                </span>
              )}
            </Popover>
          </span>
        )}
      </div>

      <div className={CSS.margin_t_m}>
        {!isLoading && error && (
          <Tooltip title={error}>
            <FontAwesomeIcon className={CSS.text_amber} icon={faTriangleExclamation} />
          </Tooltip>
        )}

        {isLoading && (
          <Spin size="small">
            <div className={kpiCss.graphLoader}></div>
          </Spin>
        )}

        {!isLoading && (
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
              tooltip: {
                trigger: 'axis',
                axisPointer: {
                  type: 'shadow'
                }
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
