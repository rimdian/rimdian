import {
  Tag,
  Table,
  Tooltip,
  Button,
  Spin,
  Select,
  Modal,
  // Badge,
  // Row,
  // Col,
  // Radio,
  Checkbox,
  Radio,
  RadioChangeEvent
} from 'antd'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { useSearchParams } from 'react-router-dom'
import { useAccount } from 'components/login/context_account'
import { Filter, Query, ResultSet, SqlData } from '@cubejs-client/core'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  PostviewsMeasuresMapDefinition,
  MeasureDefinition,
  DimensionsMapDefinition,
  DimensionDefinition
} from './postviews_definitions'
import { capitalize, cloneDeep, forEach, map, set, upperFirst } from 'lodash'
import FormatNumber from 'utils/format_number'
import FormatPercent from 'utils/format_percent'
import FormatCurrency from 'utils/format_currency'
import FormatDuration from 'utils/format_duration'
import { ButtonExpand } from 'components/common/button_table_expand'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faQuestionCircle } from '@fortawesome/free-regular-svg-icons'
import { Workspace } from 'interfaces'
import { ButtonSQLExecuted, ExecutedSQL } from './button_sql_executed'
import TableTag from 'components/common/partial_table_tag'
import CSS from 'utils/css'
import { css } from '@emotion/css'
import { Graph } from '@datastructures-js/graph'
import { Fullscreenable } from 'components/common/fullscreenable'
import { useRimdianCube } from 'components/workspace/context_cube'
import {
  dateRangeValuesFromSearchParams,
  toEndOfDay,
  toStartOfDay,
  vsDateRangeValues
} from 'components/common/partial_date_range'

interface AttributionParams {
  sortKey: string
  sortOrder: string
  dimension1: string
  dimension2: string
  dimension3: string
  conversions_filter: string
  measures: string
  date_from: string
  date_to: string
  vs_date_from: string
  vs_date_to: string
  refresh_key: string
}

// hardcode measures for now...
const defaultMeasures: MeasureDefinition[] = [
  PostviewsMeasuresMapDefinition['Postview.unique_users'],
  PostviewsMeasuresMapDefinition['Postview.count'],
  PostviewsMeasuresMapDefinition['Postview.orders_contributions'],
  PostviewsMeasuresMapDefinition['Postview.distinct_orders'],
  PostviewsMeasuresMapDefinition['Order.subtotal_sum'],
  PostviewsMeasuresMapDefinition['Postview.attribution_roles']
]

const dimensionsAvailable: Record<string, DimensionDefinition> = {
  'Postview.channel_group_id': DimensionsMapDefinition['Postview.channel_group_id'],
  'Postview.channel_id': DimensionsMapDefinition['Postview.channel_id'],
  'Postview.channel_origin_id': DimensionsMapDefinition['Postview.channel_origin_id'],
  'Postview.utm_source': DimensionsMapDefinition['Postview.utm_source'],
  'Postview.utm_medium': DimensionsMapDefinition['Postview.utm_medium'],
  'Postview.utm_campaign': DimensionsMapDefinition['Postview.utm_campaign'],
  'Postview.utm_content': DimensionsMapDefinition['Postview.utm_content'],
  'Postview.utm_term': DimensionsMapDefinition['Postview.utm_term'],
  'Postview.is_first_conversion': DimensionsMapDefinition['Postview.is_first_conversion'],
  'Postview.role': DimensionsMapDefinition['Postview.role']
}

interface TableRow {
  key: string // contains path of parents too
  loading: boolean
  query?: Query // query is used by children to extract dimensions etc... and build drill down queries
  result: any
  dimensionValues?: any[]
  children?: TableRow[]
}

const TabAttributionPostviews = () => {
  const accountCtx = useAccount()
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [searchParams, setSearchParams] = useSearchParams()
  const isMounted = useRef(true)
  const paramsHash = useRef<string | undefined>(undefined)
  const mutexObj = useMemo(() => {
    return {}
  }, [])
  const { cubeApi } = useRimdianCube()
  const [executedSQL, setExecutedSQL] = useState<ExecutedSQL[]>([])
  const [expandedRowKeys, setExpandedRowKeys] = useState<string[]>(['root'])
  const refreshKeyRef = useRef('')

  const params: AttributionParams = useMemo(() => {
    const [dateFrom, dateTo] = dateRangeValuesFromSearchParams(searchParams)
    const [vsDateFrom, vsDateTo] = vsDateRangeValues(dateFrom, dateTo)
    return {
      sortKey: searchParams.get('sortKey') || 'Postview.count',
      sortOrder: searchParams.get('sortOrder') || 'desc',
      dimension1: searchParams.get('dimension1') || 'Postview.utm_campaign',
      dimension2: searchParams.get('dimension2') || 'Postview.utm_content',
      dimension3: searchParams.get('dimension3') || 'Postview.channel_id',
      measures: searchParams.get('measures') || defaultMeasures.map((field) => field.key).join(','),
      conversions_filter: searchParams.get('conversions_filter') || 'all',
      date_from: dateFrom,
      date_to: dateTo,
      vs_date_from: vsDateFrom,
      vs_date_to: vsDateTo,
      refresh_key: searchParams.get('refresh_key') || 'default'
    }
  }, [searchParams])

  //   console.log(workspaceCtx.cubeSchemasMap)

  const graph = useMemo(() => {
    // compute graph network of tables
    const g = new Graph<string, boolean>()
    g.addVertex('user', true)
    g.addVertex('session', true)
    g.addVertex('postview', true)
    g.addVertex('order', true)
    g.addEdge('user', 'session')
    g.addEdge('user', 'postview')
    g.addEdge('user', 'order')
    g.addEdge('order', 'session')
    g.addEdge('order', 'postview')

    workspaceCtx.workspace.installed_apps.forEach((app) => {
      // add vertices first
      app.app_tables?.forEach((table) => {
        g.addVertex(table.name, true)
      })
      // add edges
      forEach(app.cube_schemas, (schema, cubeName) => {
        if (schema.joins) {
          forEach(schema.joins, (_join, tableName) => {
            g.addEdge(cubeName, tableName)
          })
        }
      })
    })
    return g
  }, [workspaceCtx.workspace.installed_apps])

  // const tableLinks = useMemo(() => {
  //   return GenerateTablesLinks(workspaceCtx.workspace.installed_apps)
  // }, [workspaceCtx.workspace.installed_apps])

  // dynamically add app measures to the definitions
  const dimensionsMap: Record<string, DimensionDefinition> = useMemo(() => {
    const result = cloneDeep(dimensionsAvailable)

    // extract app dimensions
    workspaceCtx.workspace.installed_apps.forEach((app) => {
      app.extra_columns?.forEach((extendedTable) => {
        extendedTable.columns.forEach((column) => {
          // only accept strings and booleans for now
          if (column.type === 'varchar' || column.type === 'boolean') {
            const k = `${capitalize(extendedTable.kind)}.${column.name}`

            result[k] = {
              key: k,
              title: app.name + ' - ' + column.name,
              tooltip: column.description,
              category: 'app',
              dimension: k
            } as DimensionDefinition
          }
        })
      })

      app.app_tables?.forEach((table) => {
        // check if table is linked to postview table
        if (!graph.hasEdge(table.name, 'postview')) {
          return
        }

        table.columns.forEach((column) => {
          // only accept strings and booleans for now
          if (column.type === 'varchar' || column.type === 'boolean') {
            const tableNameCube = upperFirst(table.name)
            const k = `${tableNameCube}.${column.name}`
            result[k] = {
              key: k,
              title: column.name,
              tooltip: column.description,
              category: 'app',
              dimension: k
            } as DimensionDefinition
          }
        })
      })
    })

    return result
  }, [workspaceCtx.workspace.installed_apps, graph])

  // dynamically add app measures to the definitions
  const measuresMap: Record<string, MeasureDefinition> = useMemo(() => {
    const result = cloneDeep(PostviewsMeasuresMapDefinition)

    // extract app measures
    workspaceCtx.workspace.installed_apps.forEach((app) => {
      app.app_tables?.forEach((table) => {
        // check if table is linked to postview table
        if (!graph.hasEdge(table.name, 'postview')) {
          return
        }

        table.columns.forEach((column) => {
          // number columns have SUM and AVG measures
          if (column.type === 'number') {
            const tableNameCube = upperFirst(table.name)
            const sumKey = `${tableNameCube}.${column.name}_sum`
            result[sumKey] = {
              key: sumKey,
              title: `Sum of ${column.name}`,
              tooltip: `Sum of ${column.name}`,
              category: 'app',
              measures: [sumKey]
            } as MeasureDefinition

            const avgKey = `${tableNameCube}.${column.name}_avg`
            result[avgKey] = {
              key: avgKey,
              title: `Avg. ${column.name}`,
              tooltip: `Average of ${column.name}`,
              category: 'app',
              measures: [avgKey]
            } as MeasureDefinition
          }
        })
      })
    })

    return result
  }, [workspaceCtx.workspace.installed_apps, graph])

  const measures: string[] = useMemo(() => {
    const result: string[] = []
    params.measures.split(',').forEach((fieldKey) => {
      result.push(...measuresMap[fieldKey].measures)
    })
    return result
  }, [params.measures, measuresMap])

  const baseQuery: Query = useMemo(() => {
    const filters: Filter[] = []

    // add conversion filter
    if (params.conversions_filter === 'acquisition') {
      filters.push({
        dimension: 'Order.is_first_conversion',
        operator: 'equals',
        values: ['1']
      })
    }

    if (params.conversions_filter === 'repeat') {
      filters.push({
        dimension: 'Order.is_first_conversion',
        operator: 'equals',
        values: ['0']
      })
    }

    let renewQuery = false

    if (params.refresh_key !== refreshKeyRef.current) {
      renewQuery = true
      refreshKeyRef.current = params.refresh_key
    }

    return {
      measures: measures,
      filters: filters,
      timeDimensions: [
        {
          dimension: 'Postview.created_at_trunc',
          granularity: null as any,
          compareDateRange: [
            [toStartOfDay(params.date_from), toEndOfDay(params.date_to)],
            [toStartOfDay(params.vs_date_from), toEndOfDay(params.vs_date_to)]
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
  }, [
    measures,
    params.sortKey,
    params.sortOrder,
    params.conversions_filter,
    params.date_from,
    params.date_to,
    params.vs_date_from,
    params.vs_date_to,
    accountCtx,
    params.refresh_key
  ])

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

  const selectedMeasures = params.measures.split(',').map((measure) => measuresMap[measure])
  const tableColumns = GenerateTableColumns(selectedMeasures, dimensionsMap, workspaceCtx.workspace)

  // compute total table width
  let totalX = 0
  tableColumns.forEach((col) => {
    totalX += col.width
  })

  // console.log('tableData', tableData)

  return (
    <>
      <div className={css(CSS.leftRight)}>
        <DimensionsSelector
          dimension1={params.dimension1}
          dimension2={params.dimension2}
          dimension3={params.dimension3}
          dimensionsMap={dimensionsMap}
        />
        <span className={CSS.padding_h_xs}></span>
        <MeasuresSelector measures={params.measures.split(',')} fieldsMap={measuresMap} />
        <div className={CSS.topSeparator}></div>
        <Radio.Group
          onChange={(e: RadioChangeEvent) => {
            const allParams: any = {}
            searchParams.forEach((value, key: string) => {
              allParams[key] = value
            })
            setSearchParams({ ...allParams, conversions_filter: e.target.value })
          }}
          value={params.conversions_filter}
          defaultValue={'all'}
        >
          <Radio.Button value={'all'}>All conversions</Radio.Button>
          <Radio.Button value={'acquisition'}>Acquisition</Radio.Button>
          <Radio.Button value={'repeat'}>Repeat</Radio.Button>
        </Radio.Group>
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

const GenerateTableColumns = (
  measures: MeasureDefinition[],
  dimensionsMap: Record<string, DimensionDefinition>,
  workspace: Workspace
): any[] => {
  // extract categories from fields
  const categoriesMap: any = {}

  measures.forEach((field) => {
    categoriesMap[field.category] = field.category
  })

  const categories: any[] = [
    // first category is 1st dimension
    {
      key: 'dimension1',
      title: '',
      className: CSS.borderRight.solid1,
      width: 250,
      fixed: 'left',
      render: (row: TableRow) => {
        if (row.loading) return <Spin size="small" />
        if (row.key === 'root' || !row.dimensionValues || !row.query?.dimensions) {
          return ''
        }

        // only print the last dimension value
        const lastDimensionKey = row.query?.dimensions[row.query.dimensions.length - 1]
        const lastDimensionValue = row.dimensionValues[row.dimensionValues.length - 1]
        const dimension = dimensionsMap[lastDimensionKey]

        // find channel group
        if (dimension.key === 'Session.channel_group_id') {
          const channelGroup = workspace.channel_groups.find(
            (group) => group.id === lastDimensionValue
          )
          return channelGroup ? (
            <Tag color={channelGroup.color}>{channelGroup.name}</Tag>
          ) : (
            lastDimensionValue
          )
        }

        // find channel
        if (dimension.key === 'Session.channel_id') {
          const channel = workspace.channels.find((ch) => ch.id === lastDimensionValue)
          return channel ? channel.name : lastDimensionValue
        }

        if (dimension.type === 'boolean') {
          if (lastDimensionValue === null) {
            return 'null'
          }
          if (
            lastDimensionValue === '1' ||
            lastDimensionValue === true ||
            lastDimensionValue === 'true' ||
            lastDimensionValue === 1
          ) {
            return 'Yes'
          }
          if (
            lastDimensionValue === '0' ||
            lastDimensionValue === false ||
            lastDimensionValue === 'false' ||
            lastDimensionValue === 0
          ) {
            return 'No'
          }
          return lastDimensionValue
        }

        return lastDimensionValue
      }
    }
  ]

  Object.keys(categoriesMap).forEach((key, i, cats) => {
    const isLastCategory = i === cats.length - 1
    // first column is always the 1st dimension name
    const columns: any[] = []
    let totalColumnsWidth = 0

    // find fields in this category
    measures.forEach((field) => {
      if (field.category === key) {
        totalColumnsWidth += 130
        columns.push({
          width: 130,
          title: field.tooltip ? (
            <Tooltip title={field.tooltip}>
              {field.title} <FontAwesomeIcon icon={faQuestionCircle} />
            </Tooltip>
          ) : (
            field.title
          ),
          key: field.key,
          render: (row: TableRow) => {
            if (row.loading) return <Spin size="small" />
            if (field.type === 'number') {
              return FormatNumber(row.result.currentPeriod[field.key])
            }
            if (field.type === 'percentage') {
              return FormatPercent(row.result.currentPeriod[field.key])
            }
            if (field.type === 'currency') {
              return FormatCurrency(row.result.currentPeriod[field.key], workspace.currency)
            }
            if (field.type === 'duration') {
              return FormatDuration(row.result.currentPeriod[field.key])
            }
            if (field.type === 'custom' && field.customRender) {
              return field.customRender(
                row.result.currentPeriod,
                row.result.previousPeriod,
                workspace.currency
              )
            }
            return row.result.currentPeriod[field.key]
          }
        })
      }
    })

    // add css borders to header categories
    let columnClass = CSS.borderBottom.solid1
    if (!isLastCategory) {
      columnClass = columnClass + ' ' + CSS.borderRight.solid1

      // add a right css border to the last column of the category
      columns[columns.length - 1].className =
        columns[columns.length - 1].className + ' ' + CSS.borderRight.solid1
    }

    categories.push({
      key: key,
      title: <span style={{ textTransform: 'capitalize' }}>{key}</span>,
      width: totalColumnsWidth,
      children: columns,
      className: columnClass
    })
  })

  return categories
}

interface DimensionsSelectorProps {
  dimension1: string
  dimension2: string
  dimension3: string
  dimensionsMap: Record<string, DimensionDefinition>
}

const DimensionsSelector = (props: DimensionsSelectorProps) => {
  const [searchParams, setSearchParams] = useSearchParams()
  const [dimension1, setDimension1] = useState(props.dimension1)
  const [dimension2, setDimension2] = useState(props.dimension2)
  const [dimension3, setDimension3] = useState(props.dimension3)
  const [modalVisible, setModalVisible] = useState(false)
  const renderField = (field: DimensionDefinition) => {
    return (
      <>
        <TableTag table={field.key.split('.')[0]} />
        {field.title}
      </>
    )
  }
  const valueStyle = {
    fontSize: 12,
    backgroundColor: '#F3F6FC',
    borderRadius: 3,
    padding: '1px 4px'
  }

  return (
    <>
      <Tooltip title="Select dimensions">
        <Button onClick={() => setModalVisible(true)}>
          {/* <span className="semibold padding-r-m">Dimensions:</span> */}
          <span style={valueStyle}>{renderField(props.dimensionsMap[dimension1])}</span>
          <span className={CSS.padding_h_s}>&gt;</span>
          <span style={valueStyle}>{renderField(props.dimensionsMap[dimension2])}</span>
          <span className={CSS.padding_h_s}>&gt;</span>
          <span style={valueStyle}>{renderField(props.dimensionsMap[dimension3])}</span>
        </Button>
      </Tooltip>
      {modalVisible && (
        <Modal
          title="Select dimensions"
          open={true}
          onOk={() => {
            const allParams: any = {}
            searchParams.forEach((value, key: string) => {
              allParams[key] = value
            })
            setSearchParams({ ...allParams, dimension1, dimension2, dimension3 })
            setModalVisible(false)
          }}
          onCancel={() => setModalVisible(false)}
        >
          <Select
            style={{ width: '100%' }}
            className={CSS.margin_b_m}
            dropdownMatchSelectWidth={false}
            value={dimension1}
            options={map(props.dimensionsMap, (field) => {
              return {
                value: field.key,
                label: renderField(field)
              }
            })}
            onChange={setDimension1}
          />
          <Select
            style={{ width: '100%' }}
            className={CSS.margin_b_m}
            dropdownMatchSelectWidth={false}
            value={dimension2}
            options={map(props.dimensionsMap, (field) => {
              return {
                value: field.key,
                label: (
                  <>
                    <TableTag table={field.key.split('.')[0]} />
                    {field.title}
                  </>
                )
              }
            })}
            onChange={setDimension2}
          />
          <Select
            style={{ width: '100%' }}
            className={CSS.margin_b_m}
            dropdownMatchSelectWidth={false}
            value={dimension3}
            options={map(props.dimensionsMap, (field) => {
              return {
                value: field.key,
                label: (
                  <>
                    <TableTag table={field.key.split('.')[0]} />
                    {field.title}
                  </>
                )
              }
            })}
            onChange={setDimension3}
          />
        </Modal>
      )}
    </>
  )
}

interface MeasuresSelectorProps {
  measures: string[]
  fieldsMap: { [key: string]: MeasureDefinition }
}

const MeasuresSelector = (props: MeasuresSelectorProps) => {
  const [modalVisible, setModalVisible] = useState(false)
  const [searchParams, setSearchParams] = useSearchParams()
  const renderField = (field: MeasureDefinition) => {
    return (
      <Tooltip title={field.tooltip}>
        <TableTag table={field.key.split('.')[0]} />
        {field.title}
      </Tooltip>
    )
  }

  const postviewFields = Object.values(props.fieldsMap).filter(
    (field) => field.category === 'postviews'
  )
  const ordersFields = Object.values(props.fieldsMap).filter((field) => field.category === 'orders')
  const appFields = Object.values(props.fieldsMap).filter((field) => field.category === 'app')

  // split ordersFields into 2 groups
  const splitArray = (array: any[], groupSize: number) => {
    const groups = []
    for (let i = 0; i < array.length; i += groupSize) {
      groups.push(array.slice(i, i + groupSize))
    }
    return groups
  }

  const [ordersFields1, ordersFields2] = splitArray(
    ordersFields,
    Math.ceil(ordersFields.length / 2)
  )

  const renderFields = (fields: MeasureDefinition[]) => {
    return fields.map((field: MeasureDefinition) => (
      <div key={field.key}>
        <Checkbox
          key={field.key}
          className={CSS.margin_b_xs}
          checked={props.measures.includes(field.key)}
          onChange={(e) => {
            const allParams: any = {}
            searchParams.forEach((value, key: string) => {
              allParams[key] = value
            })

            if (e.target.checked) {
              setSearchParams({
                ...allParams,
                measures: [...props.measures, field.key].join(',')
              })
            } else {
              setSearchParams({
                ...allParams,
                measures: props.measures.filter((c) => c !== field.key).join(',')
              })
            }
          }}
        >
          {renderField(field)}
        </Checkbox>
      </div>
    ))
  }

  return (
    <>
      <Tooltip title="Select measures">
        <Button onClick={() => setModalVisible(true)}>
          Measures
          <span
            className={CSS.margin_l_s}
            style={{
              backgroundColor: 'rgb(243, 246, 252)',
              color: 'rgb(78, 108, 255)',
              // backgroundColor: '#B2EBF2',
              fontSize: 11,
              fontWeight: 500,
              padding: '1px 3px',
              borderRadius: '3px'
            }}
          >
            {props.measures.length}
          </span>
        </Button>
      </Tooltip>
      {modalVisible && (
        <Modal
          title="Select measures"
          open={true}
          width="90%"
          onOk={() => {
            setModalVisible(false)
          }}
          onCancel={() => setModalVisible(false)}
        >
          {/* create flex column for each category */}
          <div style={{ display: 'flex' }}>
            <div className={CSS.padding_r_l}>
              <div className={CSS.margin_b_m}>Postviews</div>
              {renderFields(postviewFields)}
              {/* <div className={CSS.margin_v_m}>Behavior</div>
              {renderFields(behaviorFields)} */}
            </div>
            <div>
              <div className={CSS.margin_b_m}>Orders</div>
              <div style={{ display: 'flex' }}>
                <div className={CSS.padding_r_l}>{renderFields(ordersFields1)}</div>
                <div>{renderFields(ordersFields2)}</div>
              </div>
            </div>
            {appFields.length > 0 && (
              <div>
                <div className={CSS.margin_b_m}>Apps</div>
                <div style={{ display: 'flex' }}>
                  <div>{renderFields(appFields)}</div>
                </div>
              </div>
            )}
          </div>
        </Modal>
      )}
    </>
  )
}

export default TabAttributionPostviews
