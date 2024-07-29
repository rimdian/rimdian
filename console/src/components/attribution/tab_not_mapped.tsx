import { Table } from 'antd'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { useSearchParams } from 'react-router-dom'
import { useAccount } from 'components/login/context_account'
import { Query, ResultSet, SqlData } from '@cubejs-client/core'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useDateRangeCtx } from 'components/common/context_date_range'
import { cloneDeep, set } from 'lodash'
import { ButtonExpand } from 'components/common/button_table_expand'
import { ButtonSQLExecuted, ExecutedSQL } from './button_sql_executed'
import CSS from 'utils/css'
import { css } from '@emotion/css'
import { Fullscreenable } from 'components/common/fullscreenable'
import { useRimdianCube } from 'components/workspace/context_cube'
import {
  AcquisitionAttributionRoleMeasure,
  DimensionDefinition,
  MeasureDefinition,
  RetentionAttributionRoleMeasure,
  generateDatabaseGraphForSchema,
  generateDimensionsMap,
  generateMeasuresMap
} from 'components/common/schema'
import {
  AttributionParams,
  DimensionsSelector,
  GenerateTableColumns,
  TableRow
} from './tab_sessions'

const TabAttributionSessionsNotMapped = () => {
  const accountCtx = useAccount()
  const workspaceCtx = useCurrentWorkspaceCtx()
  const dateRangeCtx = useDateRangeCtx()
  const [searchParams] = useSearchParams()
  const isMounted = useRef(true)
  const paramsHash = useRef<string | undefined>(undefined)
  const mutexObj = useMemo(() => {
    return {}
  }, [])
  const { cubeApi } = useRimdianCube()
  const [executedSQL, setExecutedSQL] = useState<ExecutedSQL[]>([])
  const [expandedRowKeys, setExpandedRowKeys] = useState<string[]>(['root'])
  const refreshKeyRef = useRef('')

  const graph = useMemo(() => {
    return generateDatabaseGraphForSchema('Session', workspaceCtx.cubeSchemasMap)
  }, [workspaceCtx.cubeSchemasMap])

  const dimensionsMap: Record<string, DimensionDefinition> = useMemo(() => {
    return generateDimensionsMap(graph, workspaceCtx.cubeSchemasMap)
  }, [workspaceCtx.cubeSchemasMap, graph])

  // dynamically add app measures to the definitions
  const measuresMap: Record<string, MeasureDefinition> = useMemo(() => {
    const result = generateMeasuresMap(graph, workspaceCtx.cubeSchemasMap)
    // add roles
    result['Session.acquisition_attribution_roles'] = AcquisitionAttributionRoleMeasure
    result['Session.retention_attribution_roles'] = RetentionAttributionRoleMeasure
    return result
  }, [workspaceCtx.cubeSchemasMap, graph])

  // hardcode measures for now...
  const defaultMeasures: MeasureDefinition[] = useMemo(() => {
    return [
      // measuresMap['Session.unique_users'],
      measuresMap['Session.count'],
      // measuresMap['Session.bounce_rate'],
      // measuresMap['Session.avg_pageviews_count'],
      // measuresMap['Session.avg_duration'],
      measuresMap['Session.acquisition_orders_contributions'],
      measuresMap['Session.acquisition_linear_amount_attributed'],
      measuresMap['Session.acquisition_attribution_roles'],
      measuresMap['Session.retention_orders_contributions'],
      measuresMap['Session.retention_linear_amount_attributed'],
      measuresMap['Session.retention_attribution_roles'],
      measuresMap['Order.subtotal_sum'],
      measuresMap['Order.count']
    ]
  }, [measuresMap])

  const params: AttributionParams = useMemo(() => {
    // console.log('defaultMeasures', defaultMeasures)
    return {
      sortKey: searchParams.get('sortKey') || 'Session.count',
      sortOrder: searchParams.get('sortOrder') || 'desc',
      dimension1: searchParams.get('dimension1') || 'Session.utm_source',
      dimension2: searchParams.get('dimension2') || 'Session.utm_medium',
      dimension3: searchParams.get('dimension3') || 'Session.utm_campaign',
      measures: searchParams.get('measures') || defaultMeasures.map((field) => field.key).join(','),
      segment: searchParams.get('segment') || '_all',
      date_from: dateRangeCtx.dateRange[0].format('YYYY-MM-DD'),
      date_to: dateRangeCtx.dateRange[1].format('YYYY-MM-DD'),
      vs_date_from: dateRangeCtx.dateRangePrevious[0].format('YYYY-MM-DD'),
      vs_date_to: dateRangeCtx.dateRangePrevious[1].format('YYYY-MM-DD'),
      refresh_key: searchParams.get('refresh_key') || ''
    }
  }, [searchParams, dateRangeCtx, defaultMeasures])

  const measures: string[] = useMemo(() => {
    if (!measuresMap) return []

    const result: string[] = []
    params.measures.split(',').forEach((fieldKey) => {
      if (!measuresMap[fieldKey]) return
      result.push(...(measuresMap[fieldKey].dependsOnMeasures || [fieldKey]))
    })
    return result
  }, [params, measuresMap])

  // console.log('measures', measures)
  const baseQuery: Query = useMemo(() => {
    let renewQuery = false

    if (params.refresh_key !== refreshKeyRef.current) {
      renewQuery = true
      refreshKeyRef.current = params.refresh_key
    }

    return {
      measures: measures,
      filters: [
        {
          member: 'Session.channel_id',
          operator: 'equals',
          values: ['not-mapped']
        }
      ],
      timeDimensions: [
        {
          dimension: 'Session.created_at_trunc',
          granularity: null as any,
          compareDateRange: [
            [
              dateRangeCtx.dateRange[0].format('YYYY-MM-DD'),
              dateRangeCtx.dateRange[1].format('YYYY-MM-DD')
            ],
            [
              dateRangeCtx.dateRangePrevious[0].format('YYYY-MM-DD'),
              dateRangeCtx.dateRangePrevious[1].format('YYYY-MM-DD')
            ]
          ]
        }
      ],
      timezone: accountCtx.account?.account.timezone,
      order: {
        [params.sortKey]: params.sortOrder === 'asc' ? 'asc' : 'desc'
      },
      limit: 1000,
      renewQuery: renewQuery
    }
  }, [measures, params.sortKey, params.sortOrder, dateRangeCtx, accountCtx, params.refresh_key])

  const defaultTableData: TableRow[] = useMemo(() => {
    return [
      {
        key: 'root',
        loading: true,
        query: { ...baseQuery },
        result: null
      }
    ]
  }, [baseQuery])

  const [tableData, setTableData] = useState<TableRow[]>([...defaultTableData])

  const fetchChildren = useCallback(
    (row?: TableRow) => {
      // console.log('fetchChildren', row)

      let parentKey = row ? row.key : 'root'
      let childrenQuery = cloneDeep(baseQuery)
      const depth = parentKey.split('.').length
      // console.log('depth', depth)

      // add dimensions & filters to query
      if (row) {
        // apply parent filters & dimensions
        childrenQuery.dimensions = row.query?.dimensions ? [...row.query.dimensions] : []
        childrenQuery.filters = row.query?.filters ? [...row.query.filters] : []

        // add new dimension & filter

        // add 1st dimension
        if (depth === 1) {
          childrenQuery.dimensions.push(params.dimension1)
        }

        // add 2nd dimension, and filter with first dimension value
        if (depth === 2) {
          childrenQuery.dimensions.push(params.dimension2)
          childrenQuery.filters.push({
            dimension: params.dimension1,
            operator: 'equals',
            // values should be strings
            values: ['' + row.result.currentPeriod[params.dimension1]]
          })
        }

        // add 3rd dimension, and filter with second dimension value
        if (depth === 3) {
          childrenQuery.dimensions.push(params.dimension3)
          childrenQuery.filters.push({
            dimension: params.dimension2,
            operator: 'equals',
            // values should be strings
            values: ['' + row.result.currentPeriod[params.dimension2]]
          })
        }
      }

      // console.log('childrenQuery', childrenQuery)

      Promise.all([
        cubeApi.sql(childrenQuery, { mutexObj: mutexObj, mutexKey: 'sql_' + parentKey }),
        cubeApi.load(childrenQuery, { mutexObj: mutexObj, mutexKey: 'load_' + parentKey })
      ])
        .then(([sqlQuery, resultSet]: any[]) => {
          // console.log('sqlQuery', sqlQuery)
          // console.log('resultSet', resultSet)

          const currentPeriodSQL: SqlData = sqlQuery[0].sqlQuery.sql
          const previousPeriodSQL: SqlData = sqlQuery[1].sqlQuery.sql

          if (!isMounted.current) {
            // abort if component has been unmounted
            console.log('not mounted')
            return
          }

          setExecutedSQL((prev) => [
            ...prev,
            {
              name: 'current_period_' + parentKey,
              sql: currentPeriodSQL.sql[0],
              args: currentPeriodSQL.sql[1]
            },
            {
              name: 'previous_period_' + parentKey,
              sql: previousPeriodSQL.sql[0],
              args: previousPeriodSQL.sql[1]
            }
          ])

          const [currentPeriod, previousPeriod]: [ResultSet, ResultSet] = resultSet.decompose()

          if (!row) {
            const root: TableRow = {
              key: parentKey,
              loading: false,
              query: childrenQuery,
              result: {
                currentPeriod: currentPeriod.rawData()[0],
                previousPeriod: previousPeriod.rawData()[0]
              },
              children: [
                // inject a fake row to show the loading state
                {
                  key: parentKey + '[0]',
                  loading: true,
                  result: null
                }
              ]
            }
            setTableData([root])

            // load children just after the root is loaded
            fetchChildren(root)
            return
          }

          const children: TableRow[] = []
          const currentData = currentPeriod.rawData()
          const previousData = previousPeriod.rawData()

          // console.log('currentData', currentData)
          // console.log('previousData', previousData)

          // loop over current period data
          currentData.forEach((currentRow, i) => {
            const child: TableRow = {
              key: 'to-compute',
              loading: false,
              query: { ...childrenQuery },
              dimensionValues: childrenQuery.dimensions
                ? childrenQuery.dimensions.map((dim) => currentRow[dim])
                : [],
              result: {
                currentPeriod: GenerateDefaultResult(childrenQuery),
                previousPeriod: GenerateDefaultResult(childrenQuery)
              }
            }

            // add current period data
            Object.keys(currentRow).forEach((key) => {
              child.result.currentPeriod[key] = currentRow[key]
            })

            if (depth < 3) {
              child.children = [
                // inject a fake row to show the loading state
                {
                  key: parentKey + '[' + i + '].children[0]',
                  loading: true,
                  query: { ...childrenQuery, dimensions: [params.dimension2] },
                  result: null
                }
              ]
            }
            children.push(child)
          })

          // add previous period data or create missing rows
          previousData.forEach((previousRow, i) => {
            const dimensionValues = childrenQuery.dimensions
              ? childrenQuery.dimensions.map((dim) => previousRow[dim])
              : []

            // check if we have a corresponding existing row
            let existingRow = children.find((child) => {
              // check if dimensionValues are equal
              return child.dimensionValues?.every((dimValue) => {
                return dimensionValues.includes(dimValue)
              })
            })

            if (existingRow) {
              Object.keys(previousRow).forEach((key) => {
                ;(existingRow as TableRow).result.previousPeriod[key] = previousRow[key]
              })
            }

            if (!existingRow) {
              existingRow = {
                key: 'to-compute',
                loading: false,
                query: { ...childrenQuery },
                dimensionValues: dimensionValues,
                result: {
                  currentPeriod: GenerateDefaultResult(childrenQuery),
                  previousPeriod: GenerateDefaultResult(childrenQuery)
                }
              }
              // add previous period data
              Object.keys(previousRow).forEach((key) => {
                ;(existingRow as TableRow).result.previousPeriod[key] = previousRow[key]
              })
              children.push(existingRow)
            }
          })

          // compute children keys
          children.forEach((child, i) => {
            child.key = parentKey + '.children[' + i + ']'
          })

          // inject children into the row
          setTableData((prev) => {
            const newTableData = [...prev]

            const parentPath = parentKey.replace('root', '[0]')
            set(newTableData, parentPath + '.children', children)
            // console.log('newTableData', newTableData)
            return newTableData
          })
        })
        .catch((error) => console.error(error))
    },
    [mutexObj, baseQuery, cubeApi, params.dimension1, params.dimension2, params.dimension3]
  )

  // load the first time or when the params change
  useEffect(() => {
    // first load
    if (!paramsHash.current) {
      // console.log('first load params is', params)
      paramsHash.current = JSON.stringify(params)
      fetchChildren()
      return
    }

    // if params changed, fetch new data
    if (JSON.stringify(params) !== paramsHash.current) {
      // console.log('update params is', params)
      paramsHash.current = JSON.stringify(params)
      // reset table data
      setExpandedRowKeys(['root'])
      setTableData([...defaultTableData])
      fetchChildren()
    }
  }, [params, defaultTableData, fetchChildren])

  // unmounting component only, dont put things inside this effect
  useEffect(() => {
    isMounted.current = true
    return () => {
      isMounted.current = false
    }
  }, [])

  const GenerateDefaultResult = (query: Query): any => {
    const result: any = {}
    query.measures?.forEach((measure) => {
      result[measure] = 0
    })
    query.dimensions?.forEach((dimension) => {
      result[dimension] = null
    })
    return result
  }

  // console.log('executedSQL', executedSQL)

  const selectedMeasures = useMemo(() => {
    // retrieve measures from params in the map
    const result: MeasureDefinition[] = []
    params.measures.split(',').forEach((measure) => {
      if (measuresMap[measure]) result.push(measuresMap[measure])
    })
    return result
  }, [params.measures, measuresMap])

  // params.measures.split(',').map((measure) => measuresMap[measure])
  const tableColumns = GenerateTableColumns(selectedMeasures, dimensionsMap, workspaceCtx.workspace)

  // compute total table width
  let totalX = 0
  tableColumns.forEach((col) => {
    totalX += col.width
  })

  // console.log('params', params)

  return (
    <>
      <div className={css(CSS.leftRight)}>
        <DimensionsSelector
          dimension1={params.dimension1}
          dimension2={params.dimension2}
          dimension3={params.dimension3}
          dimensionsMap={dimensionsMap}
        />
      </div>

      <Fullscreenable>
        <Table
          scroll={{ x: totalX, y: 700 }}
          pagination={false}
          dataSource={tableData}
          size="middle"
          rowClassName={(record) => css([record.key === 'root' && CSS.tableTotals])}
          expandable={{
            expandedRowKeys: expandedRowKeys,
            expandRowByClick: true,
            indentSize: 0, // we will control indentation ourselves
            onExpand: (expanded, row) => {
              if (!expanded) {
                // remove key from expandedRowKeys
                setExpandedRowKeys((prev) => prev.filter((key) => key !== row.key))
              } else {
                // add key to expandedRowKeys
                setExpandedRowKeys((prev) => [...prev, row.key])
              }

              // only fetch children if row is expanded and has children
              if (expanded && row.children && row.children[0]?.loading) {
                fetchChildren(row)
              }
            },
            expandIcon: ({ expanded, onExpand, record }) => {
              // hide the expand icon on root row or if there are no children
              if (record.key === 'root' || !record.children) return ''
              return (
                <ButtonExpand
                  depth={record.key.split('.').length}
                  expanded={expanded}
                  onClick={(e) => onExpand(record, e)}
                />
              )
            }
          }}
          // onChange={onTableChange}
          rowKey="key"
          columns={tableColumns}
        />
      </Fullscreenable>
      <div className={CSS.text_right + ' ' + CSS.margin_t_l}>
        <ButtonSQLExecuted queries={executedSQL} />
      </div>
    </>
  )
}

export default TabAttributionSessionsNotMapped
