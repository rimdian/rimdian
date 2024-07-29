import { Alert, Tabs } from 'antd'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import DateRangeSelector from 'components/common/partial_date_range'
import CSS from 'utils/css'
import TabAttributionSessions from './tab_sessions'
// import TabAttributionPostviews from './tab_postviews'
import TabAttributionNotMapped from './tab_not_mapped'
import { useSearchParams } from 'react-router-dom'
import TabAttributionCrossChannels from './tab_cross_channels'
import TabAttributionCrossDevices from './tab_cross_devices'
import TabAttributionCrossDomains from './tab_cross_domains'
import TabTrafficMapping from './mapping/tab_mapping'
import ButtonReattributeConversions from './mapping/button_reattribute_conversions'

const RouteAttribution = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [searchParams, setSearchParams] = useSearchParams()

  const activeKey = searchParams.get('tab') || 'sessions'

  return (
    <Layout currentOrganization={workspaceCtx.organization} currentWorkspaceCtx={workspaceCtx}>
      <div className={CSS.top}>
        <h1>Attribution</h1>
        <div className={CSS.topSeparator}></div>
        {activeKey !== 'mapping' && <DateRangeSelector />}
      </div>
      <div>
        {workspaceCtx.workspace.outdated_conversions_attribution && (
          <>
            <Alert
              className={CSS.margin_b_m}
              description={
                <>
                  <p>
                    The mapping has been updated, you need to re-attribute your conversions for it
                    to take effect.
                  </p>

                  <ButtonReattributeConversions btnType="primary" btnSize="small" />
                </>
              }
              type="info"
            />
          </>
        )}
      </div>
      <Tabs
        activeKey={activeKey}
        destroyInactiveTabPane={true}
        onChange={(key) => {
          setSearchParams({ tab: key })
        }}
        items={[
          {
            key: 'sessions',
            label: 'Sessions',
            children: <TabAttributionSessions />
          },
          {
            key: 'not-mapped',
            label: 'Sessions not-mapped',
            children: <TabAttributionNotMapped />
          },
          // {
          //   key: 'postviews',
          //   label: 'Postviews',
          //   children: <TabAttributionPostviews />
          // },
          {
            key: 'cross-channels',
            label: 'Cross-channels',
            children: <TabAttributionCrossChannels />
          },
          {
            key: 'cross-devices',
            label: 'Cross-devices',
            children: <TabAttributionCrossDevices />
          },
          {
            key: 'cross-domains',
            label: 'Cross-domains',
            children: <TabAttributionCrossDomains />
          },
          {
            key: 'mapping',
            label: 'Mapping',
            children: <TabTrafficMapping />
          }
        ]}
      />
    </Layout>
  )
}

export default RouteAttribution
