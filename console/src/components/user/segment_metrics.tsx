import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { Segment } from 'components/segment/interfaces'
import CSS from 'utils/css'
import Block from 'components/common/block'
import { MetricWidget } from 'components/common/partial_metric_widget'
import { useMemo } from 'react'
import { LinesWidget } from 'components/common/partial_lines_widget'
import dayjs from 'dayjs'

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

  return (
    <>
      <Block classNames={[CSS.margin_b_l]} grid={true}>
        <MetricWidget
          workspaceCtx={workspaceCtx}
          schema="User"
          measure="avg_ltv"
          filters={filters}
          refreshKey={props.currentSegment?.id || ''}
        />
        <MetricWidget
          workspaceCtx={workspaceCtx}
          schema="User"
          measure="avg_orders"
          title="Avg orders"
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
    </>
  )
}

export default SegmentMetrics
