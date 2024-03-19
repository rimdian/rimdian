import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import CSS from 'utils/css'
import { FileManager } from './file_manager'

const RouteFiles = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  return (
    <Layout
      currentOrganization={workspaceCtx.organization}
      currentWorkspace={workspaceCtx.workspace}
    >
      <div className={CSS.top}>
        <h1>Files</h1>
      </div>

      <FileManager
        foldersTree={workspaceCtx.workspace.folders_tree}
        onError={console.error}
        height={500}
        acceptFileType="*"
        acceptItem={() => true}
        onSelect={() => {}}
      />
    </Layout>
  )
}

export default RouteFiles
