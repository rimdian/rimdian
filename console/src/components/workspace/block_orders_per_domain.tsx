import { ReactNode, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { Tooltip, Table } from 'antd'
import { Query } from '@cubejs-client/core'
import { CubeContext } from '@cubejs-client/react'
import { Domain } from 'interfaces'
import FormatCurrency from 'utils/format_currency'
import FormatPercent from 'utils/format_percent'
import FormatNumber from 'utils/format_number'
import CSS from 'utils/css'
import Block from 'components/common/block'

export type OrdersPerDomainProps = {
  workspaceId: string
  domains: Domain[]
  currency: string
  timezone: string
  refreshAt: number
  dateFrom: string
  dateTo: string
  dateFromPrevious: string
  dateToPrevious: string
  tooltip?: ReactNode | string
}

export const OrdersPerDomain = (props: OrdersPerDomainProps) => {
  const { cubeApi } = useContext(CubeContext)
  const refreshAt = useRef(0)
  const [loading, setLoading] = useState<boolean>(true)
  const [tableData, setTableData] = useState<any[]>([])
  const [error, setError] = useState<string | undefined>(undefined)

  const query: Query = useMemo(() => {
    return {
      measures: ['Order.count', 'Order.subtotal_sum', 'Order.avg_cart', 'Order.retention_ratio'],
      dimensions: ['Order.domain_id'],
      timeDimensions: [
        {
          dimension: 'Order.created_at_trunc',
          granularity: undefined, // total count
          dateRange: [props.dateFrom, props.dateTo]
        }
      ],
      timezone: props.timezone
    }
  }, [props.dateFrom, props.dateTo, props.timezone])

  useEffect(() => {
    if (refreshAt.current === props.refreshAt) {
      return
    }

    refreshAt.current = props.refreshAt

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
  }, [query, props.refreshAt, cubeApi])

  let title = 'Orders per domain'

  if (error) {
    title = `${title} - ${error.toString()}`
  }

  return (
    <Block title={title} small classNames={[CSS.margin_t_l]}>
      <Table
        rowKey="Order.domain_id"
        size="small"
        loading={loading}
        dataSource={tableData}
        pagination={{
          pageSize: 5,
          showLessItems: true,
          hideOnSinglePage: true,
          showSizeChanger: false
        }}
        columns={[
          {
            title: 'Domain',
            key: 'domain',
            render: (record: any) => {
              const domain = props.domains.find((x: any) => x.id === record['Order.domain_id'])
              if (domain) return domain.name
              return record['Order.domain_id']
            }
          },
          {
            title: 'Orders',
            key: 'orders',
            sorter: (a: any, b: any) => {
              if (a['Order.count'] < b['Order.count']) {
                return -1
              }
              if (a['Order.count'] > b['Order.count']) {
                return 1
              }
              return 0
            },
            sortOrder: 'descend',
            showSorterTooltip: false,
            render: (record: any) => (
              <span className={CSS.font_weight_semibold}>
                {FormatNumber(record['Order.count'])}
              </span>
            )
          },
          {
            title: 'Revenue',
            key: 'revenue',
            render: (record: any) => FormatCurrency(record['Order.subtotal_sum'], props.currency)
          },
          {
            title: 'Avg. cart',
            key: 'avg_cart',
            render: (record: any) => FormatCurrency(record['Order.avg_cart'], props.currency)
          },
          {
            title: <Tooltip title="Repeating orders">Retention</Tooltip>,
            key: 'retention',
            render: (record: any) => FormatPercent(record['Order.retention_ratio'])
          }
        ]}
      />
    </Block>
  )
}
