import { Radio, RadioChangeEvent, Table, Tooltip } from 'antd'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { useSearchParams } from 'react-router-dom'
import { useAccount } from 'components/login/context_account'
import { CubeContext } from '@cubejs-client/react'
import { Filter, Query, ResultSet, SqlData } from '@cubejs-client/core'
import { useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { useDateRangeCtx } from 'components/common/context_date_range'
import { cloneDeep } from 'lodash'
import FormatNumber from 'utils/format_number'
import FormatCurrency from 'utils/format_currency'
import FormatDuration from 'utils/format_duration'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faChevronRight } from '@fortawesome/free-solid-svg-icons'
import { ButtonSQLExecuted, ExecutedSQL } from './button_sql_executed'
import CSS from 'utils/css'
import { css } from '@emotion/css'
import FormatGrowth from 'utils/format_growth'
import { faCircleQuestion } from '@fortawesome/free-regular-svg-icons'
import { PartialDeviceTypeIcon } from 'components/common/partial_device_icon'
import { Fullscreenable } from 'components/common/fullscreenable'
import Block from 'components/common/block'
import { KPI } from 'components/common/partial_kpi'

interface Params {
  sortKey: string
  sortOrder: string
  conversions_filter: string
  date_from: string
  date_to: string
  vs_date_from: string
  vs_date_to: string
  refresh_key: string
}

const TabAttributionCrossDevices = () => {
  const accountCtx = useAccount()
  const workspaceCtx = useCurrentWorkspaceCtx()

  const dateRangeCtx = useDateRangeCtx()
  const dateFrom = dateRangeCtx.dateRange[0].format('YYYY-MM-DD')
  const dateTo = dateRangeCtx.dateRange[1].format('YYYY-MM-DD')
  const dateFromPrevious = dateRangeCtx.dateRangePrevious[0].format('YYYY-MM-DD')
  const dateToPrevious = dateRangeCtx.dateRangePrevious[1].format('YYYY-MM-DD')

  const [searchParams, setSearchParams] = useSearchParams()
  const isMounted = useRef(true)
  const paramsHash = useRef<string | undefined>(undefined)
  const mutexObj = useMemo(() => {
    return {}
  }, [])
  const { cubejsApi } = useContext(CubeContext)
  const [executedSQL, setExecutedSQL] = useState<ExecutedSQL[]>([])

  const params: Params = useMemo(() => {
    return {
      sortKey: searchParams.get('sortKey') || 'Order.count',
      sortOrder: searchParams.get('sortOrder') || 'desc',
      conversions_filter: searchParams.get('conversions_filter') || 'all',
      date_from: dateRangeCtx.dateRange[0].format('YYYY-MM-DD'),
      date_to: dateRangeCtx.dateRange[1].format('YYYY-MM-DD'),
      vs_date_from: dateRangeCtx.dateRangePrevious[0].format('YYYY-MM-DD'),
      vs_date_to: dateRangeCtx.dateRangePrevious[1].format('YYYY-MM-DD'),
      refresh_key: searchParams.get('refresh_key') || ''
    }
  }, [searchParams, dateRangeCtx])

  const filters = useMemo(() => {
    const filters: Filter[] = [
      {
        member: 'Order.devices_funnel',
        operator: 'set'
      },
      {
        member: 'Order.devices_funnel',
        operator: 'notEquals',
        values: ['']
      }
    ]

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

    return filters
  }, [params])

  const baseQuery: Query = useMemo(() => {
    return {
      measures: ['Order.count', 'Order.subtotal_sum', 'Order.avg_cart', 'Order.avg_ttc'],
      dimensions: ['Order.devices_funnel'],
      filters: filters,
      timeDimensions: [
        {
          dimension: 'Order.created_at_trunc',
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
      limit: 300
    }
  }, [params.sortKey, params.sortOrder, dateRangeCtx, accountCtx, filters])

  const [tableData, setTableData] = useState<any[]>([])

  const fetchData = useCallback(() => {
    // console.log('fetchData', row)

    const defaultResult = {
      'Order.count': 0,
      'Order.subtotal_sum': 0,
      'Order.avg_cart': 0,
      'Order.avg_ttc': 0,
      'Order.devices_funnel': ''
    }

    const query = cloneDeep(baseQuery)
    // console.log('query', query)

    Promise.all([
      cubejsApi.sql(query, { mutexObj: mutexObj, mutexKey: 'sql' }),
      cubejsApi.load(query, { mutexObj: mutexObj, mutexKey: 'load' })
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
            name: 'current_period',
            sql: currentPeriodSQL.sql[0],
            args: currentPeriodSQL.sql[1]
          },
          {
            name: 'previous_period',
            sql: previousPeriodSQL.sql[0],
            args: previousPeriodSQL.sql[1]
          }
        ])

        const [currentPeriod, previousPeriod]: [ResultSet, ResultSet] = resultSet.decompose()

        const currentData = currentPeriod.rawData()
        const previousData = previousPeriod.rawData()

        // console.log('currentData', currentData)
        // console.log('previousData', previousData)

        // loop over current period data
        const rows: any[] = []

        currentData.forEach((currentRow) => {
          const row = {
            currentPeriod: {} as any,
            previousPeriod: cloneDeep(defaultResult)
          }

          // add current period data
          Object.keys(currentRow).forEach((key) => {
            row.currentPeriod[key] = currentRow[key]
          })

          rows.push(row)
        })

        // add previous period data or create missing rows
        previousData.forEach((previousRow) => {
          // check if we have a corresponding existing row
          let existingRow = rows.find((r) => {
            // check if dimensions are equal
            return r.currentPeriod['Order.devices_funnel'] === previousRow['Order.devices_funnel']
          })

          if (existingRow) {
            Object.keys(previousRow).forEach((key) => {
              existingRow.previousPeriod[key] = previousRow[key]
            })
          }

          if (!existingRow) {
            existingRow = {
              currentPeriod: cloneDeep(defaultResult),
              previousPeriod: cloneDeep(defaultResult)
            }
            // add previous period data
            Object.keys(previousRow).forEach((key) => {
              existingRow.previousPeriod[key] = previousRow[key]
            })
            rows.push(existingRow)
          }
        })

        // console.log('rows', rows)
        setTableData(rows)
      })
      .catch((error) => console.error(error))
  }, [mutexObj, baseQuery, cubejsApi])

  // load the first time or when the params change
  useEffect(() => {
    // first load
    if (!paramsHash.current) {
      // console.log('first load params is', params)
      paramsHash.current = JSON.stringify(params)
      fetchData()
      return
    }

    // if params changed, fetch new data
    if (JSON.stringify(params) !== paramsHash.current) {
      // console.log('update params is', params)
      paramsHash.current = JSON.stringify(params)
      // reset table data
      setTableData([])
      fetchData()
    }
  }, [params, fetchData])

  // unmounting component only, dont put things inside this effect
  useEffect(() => {
    isMounted.current = true
    return () => {
      isMounted.current = false
    }
  }, [])

  // console.log('executedSQL', executedSQL)
  return (
    <>
      <div className={css(CSS.leftRight)}>
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

      <Block grid={true}>
        <KPI
          title="Cross-device ratio"
          tooltip="Cross device ratio (cross device count / total count)"
          valueType="percent"
          measure="Order.cross_device_ratio"
          timeDimension="Order.created_at_trunc"
          filters={filters}
          color="purple"
          workspaceId={workspaceCtx.workspace.id}
          timezone={accountCtx.account?.account.timezone || 'UTC'}
          refreshAt={dateRangeCtx.refreshAt}
          dateFrom={dateFrom}
          dateTo={dateTo}
          dateFromPrevious={dateFromPrevious}
          dateToPrevious={dateToPrevious}
        />
        <KPI
          title="Desktop only"
          tooltip="Desktop device ratio (desktop device count / total count)"
          valueType="percent"
          measure="Order.desktop_device_ratio"
          timeDimension="Order.created_at_trunc"
          filters={filters}
          color="purple"
          workspaceId={workspaceCtx.workspace.id}
          timezone={accountCtx.account?.account.timezone || 'UTC'}
          refreshAt={dateRangeCtx.refreshAt}
          dateFrom={dateFrom}
          dateTo={dateTo}
          dateFromPrevious={dateFromPrevious}
          dateToPrevious={dateToPrevious}
        />
        <KPI
          title="Mobile only"
          tooltip="Mobile device ratio (mobile device count / total count)"
          valueType="percent"
          measure="Order.mobile_device_ratio"
          timeDimension="Order.created_at_trunc"
          filters={filters}
          color="purple"
          workspaceId={workspaceCtx.workspace.id}
          timezone={accountCtx.account?.account.timezone || 'UTC'}
          refreshAt={dateRangeCtx.refreshAt}
          dateFrom={dateFrom}
          dateTo={dateTo}
          dateFromPrevious={dateFromPrevious}
          dateToPrevious={dateToPrevious}
        />
        <KPI
          title="Tablet only"
          tooltip="Tablet device ratio (tablet device count / total count)"
          valueType="percent"
          measure="Order.tablet_device_ratio"
          timeDimension="Order.created_at_trunc"
          filters={filters}
          color="purple"
          workspaceId={workspaceCtx.workspace.id}
          timezone={accountCtx.account?.account.timezone || 'UTC'}
          refreshAt={dateRangeCtx.refreshAt}
          dateFrom={dateFrom}
          dateTo={dateTo}
          dateFromPrevious={dateFromPrevious}
          dateToPrevious={dateToPrevious}
        />
      </Block>

      <Fullscreenable>
        <Table
          dataSource={tableData}
          size="middle"
          rowClassName={(record) => css([record.key === 'root' && CSS.tableTotals])}
          // onChange={onTableChange}
          pagination={{
            position: ['bottomRight'],
            pageSize: 10,
            showSizeChanger: false,
            hideOnSinglePage: true
          }}
          rowKey={(record) => record.currentPeriod['Order.devices_funnel']}
          columns={[
            {
              title: 'Devices paths (top 300)',
              key: 'path',
              render: (row) => {
                const devices = row.currentPeriod['Order.devices_funnel'].split('~')
                return devices.map((deviceType: string, i: number) => {
                  return (
                    <span className={CSS.font_size_s} key={i}>
                      {PartialDeviceTypeIcon(deviceType)}
                      {i < devices.length - 1 && (
                        <span
                          className={
                            CSS.padding_h_xs + ' ' + CSS.opacity_50 + ' ' + CSS.font_size_xxs
                          }
                        >
                          <FontAwesomeIcon icon={faChevronRight} />
                        </span>
                      )}
                    </span>
                  )
                })
              }
            },
            {
              title: (
                <Tooltip title="Sum of orders">
                  Orders <FontAwesomeIcon icon={faCircleQuestion} />
                </Tooltip>
              ),
              key: 'orders',
              // sortOrder:
              //   params.sortKey === 'Order.count'
              //     ? params.sortOrder === 'desc'
              //       ? 'descend'
              //       : 'ascend'
              //     : undefined,
              defaultSortOrder: 'descend',
              sorter: (a: any, b: any) =>
                a.currentPeriod['Order.count'] - b.currentPeriod['Order.count'],
              sortDirections: ['descend'],

              render: (row) => {
                return (
                  <>
                    {FormatNumber(row.currentPeriod['Order.count'])}
                    <span className={CSS.font_size_xxs}>
                      {FormatGrowth(
                        row.currentPeriod['Order.count'],
                        row.previousPeriod['Order.count']
                      )}
                    </span>
                  </>
                )
              }
            },
            {
              title: (
                <Tooltip title="Sum of orders subtotal">
                  Revenue <FontAwesomeIcon icon={faCircleQuestion} />
                </Tooltip>
              ),
              key: 'subtotal_sum',
              // sortOrder:
              //   params.sortKey === 'Order.subtotal_sum'
              //     ? params.sortOrder === 'desc'
              //       ? 'descend'
              //       : 'ascend'
              //     : undefined,
              sorter: (a: any, b: any) =>
                a.currentPeriod['Order.subtotal_sum'] - b.currentPeriod['Order.subtotal_sum'],
              sortDirections: ['descend'],
              render: (row) => {
                return (
                  <>
                    {FormatCurrency(
                      row.currentPeriod['Order.subtotal_sum'],
                      workspaceCtx.workspace?.currency
                    )}
                    <span className={CSS.font_size_xxs}>
                      {FormatGrowth(
                        row.currentPeriod['Order.subtotal_sum'],
                        row.previousPeriod['Order.subtotal_sum']
                      )}
                    </span>
                  </>
                )
              }
            },
            {
              title: (
                <Tooltip title="Average cart">
                  Avg. Cart <FontAwesomeIcon icon={faCircleQuestion} />
                </Tooltip>
              ),
              key: 'avg_cart',
              // sortOrder:
              //   params.sortKey === 'Order.avg_cart'
              //     ? params.sortOrder === 'desc'
              //       ? 'descend'
              //       : 'ascend'
              //     : undefined,
              sorter: (a: any, b: any) =>
                a.currentPeriod['Order.avg_cart'] - b.currentPeriod['Order.avg_cart'],
              sortDirections: ['descend'],
              render: (row) => {
                return (
                  <>
                    {FormatCurrency(
                      row.currentPeriod['Order.avg_cart'],
                      workspaceCtx.workspace?.currency
                    )}{' '}
                    <span className={CSS.font_size_xxs}>
                      {FormatGrowth(
                        row.currentPeriod['Order.avg_cart'],
                        row.previousPeriod['Order.avg_cart']
                      )}
                    </span>
                  </>
                )
              }
            },
            {
              title: (
                <Tooltip title="Average time to conversion">
                  Avg. TTC <FontAwesomeIcon icon={faCircleQuestion} />
                </Tooltip>
              ),
              key: 'avg_ttc',
              // sortOrder:
              //   params.sortKey === 'Order.avg_ttc'
              //     ? params.sortOrder === 'desc'
              //       ? 'descend'
              //       : 'ascend'
              //     : undefined,
              sorter: (a: any, b: any) =>
                a.currentPeriod['Order.avg_ttc'] - b.currentPeriod['Order.avg_ttc'],
              sortDirections: ['descend'],
              render: (row) => {
                return (
                  <>
                    {FormatDuration(row.currentPeriod['Order.avg_ttc'], 0)}{' '}
                    <span className={CSS.font_size_xxs}>
                      {FormatGrowth(
                        row.currentPeriod['Order.avg_ttc'],
                        row.previousPeriod['Order.avg_ttc']
                      )}
                    </span>
                  </>
                )
              }
            }
          ]}
        />
      </Fullscreenable>

      <div className={CSS.text_right + ' ' + CSS.margin_t_l}>
        <ButtonSQLExecuted queries={executedSQL} />
      </div>
    </>
  )
}

export default TabAttributionCrossDevices
