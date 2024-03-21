import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import CSS from 'utils/css'
import { FileManager } from './file_manager'
import ButtonFilesSettings from './button_settings'
import { Button } from 'antd'

const RouteFiles = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  return (
    <Layout
      currentOrganization={workspaceCtx.organization}
      currentWorkspace={workspaceCtx.workspace}
    >
      <div className={CSS.container}>
        <div className={CSS.top}>
          <h1>Files</h1>
          <div className={CSS.topSeparator}></div>
          <div>
            {workspaceCtx.workspace.files_settings.endpoint !== '' && (
              <ButtonFilesSettings>
                <Button type="primary" ghost>
                  Settings
                </Button>
              </ButtonFilesSettings>
            )}
          </div>
        </div>

        <FileManager
          onError={console.error}
          height={500}
          acceptFileType="*"
          acceptItem={() => true}
          onSelect={() => {}}
        />
      </div>
    </Layout>
  )
}

export default RouteFiles
