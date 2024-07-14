import { Popover, Table } from 'antd'
import CSS from 'utils/css'
import FormatCurrency from 'utils/format_currency'
import FormatGrowth from 'utils/format_growth'
import FormatNumber from 'utils/format_number'
import FormatPercent from 'utils/format_percent'

export interface MeasureDefinition {
  key: string
  title: string
  tooltip?: string
  category: 'traffic' | 'postviews' | 'behavior' | 'orders' | 'lead' | 'app'
  type: 'number' | 'currency' | 'duration' | 'percentage' | 'custom'
  customRender?: (currentPeriod: any, previousPeriod: any, currency: string) => JSX.Element | string
  measures: string[]
}

export const PostviewsMeasuresMapDefinition: Record<string, MeasureDefinition> = {
  'Order.count': {
    key: 'Order.count',
    title: 'Order',
    category: 'orders',
    type: 'number',
    measures: ['Order.count']
  },
  'Order.orders_per_user': {
    key: 'Order.orders_per_user',
    title: 'Order per user',
    category: 'orders',
    type: 'number',
    measures: ['Order.orders_per_user']
  },
  'Order.subtotal_per_user': {
    key: 'Order.subtotal_per_user',
    title: 'Subtotal per user',
    category: 'orders',
    type: 'currency',
    measures: ['Order.subtotal_per_user']
  },
  'Order.subtotal_sum': {
    key: 'Order.subtotal_sum',
    title: 'Revenue',
    tooltip: 'Sum of orders subtotals',
    category: 'orders',
    type: 'currency',
    measures: ['Order.subtotal_sum']
  },
  'Order.acquisition_subtotal_sum': {
    key: 'Order.acquisition_subtotal_sum',
    title: 'Acquisition revenue',
    tooltip: 'Sum of orders subtotals for new customers',
    category: 'orders',
    type: 'currency',
    measures: ['Order.acquisition_subtotal_sum']
  },
  'Order.retention_subtotal_sum': {
    key: 'Order.retention_subtotal_sum',
    title: 'Retention revenue',
    tooltip: 'Sum of orders subtotals repeating orders',
    category: 'orders',
    type: 'currency',
    measures: ['Order.retention_subtotal_sum']
  },
  'Order.avg_cart': {
    key: 'Order.avg_cart',
    title: 'Avg. cart',
    category: 'orders',
    type: 'currency',
    measures: ['Order.avg_cart']
  },
  'Order.acquisition_avg_cart': {
    key: 'Order.acquisition_avg_cart',
    title: 'Acquisition avg. cart',
    category: 'orders',
    type: 'currency',
    measures: ['Order.acquisition_avg_cart']
  },
  'Order.retention_avg_cart': {
    key: 'Order.retention_avg_cart',
    title: 'Retention avg. cart',
    category: 'orders',
    type: 'currency',
    measures: ['Order.retention_avg_cart']
  },
  'Order.avg_ttc': {
    key: 'Order.avg_ttc',
    title: 'Avg. TTC',
    tooltip: 'Average time to conversion',
    category: 'orders',
    type: 'duration',
    measures: ['Order.avg_ttc']
  },
  'Order.aquisition_avg_ttc': {
    key: 'Order.aquisition_avg_ttc',
    title: 'Acquisition avg. TTC',
    tooltip: 'Average time to conversion for first orders',
    category: 'orders',
    type: 'duration',
    measures: ['Order.aquisition_avg_ttc']
  },
  'Order.retention_avg_ttc': {
    key: 'Order.retention_avg_ttc',
    title: 'Retention avg. TTC',
    tooltip: 'Average time to conversion for repeating orders',
    category: 'orders',
    type: 'duration',
    measures: ['Order.retention_avg_ttc']
  },
  'Order.acquisition_count': {
    key: 'Order.acquisition_count',
    title: 'Acquisition orders',
    category: 'orders',
    type: 'number',
    measures: ['Order.acquisition_count']
  },
  'Order.retention_count': {
    key: 'Order.retention_count',
    title: 'Retention orders',
    category: 'orders',
    type: 'number',
    measures: ['Order.retention_count']
  },
  // POSTVIEWS
  'Postview.distinct_orders': {
    key: 'Postview.distinct_orders',
    title: 'Distinct orders',
    tooltip: "Count of distinct postview.conversion_id, where conversion_type = 'order'",
    category: 'orders',
    type: 'number',
    measures: ['Postview.distinct_orders']
  },
  'Postview.orders_contributions': {
    key: 'Postview.orders_contributions',
    title: 'Contributions',
    tooltip: 'Count of postviews that contributed to an order',
    category: 'orders',
    type: 'number',
    measures: ['Postview.orders_contributions']
  },

  'Postview.unique_users': {
    key: 'Postview.unique_users',
    title: 'Customers',
    tooltip: 'Postviews distinct customers',
    category: 'postviews',
    type: 'number',
    measures: ['Postview.unique_users']
  },
  'Postview.count': {
    key: 'Postview.count',
    title: 'Postviews',
    category: 'postviews',
    type: 'number',
    measures: ['Postview.count']
  },
  'Postview.contributions_count': {
    key: 'Postview.contributions_count',
    title: 'Contributions',
    category: 'orders',
    type: 'number',
    measures: ['Postview.contributions_count']
  },
  'Postview.unique_conversions': {
    key: 'Postview.unique_conversions',
    title: 'Conversions',
    category: 'orders',
    type: 'number',
    measures: ['Postview.unique_conversions']
  },
  'Postview.conversion_rate': {
    key: 'Postview.conversion_rate',
    title: 'Conversion rate',
    category: 'orders',
    type: 'percentage',
    measures: ['Postview.conversion_rate']
  },
  'Postview.linear_amount_attributed': {
    key: 'Postview.linear_amount_attributed',
    title: 'Linear revenue',
    tooltip: 'Linear revenue attributed',
    category: 'orders',
    type: 'currency',
    measures: ['Postview.linear_amount_attributed']
  },
  'Postview.linear_percentage_attributed': {
    key: 'Postview.linear_percentage_attributed',
    title: 'Linear revenue %',
    tooltip: 'Linear revenue attributed %',
    category: 'orders',
    type: 'percentage',
    measures: ['Postview.linear_percentage_attributed']
  },
  'Postview.linear_conversions_attributed': {
    key: 'Postview.linear_conversions_attributed',
    title: 'Linear conversions',
    tooltip: 'Linear conversions attributed',
    category: 'orders',
    type: 'number',
    measures: ['Postview.linear_conversions_attributed']
  },
  'Postview.alone_count': {
    key: 'Postview.alone_count',
    title: 'Alone sessions',
    tooltip: 'Postviews that are alone in the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Postview.alone_count']
  },
  'Postview.initiator_count': {
    key: 'Postview.initiator_count',
    title: 'Initiators sessions',
    tooltip: 'Postviews that are initiating the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Postview.initiator_count']
  },
  'Postview.assisting_count': {
    key: 'Postview.assisting_count',
    title: 'Assisting sessions',
    tooltip: 'Postviews that are assisting the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Postview.assisting_count']
  },
  'Postview.closer_count': {
    key: 'Postview.closer_count',
    title: 'Closer sessions',
    tooltip: 'Postviews that are closing the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Postview.closer_count']
  },
  'Postview.alone_ratio': {
    key: 'Postview.alone_ratio',
    title: 'Alone ratio',
    tooltip: 'Ratio of sessions that are alone in the conversion path',
    category: 'orders',
    type: 'percentage',
    measures: ['Postview.alone_ratio']
  },
  'Postview.initiator_ratio': {
    key: 'Postview.initiator_ratio',
    title: 'Initiator ratio',
    tooltip: 'Ratio of sessions that are initiating the conversion path',
    category: 'orders',
    type: 'percentage',
    measures: ['Postview.initiator_ratio']
  },
  'Postview.assisting_ratio': {
    key: 'Postview.assisting_ratio',
    title: 'Assisting ratio',
    tooltip: 'Ratio of sessions that are assisting the conversion path',
    category: 'orders',
    type: 'percentage',
    measures: ['Postview.assisting_ratio']
  },
  'Postview.closer_ratio': {
    key: 'Postview.closer_ratio',
    title: 'Closer ratio',
    tooltip: 'Ratio of sessions that are closing the conversion path',
    category: 'orders',
    type: 'percentage',
    measures: ['Postview.closer_ratio']
  },
  'Postview.alone_linear_conversions_attributed': {
    key: 'Postview.alone_linear_conversions_attributed',
    title: 'Alone linear conversions',
    tooltip: 'Linear conversions attributed for sessions that are alone in the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Postview.alone_linear_conversions_attributed']
  },
  'Postview.initiator_linear_conversions_attributed': {
    key: 'Postview.initiator_linear_conversions_attributed',
    title: 'Initiator linear conversions',
    tooltip: 'Linear conversions attributed for sessions that are initiating the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Postview.initiator_linear_conversions_attributed']
  },
  'Postview.assisting_linear_conversions_attributed': {
    key: 'Postview.assisting_linear_conversions_attributed',
    title: 'Assisting linear conversions',
    tooltip: 'Linear conversions attributed for sessions that are assisting the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Postview.assisting_linear_conversions_attributed']
  },
  'Postview.closer_linear_conversions_attributed': {
    key: 'Postview.closer_linear_conversions_attributed',
    title: 'Closer linear conversions',
    tooltip: 'Linear conversions attributed for sessions that are closing the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Postview.closer_linear_conversions_attributed']
  },
  'Postview.alone_linear_amount_attributed': {
    key: 'Postview.alone_linear_amount_attributed',
    title: 'Alone linear revenue',
    tooltip: 'Linear revenue attributed for sessions that are alone in the conversion path',
    category: 'orders',
    type: 'currency',
    measures: ['Postview.alone_linear_amount_attributed']
  },
  'Postview.initiator_linear_amount_attributed': {
    key: 'Postview.initiator_linear_amount_attributed',
    title: 'Initiator linear revenue',
    tooltip: 'Linear revenue attributed for sessions that are initiating the conversion path',
    category: 'orders',
    type: 'currency',
    measures: ['Postview.initiator_linear_amount_attributed']
  },
  'Postview.assisting_linear_amount_attributed': {
    key: 'Postview.assisting_linear_amount_attributed',
    title: 'Assisting linear revenue',
    tooltip: 'Linear revenue attributed for sessions that are assisting the conversion path',
    category: 'orders',
    type: 'currency',
    measures: ['Postview.assisting_linear_amount_attributed']
  },
  'Postview.closer_linear_amount_attributed': {
    key: 'Postview.closer_linear_amount_attributed',
    title: 'Closer linear revenue',
    tooltip: 'Linear revenue attributed for sessions that are closing the conversion path',
    category: 'orders',
    type: 'currency',
    measures: ['Postview.closer_linear_amount_attributed']
  },
  'Postview.attribution_roles': {
    key: 'Postview.attribution_roles',
    title: 'Role',
    category: 'orders',
    type: 'custom',
    measures: [
      'Postview.alone_count',
      'Postview.alone_ratio',
      'Postview.initiator_count',
      'Postview.initiator_ratio',
      'Postview.assisting_count',
      'Postview.assisting_ratio',
      'Postview.closer_count',
      'Postview.closer_ratio',
      'Postview.alone_linear_conversions_attributed',
      'Postview.alone_linear_amount_attributed',
      'Postview.initiator_linear_conversions_attributed',
      'Postview.initiator_linear_amount_attributed',
      'Postview.assisting_linear_conversions_attributed',
      'Postview.assisting_linear_amount_attributed',
      'Postview.closer_linear_conversions_attributed',
      'Postview.closer_linear_amount_attributed'
    ],
    customRender: (currentPeriod: any, previousPeriod: any, currency: string) => {
      const data = [
        {
          key: 'alone',
          title: 'Alone',
          width: (currentPeriod['Postview.alone_ratio'] || 0) * 100 + 'px',
          color: '#607D8B',
          contributions: currentPeriod['Postview.alone_count'] || 0,
          contributionsRatio: currentPeriod['Postview.alone_ratio'] || 0,
          previousContributionsRatio: previousPeriod['Postview.alone_ratio'] || 0,
          linearConversions: currentPeriod['Postview.alone_linear_conversions_attributed'] || 0,
          linearRevenue: currentPeriod['Postview.alone_linear_amount_attributed'] || 0
        },
        {
          key: 'initiator',
          title: 'Initiator',
          width: (currentPeriod['Postview.initiator_ratio'] || 0) * 100 + 'px',
          color: '#00BCD4',
          contributions: currentPeriod['Postview.initiator_count'] || 0,
          contributionsRatio: currentPeriod['Postview.initiator_ratio'] || 0,
          previousContributionsRatio: previousPeriod['Postview.initiator_ratio'] || 0,
          linearConversions: currentPeriod['Postview.initiator_linear_conversions_attributed'] || 0,
          linearRevenue: currentPeriod['Postview.initiator_linear_amount_attributed'] || 0
        },
        {
          key: 'assisting',
          title: 'Assisting',
          width: (currentPeriod['Postview.assisting_ratio'] || 0) * 100 + 'px',
          color: '#CDDC39',
          contributions: currentPeriod['Postview.assisting_count'] || 0,
          contributionsRatio: currentPeriod['Postview.assisting_ratio'] || 0,
          previousContributionsRatio: previousPeriod['Postview.assisting_ratio'] || 0,
          linearConversions: currentPeriod['Postview.assisting_linear_conversions_attributed'] || 0,
          linearRevenue: currentPeriod['Postview.assisting_linear_amount_attributed'] || 0
        },
        {
          key: 'closer',
          title: 'Closer',
          width: (currentPeriod['Postview.closer_ratio'] || 0) * 100 + 'px',
          color: '#F06292',
          contributions: currentPeriod['Postview.closer_count'] || 0,
          contributionsRatio: currentPeriod['Postview.closer_ratio'] || 0,
          previousContributionsRatio: previousPeriod['Postview.closer_ratio'] || 0,
          linearConversions: currentPeriod['Postview.closer_linear_conversions_attributed'] || 0,
          linearRevenue: currentPeriod['Postview.closer_linear_amount_attributed'] || 0
        }
      ]

      const content = (
        <Table
          rowKey="key"
          dataSource={data}
          size="small"
          pagination={false}
          columns={[
            {
              key: 'title',
              title: '',
              render: (record: any) => record.title
            },
            {
              key: 'bar',
              title: '',
              render: (record: any) => (
                <div>
                  <span
                    style={{
                      width: record.width,
                      display: 'inline-block',
                      backgroundColor: record.color,
                      height: '5px'
                    }}
                  ></span>
                  <div className={CSS.font_size_xxs}>
                    {FormatGrowth(record.contributionsRatio, record.previousContributionsRatio)}
                  </div>
                </div>
              )
            },
            {
              key: 'contributionRatio',
              title: '',
              render: (record: any) => FormatPercent(record.contributionsRatio)
            },
            {
              key: 'contributions',
              title: 'Contributions',
              render: (record: any) => FormatNumber(record.contributions)
            },
            {
              key: 'linearConversions',
              title: 'Linear conversions',
              render: (record: any) => FormatNumber(record.linearConversions)
            },
            {
              key: 'linearRevenue',
              title: 'Linear revenue',
              render: (record: any) =>
                FormatCurrency(record.linearRevenue, currency, { light: true })
            }
          ]}
        />
      )
      return (
        <Popover
          content={content}
          title={null}
          trigger={['hover', 'click']}
          placement="left"
          className={CSS.padding_v_m}
        >
          <div style={{ cursor: 'help', width: '100px' }}>
            <div
              style={{
                marginBottom: '2px',
                width: (currentPeriod['Postview.alone_ratio'] || 0) * 100 + '%',
                backgroundColor: '#607D8B',
                height: '3px'
              }}
            ></div>
            <div
              style={{
                marginBottom: '2px',
                width: (currentPeriod['Postview.initiator_ratio'] || 0) * 100 + '%',
                backgroundColor: '#00BCD4',
                height: '3px'
              }}
            ></div>
            <div
              style={{
                marginBottom: '2px',
                width: (currentPeriod['Postview.assisting_ratio'] || 0) * 100 + '%',
                backgroundColor: '#CDDC39',
                height: '3px'
              }}
            ></div>
            <div
              style={{
                marginBottom: '2px',
                width: (currentPeriod['Postview.closer_ratio'] || 0) * 100 + '%',
                backgroundColor: '#F06292',
                height: '3px'
              }}
            ></div>
          </div>
        </Popover>
      )
    }
  }
}
export interface DimensionDefinition {
  key: string
  title: string
  tooltip?: string
  category: 'session' | 'postview' | 'impression' | 'user' | 'order' | 'cart' | 'lead' | 'app'
  type: 'string' | 'boolean'
  customRender?: (values: any[], currency: string) => JSX.Element | string
  dimension: string
}

export const DimensionsMapDefinition: Record<string, DimensionDefinition> = {
  // Postviews
  'Postview.channel_group_id': {
    key: 'Postview.channel_group_id',
    title: 'Channel group',
    category: 'postview',
    type: 'string',
    dimension: 'Postview.channel_group_id'
  },
  'Postview.channel_id': {
    key: 'Postview.channel_id',
    title: 'Channel',
    category: 'postview',
    type: 'string',
    dimension: 'Postview.channel_id'
  },
  'Postview.channel_origin_id': {
    key: 'Postview.channel_origin_id',
    title: 'Channel origin',
    category: 'postview',
    type: 'string',
    dimension: 'Postview.channel_origin_id'
  },
  'Postview.domain_id': {
    key: 'Postview.domain_id',
    title: 'Domain',
    category: 'postview',
    type: 'string',
    dimension: 'Postview.domain_id'
  },
  'Postview.utm_source': {
    key: 'Postview.utm_source',
    title: 'UTM source',
    category: 'postview',
    type: 'string',
    dimension: 'Postview.utm_source'
  },
  'Postview.utm_medium': {
    key: 'Postview.utm_medium',
    title: 'UTM medium',
    category: 'postview',
    type: 'string',
    dimension: 'Postview.utm_medium'
  },
  'Postview.utm_campaign': {
    key: 'Postview.utm_campaign',
    title: 'UTM campaign',
    category: 'postview',
    type: 'string',
    dimension: 'Postview.utm_campaign'
  },
  'Postview.utm_term': {
    key: 'Postview.utm_term',
    title: 'UTM term',
    category: 'postview',
    type: 'string',
    dimension: 'Postview.utm_term'
  },
  'Postview.utm_content': {
    key: 'Postview.utm_content',
    title: 'UTM content',
    category: 'postview',
    type: 'string',
    dimension: 'Postview.utm_content'
  },
  'Postview.is_first_conversion': {
    key: 'Postview.is_first_conversion',
    title: 'First conversion',
    category: 'postview',
    type: 'boolean',
    dimension: 'Postview.is_first_conversion'
  },
  'Postview.bounced': {
    key: 'Postview.bounced',
    title: 'Bounced',
    category: 'postview',
    type: 'boolean',
    dimension: 'Postview.bounced'
  },
  'Postview.role': {
    key: 'Postview.role',
    title: 'Role',
    category: 'postview',
    type: 'string',
    dimension: 'Postview.role'
  }
}
