import { useEffect, useMemo, useRef, useState } from 'react'
import { Popover, Spin, Tooltip } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import FormatCurrency from 'utils/format_currency'
import FormatDuration from 'utils/format_duration'
import FormatPercent from 'utils/format_percent'
import FormatNumber from 'utils/format_number'
import CSS, { colorLabel } from 'utils/css'
import { css } from '@emotion/css'
import { faCircleInfo, faDatabase, faTriangleExclamation } from '@fortawesome/free-solid-svg-icons'
import { useRimdianCube } from 'components/workspace/context_cube'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { CubeSchemaMeasure } from 'interfaces'

export interface ExecutedSQL {
  name: string
  sql: string
  args: any[]
}

export type MetricWidgetProps = {
  workspaceCtx: CurrentWorkspaceCtxValue
  schema: string
  measure: string
  filters?: any[]
  refreshKey: string
  title?: string
}

export const MetricWidget = (props: MetricWidgetProps) => {
  const { cubeApi } = useRimdianCube()
  const [isLoading, setIsLoading] = useState(true)
  const [value, setValue] = useState<number | undefined>(undefined)
  const [error, setError] = useState<string | undefined>(undefined)
  const refreshKeyRef = useRef<string | undefined>(undefined)
  // graph
  const [executedSQL, setExecutedSQL] = useState<ExecutedSQL[]>([])

  const definition: CubeSchemaMeasure = useMemo(() => {
    return props.workspaceCtx.cubeSchemasMap[props.schema].measures[props.measure]
  }, [props.schema, props.measure, props.workspaceCtx.cubeSchemasMap])

  const query = useMemo(() => {
    const data = {
      measures: [props.schema + '.' + props.measure],
      filters: props.filters || [],
      // dimensions: ['User_segment.segment_id'],
      timezone: props.workspaceCtx.accountCtx.account?.account.timezone || 'UTC',
      renewQuery: refreshKeyRef.current !== props.refreshKey
    }

    return data
  }, [
    props.schema,
    props.measure,
    props.filters,
    props.workspaceCtx.accountCtx.account?.account.timezone,
    props.refreshKey,
    refreshKeyRef
  ])

  // fetch
  useEffect(() => {
    if (props.refreshKey === refreshKeyRef.current) {
      return
    } else {
      refreshKeyRef.current = props.refreshKey
    }

    setIsLoading(true)

    Promise.all([cubeApi.sql(query), cubeApi.load(query)])
      .then(([sqlQuery, resultSet]: any[]) => {
        const [currentTotal] = resultSet.decompose()
        setValue(currentTotal.series()[0]?.series[0].value)
        setIsLoading(false)
        setError(undefined)

        setExecutedSQL([
          {
            name: 'Metric',
            sql: sqlQuery.sqlQuery.sql.sql[0],
            args: sqlQuery.sqlQuery.sql.sql[1]
          }
        ])
      })
      .catch((error) => {
        setError(error.toString())
        setIsLoading(false)
      })
  }, [query, cubeApi, props.refreshKey])

  const kpiProps = {
    loading: isLoading,
    value: value,
    query: query,
    title: props.title,
    error: error,
    definition: definition,
    currency: props.workspaceCtx.workspace.currency,
    executedSQL: executedSQL
  } as RenderMetricWidgetProps

  return <RenderMetricWidget {...kpiProps} />
}

export type RenderMetricWidgetProps = {
  definition: CubeSchemaMeasure
  currency: string
  query: any
  value: number
  title?: string
  error?: string
  loading: boolean
  executedSQL?: ExecutedSQL[]
}

export const kpiCss = {
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
  })
}

export const RenderMetricWidget = (props: RenderMetricWidgetProps) => {
  let value: any = props.value

  if (!props.loading && !props.error) {
    if (props.definition.meta?.rimdian_format === 'percentage') {
      value = FormatPercent(props.value)
    } else if (props.definition.meta?.rimdian_format === 'currency') {
      value = FormatCurrency(props.value, props.currency)
    } else if (props.definition.meta?.rimdian_format === 'duration') {
      value = FormatDuration(props.value)
    } else if (props.definition.type === 'number') {
      value = FormatNumber(props.value)
    }
  }

  return (
    <div className={kpiCss.self}>
      <div className={kpiCss.title}>
        <Tooltip title={props.definition.description}>
          {props.title || props.definition.title}{' '}
          <FontAwesomeIcon icon={faCircleInfo} className={css([CSS.margin_l_xs, CSS.opacity_30])} />
        </Tooltip>

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
                </span>
              )}
            </Popover>
          </span>
        )}
      </div>
      <span className={kpiCss.value}>
        {props.loading && <Spin size="small" />}
        {!props.loading && !props.error && <>{value}</>}
        {!props.loading && props.error && (
          <>
            <Tooltip title={props.error + ' ' + JSON.stringify(props.query)}>
              <FontAwesomeIcon className={CSS.text_amber} icon={faTriangleExclamation} />
            </Tooltip>
          </>
        )}
      </span>
    </div>
  )
}
