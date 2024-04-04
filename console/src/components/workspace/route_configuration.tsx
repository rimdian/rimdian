import { Tabs } from 'antd'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import { useNavigate, useSearchParams } from 'react-router-dom'
import BlockWorkspaceSettings from './block_settings'
import BlockDomains from 'components/domain/block_list'
import CSS from 'utils/css'
import BlockDataHooks from 'components/data_hooks/block_list'
import BlockMessagingSettings from './block_messaging_settings'

const RouteWorkspaceConfiguration = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [params] = useSearchParams()
  const navigate = useNavigate()

  const changeTab = (value: string) => {
    navigate(
      '/orgs/' +
        workspaceCtx.organization.id +
        '/workspaces/' +
        workspaceCtx.workspace.id +
        '/system/configuration?tab=' +
        value
    )
  }

  const tab = params.get('tab') || 'domains'

  return (
    <Layout
      currentOrganization={workspaceCtx.organization}
      currentWorkspace={workspaceCtx.workspace}
    >
      <div className={CSS.container}>
        <div className={CSS.top + ' ' + CSS.margin_l_l}>
          <h1>Configuration</h1>
        </div>

        <Tabs
          activeKey={tab}
          onChange={changeTab}
          destroyInactiveTabPane={true}
          tabPosition="left"
          items={[
            {
              key: 'domains',
              label: 'Domains',
              children: <BlockDomains />
            },
            {
              key: 'data-hooks',
              label: 'Data hooks',
              children: <BlockDataHooks />
            },
            {
              key: 'messaging-settings',
              label: 'Messaging settings',
              children: <BlockMessagingSettings />
            },
            {
              key: 'settings',
              label: 'Workspace settings',
              children: <BlockWorkspaceSettings />
            }
          ]}
        />
      </div>
    </Layout>
  )
}

export default RouteWorkspaceConfiguration
