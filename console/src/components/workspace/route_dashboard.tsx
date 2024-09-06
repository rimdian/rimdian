import Layout from 'components/common/layout'
import { useCurrentWorkspaceCtx } from './context_current_workspace'
import { Row, Col } from 'antd'
import { useAccount } from 'components/login/context_account'
import { Account } from 'interfaces'
import { KPI } from 'components/common/partial_kpi'
import { UsersPerCountry } from './block_users_per_country'
import { UsersPerDevice } from './block_users_per_device'
import { OrdersPerDomain } from './block_orders_per_domain'
import { TrafficSources } from './block_traffic_sources'
import { UsersOnline } from './block_users_online'
import DateRangeSelector, {
  dateRangeValuesFromSearchParams,
  updateSearchParams,
  vsDateRangeValues
} from 'components/common/partial_date_range'
import { useMemo } from 'react'
import Block from 'components/common/block'
import CSS from 'utils/css'
import { useSearchParams } from 'react-router-dom'

const RouteWorkspaceDashboard = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const accountCtx = useAccount()
  const account = accountCtx.account?.account as Account
  const [searchParams, setSearchParams] = useSearchParams()

  const [dateFrom, dateTo] = dateRangeValuesFromSearchParams(searchParams)
  const [dateFromPrevious, dateToPrevious] = vsDateRangeValues(dateFrom, dateTo)
  const refreshKey = searchParams.get('refresh_key') || 'default'

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
    <Layout currentOrganization={workspaceCtx.organization} currentWorkspaceCtx={workspaceCtx}>
      <Row gutter={24} className={CSS.margin_t_l}>
        <Col span={12}>
          <UsersOnline
            refreshKey={refreshKey}
            workspaceId={workspaceCtx.workspace.id}
            timezone={account.timezone}
          />
        </Col>
        <Col span={12}>
          <div className={CSS.text_right}>
            <DateRangeSelector
              preset={'30D'}
              // timezone={accountCtx.account?.account.timezone || 'UTC'}
              onChange={(preset, range) => {
                updateSearchParams(searchParams, setSearchParams, preset, range)
              }}
            />
          </div>
        </Col>
      </Row>

      <Row gutter={24}>
        <Col span={12}>
          <Block classNames={[CSS.margin_t_l]} grid={true}>
            <KPI
              title="Users"
              tooltip="Total number of distinct users who performed a session"
              valueType="number"
              measure="Session.unique_users"
              timeDimension="Session.created_at_trunc"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
            <KPI
              title="Signed up"
              tooltip="Total number users who signed up"
              valueType="number"
              measure="User.count"
              timeDimension="User.signed_up_at"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
            <KPI
              title="Authenticated"
              tooltip="Total number of distinct authenticated users who performed a session"
              valueType="number"
              measure="Session.unique_users"
              timeDimension="Session.created_at_trunc"
              filters={[
                {
                  member: 'User.is_authenticated',
                  operator: 'equals',
                  values: ['1'] // use a string for boolean, otherwise it fails
                }
              ]}
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
          </Block>

          <Block classNames={[CSS.margin_t_l]} grid={true}>
            <KPI
              title="Sessions"
              valueType="number"
              measure="Session.count"
              timeDimension="Session.created_at_trunc"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
            <KPI
              title="Sessions bounce rate"
              tooltip="The percentage of sessions that had only one interaction (pageview...), or stayed less than 15 seconds on the pageview."
              valueType="percent"
              goodIsBad={true}
              measure="Session.bounce_rate"
              timeDimension="Session.created_at_trunc"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
            <KPI
              title="Avg. session duration"
              valueType="duration"
              measure="Session.avg_duration"
              timeDimension="Session.created_at_trunc"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
          </Block>

          <UsersPerCountry
            workspaceId={workspaceCtx.workspace.id}
            timezone={account.timezone}
            refreshKey={refreshKey}
            dateFrom={dateFrom}
            dateTo={dateTo}
            dateFromPrevious={dateFromPrevious}
            dateToPrevious={dateToPrevious}
          />

          <UsersPerDevice
            workspaceId={workspaceCtx.workspace.id}
            timezone={account.timezone}
            refreshKey={refreshKey}
            dateFrom={dateFrom}
            dateTo={dateTo}
            dateFromPrevious={dateFromPrevious}
            dateToPrevious={dateToPrevious}
          />
        </Col>

        <Col span={12}>
          <Block classNames={[CSS.margin_t_l]} grid={true}>
            <KPI
              title="Orders"
              valueType="number"
              measure="Order.count"
              timeDimension="Order.created_at_trunc"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
            <KPI
              title="Revenue"
              tooltip="Sum of orders subtotal"
              valueType="currency"
              currency={workspaceCtx.workspace.currency}
              measure="Order.subtotal_sum"
              timeDimension="Order.created_at_trunc"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
            <KPI
              title="Average cart"
              tooltip="Average of orders subtotal"
              valueType="currency"
              currency={workspaceCtx.workspace.currency}
              measure="Order.avg_cart"
              timeDimension="Order.created_at_trunc"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
          </Block>

          <Block classNames={[CSS.margin_t_l]} grid={true}>
            <KPI
              title="Online conversion rate"
              tooltip="Web orders / web sessions"
              valueType="percent"
              measure="Session.conversion_rate"
              filters={[
                {
                  member: 'Session.domain_id',
                  operator: 'equals',
                  values: webDomains
                }
              ]}
              timeDimension="Session.created_at_trunc"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
            <KPI
              title="Avg. time to conversion"
              tooltip="Average time to conversion for web orders"
              valueType="duration"
              measure="Order.avg_ttc"
              filters={[
                {
                  member: 'Order.domain_id',
                  operator: 'equals',
                  values: webDomains
                }
              ]}
              timeDimension="Order.created_at_trunc"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
            <KPI
              title="Abandoned carts"
              tooltip="Carts with status=0"
              valueType="number"
              measure="Cart.abandoned_count"
              timeDimension="Cart.created_at_trunc"
              color="purple"
              workspaceId={workspaceCtx.workspace.id}
              timezone={account.timezone}
              refreshKey={refreshKey}
              dateFrom={dateFrom}
              dateTo={dateTo}
              dateFromPrevious={dateFromPrevious}
              dateToPrevious={dateToPrevious}
            />
          </Block>

          <OrdersPerDomain
            workspaceId={workspaceCtx.workspace.id}
            domains={workspaceCtx.workspace.domains}
            currency={workspaceCtx.workspace.currency}
            timezone={account.timezone}
            refreshKey={refreshKey}
            dateFrom={dateFrom}
            dateTo={dateTo}
            dateFromPrevious={dateFromPrevious}
            dateToPrevious={dateToPrevious}
          />

          <TrafficSources
            workspaceId={workspaceCtx.workspace.id}
            currency={workspaceCtx.workspace.currency}
            timezone={account.timezone}
            refreshKey={refreshKey}
            dateFrom={dateFrom}
            dateTo={dateTo}
            dateFromPrevious={dateFromPrevious}
            dateToPrevious={dateToPrevious}
          />
        </Col>
      </Row>
    </Layout>
  )
}
export default RouteWorkspaceDashboard
