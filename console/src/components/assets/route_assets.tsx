import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import CSS from 'utils/css'
import { FileManager } from './files/file_manager'
import { Tabs } from 'antd'
import { useNavigate, useSearchParams } from 'react-router-dom'
import ListTemplates from './message_template/block_list'

const RouteFiles = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [params] = useSearchParams()
  const navigate = useNavigate()

  const changeTab = (value: string) => {
    navigate(
      '/orgs/' +
        workspaceCtx.organization.id +
        '/workspaces/' +
        workspaceCtx.workspace.id +
        '/assets?tab=' +
        value
    )
  }

  const tab = params.get('tab') || 'files'

  return (
    <Layout
      currentOrganization={workspaceCtx.organization}
      currentWorkspace={workspaceCtx.workspace}
    >
      <div className={CSS.container}>
        <div className={CSS.top + ' ' + CSS.margin_l_l}>
          <h1>Assets</h1>
        </div>

        <Tabs
          activeKey={tab}
          onChange={changeTab}
          destroyInactiveTabPane={true}
          tabPosition="left"
          items={[
            {
              key: 'files',
              label: 'Files',
              children: (
                <FileManager
                  onError={console.error}
                  height={500}
                  acceptFileType="*"
                  acceptItem={() => true}
                  onSelect={() => {}}
                />
              )
            },
            {
              key: 'templates',
              label: 'Message templates',
              children: <ListTemplates />
            }
          ]}
        />
      </div>
    </Layout>
  )
}

export default RouteFiles
