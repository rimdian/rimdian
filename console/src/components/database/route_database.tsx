import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { TableInformationSchema } from './schemas'
import { useQuery } from '@tanstack/react-query'
import { useNavigate, useSearchParams } from 'react-router-dom'
import CSS from 'utils/css'
import Layout from 'components/common/layout'
import BlockDBSchemas from './block_list'
import { Tabs } from 'antd'
import BlockDBDiagram from './block_diagram'

const RouteDatabase = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [params] = useSearchParams()
  const navigate = useNavigate()

  const changeTab = (value: string) => {
    navigate(
      '/orgs/' +
        workspaceCtx.organization.id +
        '/workspaces/' +
        workspaceCtx.workspace.id +
        '/database?tab=' +
        value
    )
  }

  const tab = params.get('tab') || 'tables'

  const { isLoading, data, isFetching } = useQuery<TableInformationSchema[]>(
    ['dbSchema', workspaceCtx.workspace.id],
    (): Promise<TableInformationSchema[]> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET('/workspace.showTables?workspace_id=' + workspaceCtx.workspace.id)
          .then((data: any) => {
            resolve(data.tables as TableInformationSchema[])
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  console.log(data)

  return (
    <Layout
      currentOrganization={workspaceCtx.organization}
      currentWorkspace={workspaceCtx.workspace}
    >
      <div className={CSS.top}>
        <h1>Database</h1>
      </div>
      <Tabs
        activeKey={tab}
        onChange={changeTab}
        destroyInactiveTabPane={true}
        items={[
          {
            key: 'tables',
            label: 'Tables',
            children: <BlockDBSchemas data={data} isLoading={isLoading} isFetching={isFetching} />
          },
          {
            key: 'diagram',
            label: 'Diagram',
            children: <BlockDBDiagram data={data} isLoading={isLoading} isFetching={isFetching} />
          }
        ]}
      />
    </Layout>
  )
}

export default RouteDatabase
