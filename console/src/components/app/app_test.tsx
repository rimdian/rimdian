import { Button, message, Space } from 'antd'
import { App } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { useState } from 'react'

interface AppTestProps {
  app: App
}

const AppTest = (props: AppTestProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [loading, setLoading] = useState(false)
  // console.log('props.app.state', props.app.state)

  const activate = () => {
    workspaceCtx
      .apiPOST('/app.activate', {
        workspace_id: workspaceCtx.workspace.id,
        id: props.app.id
      })
      .then(() => {
        workspaceCtx.refetchApps().then(() => {
          message.success('Test app has been activated!')
          setLoading(false)
        })
      })
      .catch((err) => {
        console.error(err)
        message.error(err.message)
        setLoading(false)
      })
  }
  return (
    <div style={{ paddingTop: '150px' }}>
      {/* <div>Test app</div> */}
      <Space>
        {props.app.status === 'initializing' && (
          <Button type="primary" loading={loading} onClick={activate}>
            Activate app
          </Button>
        )}
      </Space>
    </div>
  )
}

export default AppTest
