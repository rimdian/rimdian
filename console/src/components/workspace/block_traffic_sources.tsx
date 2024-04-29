import { ReactNode, useEffect, useMemo, useRef, useState } from 'react'
import { Tooltip, Table } from 'antd'
import { Query } from '@cubejs-client/core'
import FormatCurrency from 'utils/format_currency'
import FormatNumber from 'utils/format_number'
import CSS from 'utils/css'
import Block from 'components/common/block'
import { useRimdianCube } from './context_cube'

export type TrafficSourcesProps = {
  workspaceId: string
  currency: string
  timezone: string
  refreshAt: number
  dateFrom: string
  dateTo: string
  dateFromPrevious: string
  dateToPrevious: string
  tooltip?: ReactNode | string
}

export const TrafficSources = (props: TrafficSourcesProps) => {
  const { cubeApi } = useRimdianCube()
  const refreshAtRef = useRef(0)
  const [loading, setLoading] = useState<boolean>(true)
  const [tableData, setTableData] = useState<any[]>([])
  const [error, setError] = useState<string | undefined>(undefined)

  const query: Query = useMemo(() => {
    return {
      measures: ['Session.unique_users', 'Order.count', 'Order.subtotal_sum'],
      dimensions: ['Session.channel_origin_id'],
      timeDimensions: [
        {
          dimension: 'Session.created_at_trunc',
          granularity: undefined, // total count
          dateRange: [props.dateFrom, props.dateTo]
        }
      ],
      timezone: props.timezone,
      order: { 'Session.unique_users': 'desc' },
      limit: 100,
      renewQuery: props.refreshAt !== refreshAtRef.current
    }
  }, [props.dateFrom, props.dateTo, props.timezone, props.refreshAt])

  useEffect(() => {
    if (props.refreshAt === refreshAtRef.current) {
      return
    } else {
      refreshAtRef.current = props.refreshAt
    }

    setLoading(true)

    cubeApi
      .load(query)
      .then((resultSet) => {
        setTableData(resultSet.tablePivot())
        setLoading(false)
        setError(undefined)
      })
      .catch((error) => {
        setError(error.toString())
      })
  }, [query, cubeApi, props.refreshAt])

  let title = 'Traffic sources'

  if (error) {
    title = `${title} - ${error.toString()}`
  }

  return (
    <Block title={title} small classNames={[CSS.margin_t_l]}>
      <Table
        rowKey="Session.channel_origin_id"
        size="small"
        loading={loading}
        dataSource={tableData}
        pagination={{
          pageSize: 20,
          showLessItems: true,
          hideOnSinglePage: true,
          showSizeChanger: false
        }}
        columns={[
          {
            title: 'Traffic origin',
            key: 'channel_origin_id',
            render: (x: any) => (
              <span>
                {x['Session.channel_origin_id'] ? x['Session.channel_origin_id'] : 'not-mapped'}
              </span>
            )
          },
          {
            title: 'Users',
            key: 'users',
            sorter: (a: any, b: any) => {
              if (a['Session.unique_users'] < b['Session.unique_users']) {
                return -1
              }
              if (a['Session.unique_users'] > b['Session.unique_users']) {
                return 1
              }
              return 0
            },
            sortOrder: 'descend',
            showSorterTooltip: false,
            render: (x: any) => (
              <span className={CSS.font_weight_semibold}>
                {FormatNumber(x['Session.unique_users'])}
              </span>
            )
          },
          {
            title: 'Conversions',
            key: 'conversions',
            render: (x: any) => <span>{FormatNumber(x['Order.count'])}</span>
          },
          {
            title: <Tooltip title="Orders subtotal sum">Revenue</Tooltip>,
            key: 'subtotal_sum',
            render: (x: any) => (
              <span>{FormatCurrency(x['Order.subtotal_sum'], props.currency)}</span>
            )
          }
        ]}
      />
    </Block>
  )
}
