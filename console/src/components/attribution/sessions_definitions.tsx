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

export const SessionsMeasuresMapDefinition: Record<string, MeasureDefinition> = {
  'Session.unique_users': {
    key: 'Session.unique_users',
    title: 'Users',
    tooltip: 'Sessions distinct users',
    category: 'traffic',
    type: 'number',
    measures: ['Session.unique_users']
  },
  'Session.count': {
    key: 'Session.count',
    title: 'Sessions',
    category: 'traffic',
    type: 'number',
    measures: ['Session.count']
  },
  // BEHAVIOR
  'Session.bounce_rate': {
    key: 'Session.bounce_rate',
    title: 'Bounce rate',
    category: 'behavior',
    type: 'percentage',
    measures: ['Session.bounce_rate']
  },
  'Session.avg_pageviews_count': {
    key: 'Session.avg_pageviews_count',
    title: 'Avg. pageviews',
    tooltip: 'Avg. pageviews per session',
    category: 'behavior',
    type: 'number',
    measures: ['Session.avg_pageviews_count']
  },
  'Session.avg_duration': {
    key: 'Session.avg_duration',
    title: 'Avg. duration',
    tooltip: 'Avg. duration of a session',
    category: 'behavior',
    type: 'duration',
    measures: ['Session.avg_duration']
  },
  // ORDERS
  'Session.distinct_orders': {
    key: 'Session.distinct_orders',
    title: 'Distinct orders',
    tooltip: "Count of distinct session.conversion_id, where conversion_type = 'order'",
    category: 'orders',
    type: 'number',
    measures: ['Session.distinct_orders']
  },
  'Session.orders_contributions': {
    key: 'Session.orders_contributions',
    title: 'Contributions',
    tooltip: 'Count of sessions that contributed to an order',
    category: 'orders',
    type: 'number',
    measures: ['Session.orders_contributions']
  },
  'Session.contributions_count': {
    key: 'Session.contributions_count',
    title: 'Contributions',
    category: 'orders',
    type: 'number',
    measures: ['Session.contributions_count']
  },
  'Session.unique_conversions': {
    key: 'Session.unique_conversions',
    title: 'Conversions',
    category: 'orders',
    type: 'number',
    measures: ['Session.unique_conversions']
  },
  'Session.conversion_rate': {
    key: 'Session.conversion_rate',
    title: 'Conversion rate',
    category: 'orders',
    type: 'percentage',
    measures: ['Session.conversion_rate']
  },
  'Session.linear_amount_attributed': {
    key: 'Session.linear_amount_attributed',
    title: 'Linear revenue',
    tooltip: 'Linear revenue attributed',
    category: 'orders',
    type: 'currency',
    measures: ['Session.linear_amount_attributed']
  },
  'Session.linear_percentage_attributed': {
    key: 'Session.linear_percentage_attributed',
    title: 'Linear revenue %',
    tooltip: 'Linear revenue attributed %',
    category: 'orders',
    type: 'percentage',
    measures: ['Session.linear_percentage_attributed']
  },
  'Session.linear_conversions_attributed': {
    key: 'Session.linear_conversions_attributed',
    title: 'Linear conversions',
    tooltip: 'Linear conversions attributed',
    category: 'orders',
    type: 'number',
    measures: ['Session.linear_conversions_attributed']
  },
  'Session.alone_count': {
    key: 'Session.alone_count',
    title: 'Alone sessions',
    tooltip: 'Sessions that are alone in the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Session.alone_count']
  },
  'Session.initiator_count': {
    key: 'Session.initiator_count',
    title: 'Initiators sessions',
    tooltip: 'Sessions that are initiating the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Session.initiator_count']
  },
  'Session.assisting_count': {
    key: 'Session.assisting_count',
    title: 'Assisting sessions',
    tooltip: 'Sessions that are assisting the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Session.assisting_count']
  },
  'Session.closer_count': {
    key: 'Session.closer_count',
    title: 'Closer sessions',
    tooltip: 'Sessions that are closing the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Session.closer_count']
  },
  'Session.alone_ratio': {
    key: 'Session.alone_ratio',
    title: 'Alone ratio',
    tooltip: 'Ratio of sessions that are alone in the conversion path',
    category: 'orders',
    type: 'percentage',
    measures: ['Session.alone_ratio']
  },
  'Session.initiator_ratio': {
    key: 'Session.initiator_ratio',
    title: 'Initiator ratio',
    tooltip: 'Ratio of sessions that are initiating the conversion path',
    category: 'orders',
    type: 'percentage',
    measures: ['Session.initiator_ratio']
  },
  'Session.assisting_ratio': {
    key: 'Session.assisting_ratio',
    title: 'Assisting ratio',
    tooltip: 'Ratio of sessions that are assisting the conversion path',
    category: 'orders',
    type: 'percentage',
    measures: ['Session.assisting_ratio']
  },
  'Session.closer_ratio': {
    key: 'Session.closer_ratio',
    title: 'Closer ratio',
    tooltip: 'Ratio of sessions that are closing the conversion path',
    category: 'orders',
    type: 'percentage',
    measures: ['Session.closer_ratio']
  },
  'Session.alone_linear_conversions_attributed': {
    key: 'Session.alone_linear_conversions_attributed',
    title: 'Alone linear conversions',
    tooltip: 'Linear conversions attributed for sessions that are alone in the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Session.alone_linear_conversions_attributed']
  },
  'Session.initiator_linear_conversions_attributed': {
    key: 'Session.initiator_linear_conversions_attributed',
    title: 'Initiator linear conversions',
    tooltip: 'Linear conversions attributed for sessions that are initiating the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Session.initiator_linear_conversions_attributed']
  },
  'Session.assisting_linear_conversions_attributed': {
    key: 'Session.assisting_linear_conversions_attributed',
    title: 'Assisting linear conversions',
    tooltip: 'Linear conversions attributed for sessions that are assisting the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Session.assisting_linear_conversions_attributed']
  },
  'Session.closer_linear_conversions_attributed': {
    key: 'Session.closer_linear_conversions_attributed',
    title: 'Closer linear conversions',
    tooltip: 'Linear conversions attributed for sessions that are closing the conversion path',
    category: 'orders',
    type: 'number',
    measures: ['Session.closer_linear_conversions_attributed']
  },
  'Session.alone_linear_amount_attributed': {
    key: 'Session.alone_linear_amount_attributed',
    title: 'Alone linear revenue',
    tooltip: 'Linear revenue attributed for sessions that are alone in the conversion path',
    category: 'orders',
    type: 'currency',
    measures: ['Session.alone_linear_amount_attributed']
  },
  'Session.initiator_linear_amount_attributed': {
    key: 'Session.initiator_linear_amount_attributed',
    title: 'Initiator linear revenue',
    tooltip: 'Linear revenue attributed for sessions that are initiating the conversion path',
    category: 'orders',
    type: 'currency',
    measures: ['Session.initiator_linear_amount_attributed']
  },
  'Session.assisting_linear_amount_attributed': {
    key: 'Session.assisting_linear_amount_attributed',
    title: 'Assisting linear revenue',
    tooltip: 'Linear revenue attributed for sessions that are assisting the conversion path',
    category: 'orders',
    type: 'currency',
    measures: ['Session.assisting_linear_amount_attributed']
  },
  'Session.closer_linear_amount_attributed': {
    key: 'Session.closer_linear_amount_attributed',
    title: 'Closer linear revenue',
    tooltip: 'Linear revenue attributed for sessions that are closing the conversion path',
    category: 'orders',
    type: 'currency',
    measures: ['Session.closer_linear_amount_attributed']
  },
  'Session.attribution_roles': {
    key: 'Session.attribution_roles',
    title: 'Role',
    category: 'orders',
    type: 'custom',
    measures: [
      'Session.alone_count',
      'Session.alone_ratio',
      'Session.initiator_count',
      'Session.initiator_ratio',
      'Session.assisting_count',
      'Session.assisting_ratio',
      'Session.closer_count',
      'Session.closer_ratio',
      'Session.alone_linear_conversions_attributed',
      'Session.alone_linear_amount_attributed',
      'Session.initiator_linear_conversions_attributed',
      'Session.initiator_linear_amount_attributed',
      'Session.assisting_linear_conversions_attributed',
      'Session.assisting_linear_amount_attributed',
      'Session.closer_linear_conversions_attributed',
      'Session.closer_linear_amount_attributed'
    ],
    customRender: (currentPeriod: any, previousPeriod: any, currency: string) => {
      const data = [
        {
          key: 'alone',
          title: 'Alone',
          width: (currentPeriod['Session.alone_ratio'] || 0) * 100 + 'px',
          color: '#607D8B',
          contributions: currentPeriod['Session.alone_count'] || 0,
          contributionsRatio: currentPeriod['Session.alone_ratio'] || 0,
          previousContributionsRatio: previousPeriod['Session.alone_ratio'] || 0,
          linearConversions: currentPeriod['Session.alone_linear_conversions_attributed'] || 0,
          linearRevenue: currentPeriod['Session.alone_linear_amount_attributed'] || 0
        },
        {
          key: 'initiator',
          title: 'Initiator',
          width: (currentPeriod['Session.initiator_ratio'] || 0) * 100 + 'px',
          color: '#00BCD4',
          contributions: currentPeriod['Session.initiator_count'] || 0,
          contributionsRatio: currentPeriod['Session.initiator_ratio'] || 0,
          previousContributionsRatio: previousPeriod['Session.initiator_ratio'] || 0,
          linearConversions: currentPeriod['Session.initiator_linear_conversions_attributed'] || 0,
          linearRevenue: currentPeriod['Session.initiator_linear_amount_attributed'] || 0
        },
        {
          key: 'assisting',
          title: 'Assisting',
          width: (currentPeriod['Session.assisting_ratio'] || 0) * 100 + 'px',
          color: '#CDDC39',
          contributions: currentPeriod['Session.assisting_count'] || 0,
          contributionsRatio: currentPeriod['Session.assisting_ratio'] || 0,
          previousContributionsRatio: previousPeriod['Session.assisting_ratio'] || 0,
          linearConversions: currentPeriod['Session.assisting_linear_conversions_attributed'] || 0,
          linearRevenue: currentPeriod['Session.assisting_linear_amount_attributed'] || 0
        },
        {
          key: 'closer',
          title: 'Closer',
          width: (currentPeriod['Session.closer_ratio'] || 0) * 100 + 'px',
          color: '#F06292',
          contributions: currentPeriod['Session.closer_count'] || 0,
          contributionsRatio: currentPeriod['Session.closer_ratio'] || 0,
          previousContributionsRatio: previousPeriod['Session.closer_ratio'] || 0,
          linearConversions: currentPeriod['Session.closer_linear_conversions_attributed'] || 0,
          linearRevenue: currentPeriod['Session.closer_linear_amount_attributed'] || 0
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
          trigger={['hover', 'onClick']}
          placement="left"
          className={CSS.padding_v_m}
        >
          <div style={{ cursor: 'help', width: '100px' }}>
            <div
              style={{
                marginBottom: '2px',
                width: (currentPeriod['Session.alone_ratio'] || 0) * 100 + '%',
                backgroundColor: '#607D8B',
                height: '3px'
              }}
            ></div>
            <div
              style={{
                marginBottom: '2px',
                width: (currentPeriod['Session.initiator_ratio'] || 0) * 100 + '%',
                backgroundColor: '#00BCD4',
                height: '3px'
              }}
            ></div>
            <div
              style={{
                marginBottom: '2px',
                width: (currentPeriod['Session.assisting_ratio'] || 0) * 100 + '%',
                backgroundColor: '#CDDC39',
                height: '3px'
              }}
            ></div>
            <div
              style={{
                marginBottom: '2px',
                width: (currentPeriod['Session.closer_ratio'] || 0) * 100 + '%',
                backgroundColor: '#F06292',
                height: '3px'
              }}
            ></div>
          </div>
        </Popover>
      )
    }
  },
  'Order.count': {
    key: 'Order.count',
    title: 'Order count',
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
  'Session.channel_group_id': {
    key: 'Session.channel_group_id',
    title: 'Channel group',
    category: 'session',
    type: 'string',
    dimension: 'Session.channel_group_id'
  },
  'Session.channel_id': {
    key: 'Session.channel_id',
    title: 'Channel',
    category: 'session',
    type: 'string',
    dimension: 'Session.channel_id'
  },
  'Session.channel_origin_id': {
    key: 'Session.channel_origin_id',
    title: 'Channel origin',
    category: 'session',
    type: 'string',
    dimension: 'Session.channel_origin_id'
  },
  'Session.domain_id': {
    key: 'Session.domain_id',
    title: 'Domain',
    category: 'session',
    type: 'string',
    dimension: 'Session.domain_id'
  },
  'Session.utm_source': {
    key: 'Session.utm_source',
    title: 'UTM source',
    category: 'session',
    type: 'string',
    dimension: 'Session.utm_source'
  },
  'Session.utm_medium': {
    key: 'Session.utm_medium',
    title: 'UTM medium',
    category: 'session',
    type: 'string',
    dimension: 'Session.utm_medium'
  },
  'Session.utm_campaign': {
    key: 'Session.utm_campaign',
    title: 'UTM campaign',
    category: 'session',
    type: 'string',
    dimension: 'Session.utm_campaign'
  },
  'Session.utm_term': {
    key: 'Session.utm_term',
    title: 'UTM term',
    category: 'session',
    type: 'string',
    dimension: 'Session.utm_term'
  },
  'Session.utm_content': {
    key: 'Session.utm_content',
    title: 'UTM content',
    category: 'session',
    type: 'string',
    dimension: 'Session.utm_content'
  },
  // 'Session.landing_page': {
  //   key: 'Session.landing_page',
  //   title: 'Landing page',
  //   category: 'session',
  //   type: 'string',
  //   dimension: 'Session.landing_page'
  // },
  'Session.landing_page_path': {
    key: 'Session.landing_page_path',
    title: 'Landing page path',
    category: 'session',
    type: 'string',
    dimension: 'Session.landing_page_path'
  },
  // 'Session.referrer': {
  //   key: 'Session.referrer',
  //   title: 'Referrer',
  //   category: 'session',
  //   type: 'string',
  //   dimension: 'Session.referrer'
  // },
  'Session.referrer_domain': {
    key: 'Session.referrer_domain',
    title: 'Referrer domain',
    category: 'session',
    type: 'string',
    dimension: 'Session.referrer_domain'
  },
  'Session.referrer_path': {
    key: 'Session.referrer_path',
    title: 'Referrer path',
    category: 'session',
    type: 'string',
    dimension: 'Session.referrer_path'
  },
  'Session.is_first_conversion': {
    key: 'Session.is_first_conversion',
    title: 'First conversion',
    category: 'session',
    type: 'boolean',
    dimension: 'Session.is_first_conversion'
  },
  'Session.bounced': {
    key: 'Session.bounced',
    title: 'Bounced',
    category: 'session',
    type: 'boolean',
    dimension: 'Session.bounced'
  },
  'Session.role': {
    key: 'Session.role',
    title: 'Role',
    category: 'session',
    type: 'string',
    dimension: 'Session.role'
  }
}
