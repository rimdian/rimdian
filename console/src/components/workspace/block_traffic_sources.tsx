import { ReactNode, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { Tooltip, Table } from 'antd'
import { Query } from '@cubejs-client/core'
import { CubeContext } from '@cubejs-client/react'
// import { Domain } from 'interfaces'
import FormatCurrency from 'utils/format_currency'
// import FormatPercent from 'utils/format_percent'
import FormatNumber from 'utils/format_number'
import CSS from 'utils/css'
import Block from 'components/common/block'

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
  const { cubejsApi } = useContext(CubeContext)
  const refreshAt = useRef(0)
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
      limit: 100
    }
  }, [props.dateFrom, props.dateTo, props.timezone])

  useEffect(() => {
    if (refreshAt.current === props.refreshAt) {
      return
    }

    refreshAt.current = props.refreshAt

    setLoading(true)

    cubejsApi
      .load(query)
      .then((resultSet) => {
        setTableData(resultSet.tablePivot())
        setLoading(false)
        setError(undefined)
      })
      .catch((error) => {
        setError(error.toString())
      })
  }, [query, props.refreshAt, cubejsApi])

  // debug sql query
  //  cubejsApi
  //     .sql(query)
  //     .then((q: any) => {
  //       console.log(q[0].sqlQuery.sql.sql)
  //     })
  //     .catch((error) => {
  //       console.log(error)
  //     })

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
