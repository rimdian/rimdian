import { ReactNode, useEffect, useMemo, useRef, useState } from 'react'
import { Popover, Table, Tooltip } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import FormatCurrency from 'utils/format_currency'
import FormatDuration from 'utils/format_duration'
import FormatPercent from 'utils/format_percent'
import FormatNumber from 'utils/format_number'
import CSS from 'utils/css'
import { css, CSSInterpolation } from '@emotion/css'
import { faDatabase } from '@fortawesome/free-solid-svg-icons'
import { useRimdianCube } from 'components/workspace/context_cube'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { CubeSchemaDimension, CubeSchemaMeasure } from 'interfaces'
import Block from './block'
import { Query, ResultSet } from '@cubejs-client/core'
import { SizeType } from 'antd/lib/config-provider/SizeContext'

export interface ExecutedSQL {
  name: string
  sql: string
  args: any[]
}

export type TableWidgetProps = {
  workspaceCtx: CurrentWorkspaceCtxValue
  title?: string
  size?: SizeType
  measures: TableWidgetMeasure[]
  dimensions: TableWidgetDimension[]
  order: any
  limit: number
  filters?: any[]
  timeDimension?: string
  dateFrom?: string
  dateTo?: string
  refreshKey: string
  classNames?: CSSInterpolation[]
}

export type TableWidgetMeasure = {
  measure: string
  title?: string // override title
}

export type TableWidgetDimension = {
  dimension: string
  render?: (x: any) => ReactNode
  title?: string // override title
}

export const TableWidget = (props: TableWidgetProps) => {
  const { cubeApi } = useRimdianCube()
  const [loading, setLoading] = useState(true)
  const [data, setData] = useState<ResultSet | undefined>(undefined)
  const [error, setError] = useState<string | undefined>(undefined)
  const refreshKeyRef = useRef<string | undefined>(undefined)
  // graph
  const [executedSQL, setExecutedSQL] = useState<ExecutedSQL[]>([])

  const query = useMemo(() => {
    const data: Query = {
      measures: props.measures.map((m) => m.measure),
      dimensions: props.dimensions.map((d) => d.dimension),
      filters: props.filters || [],
      timezone: props.workspaceCtx.accountCtx.account?.account.timezone || 'UTC',
      renewQuery: refreshKeyRef.current !== props.refreshKey,
      order: props.order,
      limit: props.limit
    }

    if (props.timeDimension && props.dateFrom && props.dateTo) {
      data['timeDimensions'] = [
        {
          dimension: props.timeDimension,
          dateRange: [props.dateFrom, props.dateTo]
        }
      ]
    }

    return data
  }, [
    props.measures,
    props.dimensions,
    props.timeDimension,
    props.dateFrom,
    props.dateTo,
    props.filters,
    props.workspaceCtx.accountCtx.account?.account.timezone,
    props.order,
    props.limit,
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

    setLoading(true)

    Promise.all([cubeApi.sql(query), cubeApi.load(query)])
      .then(([sqlQuery, resultSet]: any[]) => {
        resultSet = resultSet.decompose()[0] as ResultSet
        setData(resultSet)
        setLoading(false)
        setError(undefined)

        setExecutedSQL([
          {
            name: props.title || 'Query',
            sql: sqlQuery.sqlQuery.sql.sql[0],
            args: sqlQuery.sqlQuery.sql.sql[1]
          }
        ])
      })
      .catch((error) => {
        setError(error.toString())
        setLoading(false)
      })
  }, [query, cubeApi, props.refreshKey, props.title])

  const columns = useMemo(() => {
    const cols: any[] = props.dimensions.map((d) => {
      const [schema, dimension] = d.dimension.split('.')
      const definition: CubeSchemaDimension =
        props.workspaceCtx.cubeSchemasMap[schema].dimensions[dimension]
      return {
        title: d.title || definition.title,
        key: d.dimension,
        render: (x: any) => {
          if (d.render) {
            return d.render(x)
          }
          return x[d.dimension]
        }
      }
    })

    props.measures.forEach((m) => {
      // find measure definition
      // split measure into schema and measure
      const [schema, measure] = m.measure.split('.')
      const definition: CubeSchemaMeasure =
        props.workspaceCtx.cubeSchemasMap[schema].measures[measure]

      const isSorted = props.order[m.measure] === 'desc'

      cols.push({
        title: <Tooltip title={definition.description}>{m.title || definition.title}</Tooltip>,
        key: m.measure,
        sorter: isSorted ? (a: any, b: any) => a[m.measure] - b[m.measure] : undefined,
        sortDirections: isSorted ? ['descend'] : undefined,
        render: (x: any) => {
          let value: any = x[m.measure]

          if (definition.meta?.rimdian_format === 'percentage') {
            value = FormatPercent(value)
          } else if (definition.meta?.rimdian_format === 'currency') {
            value = FormatCurrency(value, props.workspaceCtx.workspace.currency)
          } else if (definition.meta?.rimdian_format === 'duration') {
            value = FormatDuration(value)
          } else if (definition.type === 'number') {
            value = FormatNumber(value)
          }

          return value
        }
      })
    })

    return cols
  }, [
    props.measures,
    props.dimensions,
    props.order,
    props.workspaceCtx.cubeSchemasMap,
    props.workspaceCtx.workspace.currency
  ])

  return (
    <Block
      title={props.title}
      small
      classNames={props.classNames || []}
      extra={
        <>
          {executedSQL && (
            <Popover
              title="SQL queries"
              placement="left"
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
                <span className={css([CSS.opacity_30, CSS.font_size_xs])}>
                  <FontAwesomeIcon icon={faDatabase} />
                </span>
              )}
            </Popover>
          )}
        </>
      }
    >
      <Table
        rowKey="Session.channel_origin_id"
        size={props.size}
        loading={loading}
        locale={{
          emptyText: error ? <div className={CSS.text_red}>{error}</div> : 'No data'
        }}
        dataSource={data ? data.tablePivot() : []}
        pagination={false}
        columns={columns}
      />
    </Block>
  )
}
