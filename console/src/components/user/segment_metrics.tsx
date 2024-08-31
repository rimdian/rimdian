import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { Segment } from 'components/segment/interfaces'
import CSS from 'utils/css'
import Block from 'components/common/block'
import { MetricWidget } from 'components/common/partial_metric_widget'
import { useMemo } from 'react'
import { LinesWidget } from 'components/common/partial_lines_widget'
import dayjs from 'dayjs'
import { TableWidget } from 'components/common/partial_table_widget'
import { Col, Divider, Image, Row, Tag } from 'antd'

interface SegmentMetricsProps {
  timezone: string
  currentSegment?: Segment
}

const SegmentMetrics = (props: SegmentMetricsProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  const filters = useMemo(() => {
    if (props.currentSegment) {
      return [
        { member: 'User_segment.segment_id', operator: 'equals', values: [props.currentSegment.id] }
      ]
    }
    return []
  }, [props.currentSegment])

  const last30DaysFrom = dayjs().subtract(30, 'days').startOf('day').format('YYYY-MM-DD')
  const last30DaysTo = dayjs().endOf('day').format('YYYY-MM-DD')

  const webDomains = useMemo(() => {
    const domains: string[] = []
    workspaceCtx.workspace.domains.forEach((domain) => {
      if (domain.type === 'web') {
        domains.push(domain.id)
      }
    })
    return domains
  }, [workspaceCtx.workspace.domains])

  return (
    <>
      <Block grid={true}>
        <MetricWidget
          workspaceCtx={workspaceCtx}
          schema="User"
          measure="avg_ltv"
          title="LTV per user"
          filters={filters}
          refreshKey={props.currentSegment?.id || ''}
        />
        <MetricWidget
          workspaceCtx={workspaceCtx}
          schema="User"
          measure="avg_orders"
          title="Orders per user"
          filters={filters}
          refreshKey={props.currentSegment?.id || ''}
        />
        <MetricWidget
          workspaceCtx={workspaceCtx}
          schema="Session"
          measure="avg_duration"
          title="Avg session"
          filters={filters}
          refreshKey={props.currentSegment?.id || ''}
        />
        <LinesWidget
          workspaceCtx={workspaceCtx}
          schema="Data_log"
          line1={{
            name: 'Enter',
            measure: 'count',
            color: '#9CCC65',
            filters: [
              {
                member: 'Data_log.kind',
                operator: 'equals',
                values: ['segment']
              },
              {
                member: 'Data_log.action',
                operator: 'equals',
                values: ['enter']
              },
              {
                member: 'Data_log.item_external_id',
                operator: 'equals',
                values: [props.currentSegment?.id]
              }
            ],
            dateFrom: last30DaysFrom,
            dateTo: last30DaysTo
          }}
          line2={{
            name: 'Exit',
            measure: 'count',
            color: '#FFA726',
            filters: [
              {
                member: 'Data_log.kind',
                operator: 'equals',
                values: ['segment']
              },
              {
                member: 'Data_log.action',
                operator: 'equals',
                values: ['exit']
              },
              {
                member: 'Data_log.item_external_id',
                operator: 'equals',
                values: [props.currentSegment?.id]
              }
            ],
            dateFrom: last30DaysFrom,
            dateTo: last30DaysTo
          }}
          timeDimension="Data_log.event_at_trunc"
          title="Enter vs exit segment (30 days)"
          refreshKey={props.currentSegment?.id || ''}
        />
      </Block>

      {/* <Divider plain>
        Acquisition vs Retention
      </Divider> */}
      <Row gutter={24}>
        <Col span={12}>
          <Divider plain orientation="left">
            Acquisition
          </Divider>

          <Block classNames={[CSS.margin_b_l]} grid={true}>
            <MetricWidget
              workspaceCtx={workspaceCtx}
              schema="Order"
              measure="acquisition_avg_ttc"
              title="Avg time to convert"
              filters={[
                ...filters,
                {
                  member: 'Order.domain_id',
                  operator: 'equals',
                  values: webDomains
                }
              ]}
              refreshKey={props.currentSegment?.id || ''}
            />
            <MetricWidget
              workspaceCtx={workspaceCtx}
              schema="Order"
              measure="acquisition_avg_cart"
              title="Avg cart"
              filters={filters}
              refreshKey={props.currentSegment?.id || ''}
            />
          </Block>
          <TableWidget
            workspaceCtx={workspaceCtx}
            title="Top 5 channels (90 days)"
            measures={[
              { measure: 'Session.orders_contributions', title: 'Contributions' },
              { measure: 'Session.linear_amount_attributed', title: 'Linear revenue' }
            ]}
            timeDimension="Session.event_at_trunc"
            dateFrom={dayjs().subtract(90, 'days').startOf('day').format('YYYY-MM-DD')}
            dateTo={dayjs().endOf('day').format('YYYY-MM-DD')}
            dimensions={[
              {
                dimension: 'Session.channel_id',
                title: ' ',
                render: (x) => {
                  // find channel in workspace
                  const channel = workspaceCtx.workspace.channels.find(
                    (c) => c.id === x['Session.channel_id']
                  )
                  if (!channel) {
                    return x['Session.channel_id']
                  }

                  // find channel group
                  const group = workspaceCtx.workspace.channel_groups.find(
                    (g) => g.id === channel.group_id
                  )

                  return (
                    <>
                      <Tag color={group?.color}>{group?.name}</Tag> {channel.name}
                    </>
                  )
                }
              }
            ]}
            filters={[
              ...filters,
              {
                member: 'Session.is_first_conversion',
                operator: 'equals',
                values: [1]
              }
            ]}
            order={{ 'Session.orders_contributions': 'desc' }}
            limit={5}
            size="middle"
            refreshKey={props.currentSegment?.id || ''}
          />
          <TableWidget
            workspaceCtx={workspaceCtx}
            title="Top 5 products (90 days)"
            measures={[{ measure: 'Order_item.count', title: 'Sold' }]}
            timeDimension="Session.event_at_trunc"
            dateFrom={dayjs().subtract(90, 'days').startOf('day').format('YYYY-MM-DD')}
            dateTo={dayjs().endOf('day').format('YYYY-MM-DD')}
            dimensions={[
              {
                dimension: 'Order_item.image_url',
                title: ' ',
                render: (x) => {
                  if (!x['Order_item.image_url'] || x['Order_item.image_url'] === '') {
                    return null
                  }
                  return (
                    <Image
                      width={24}
                      height={24}
                      src={x['Order_item.image_url']}
                      fallback="data:image/png;base64,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"
                      preview={true}
                    />
                  )
                }
              },
              {
                dimension: 'Order_item.name',
                title: ' ',
                render: (x) => {
                  return <>{x['Order_item.name']}</>
                }
              }
            ]}
            filters={[
              ...filters,
              {
                member: 'Order.is_first_conversion',
                operator: 'equals',
                values: [1]
              }
            ]}
            order={{ 'Order_item.count': 'desc' }}
            limit={5}
            size="middle"
            refreshKey={props.currentSegment?.id || ''}
          />
        </Col>
        <Col span={12}>
          <Divider plain orientation="left">
            Retention
          </Divider>

          <Block classNames={[CSS.margin_b_l]} grid={true}>
            <MetricWidget
              workspaceCtx={workspaceCtx}
              schema="Order"
              measure="retention_avg_ttc"
              title="Avg time to convert"
              filters={[
                ...filters,
                {
                  member: 'Order.domain_id',
                  operator: 'equals',
                  values: webDomains
                }
              ]}
              refreshKey={props.currentSegment?.id || ''}
            />
            <MetricWidget
              workspaceCtx={workspaceCtx}
              schema="Order"
              measure="retention_avg_cart"
              title="Avg cart"
              filters={filters}
              refreshKey={props.currentSegment?.id || ''}
            />
          </Block>
          <TableWidget
            workspaceCtx={workspaceCtx}
            title="Top 5 channels (90 days)"
            measures={[
              { measure: 'Session.orders_contributions', title: 'Contributions' },
              { measure: 'Session.linear_amount_attributed', title: 'Linear revenue' }
            ]}
            timeDimension="Session.event_at_trunc"
            dateFrom={dayjs().subtract(90, 'days').startOf('day').format('YYYY-MM-DD')}
            dateTo={dayjs().endOf('day').format('YYYY-MM-DD')}
            dimensions={[
              {
                dimension: 'Session.channel_id',
                title: ' ',
                render: (x) => {
                  // find channel in workspace
                  const channel = workspaceCtx.workspace.channels.find(
                    (c) => c.id === x['Session.channel_id']
                  )
                  if (!channel) {
                    return x['Session.channel_id']
                  }

                  // find channel group
                  const group = workspaceCtx.workspace.channel_groups.find(
                    (g) => g.id === channel.group_id
                  )

                  return (
                    <>
                      <Tag color={group?.color}>{group?.name}</Tag> {channel.name}
                    </>
                  )
                }
              }
            ]}
            filters={[
              ...filters,
              {
                member: 'Session.is_first_conversion',
                operator: 'equals',
                values: [0]
              }
            ]}
            order={{ 'Session.orders_contributions': 'desc' }}
            limit={5}
            size="middle"
            refreshKey={props.currentSegment?.id || ''}
          />
          <TableWidget
            workspaceCtx={workspaceCtx}
            title="Top 5 products (90 days)"
            measures={[{ measure: 'Order_item.count', title: 'Sold' }]}
            timeDimension="Session.event_at_trunc"
            dateFrom={dayjs().subtract(90, 'days').startOf('day').format('YYYY-MM-DD')}
            dateTo={dayjs().endOf('day').format('YYYY-MM-DD')}
            dimensions={[
              {
                dimension: 'Order_item.image_url',
                title: ' ',
                render: (x) => {
                  if (!x['Order_item.image_url'] || x['Order_item.image_url'] === '') {
                    return null
                  }
                  return (
                    <Image
                      width={24}
                      height={24}
                      src={x['Order_item.image_url']}
                      fallback="data:image/png;base64,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"
                      preview={true}
                    />
                  )
                }
              },
              {
                dimension: 'Order_item.name',
                title: ' ',
                render: (x) => {
                  return <>{x['Order_item.name']}</>
                }
              }
            ]}
            filters={[
              ...filters,
              {
                member: 'Order.is_first_conversion',
                operator: 'equals',
                values: [0]
              }
            ]}
            order={{ 'Order_item.count': 'desc' }}
            limit={5}
            size="middle"
            refreshKey={props.currentSegment?.id || ''}
          />
        </Col>
      </Row>
    </>
  )
}

export default SegmentMetrics
