import { useState } from 'react'
import { Button, message, Popconfirm } from 'antd'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'

type Props = {
  btnSize?: SizeType
  btnType?: 'link' | 'text' | 'default' | 'primary' | 'dashed' | undefined
}

const ButtonReattributeConversions = (props: Props) => {
  const [loading, setLoading] = useState(false)
  const workspaceCtx = useCurrentWorkspaceCtx()

  const onConfirm = () => {
    setLoading(true)

    workspaceCtx
      .apiPOST('/task.run', {
        id: 'system_reattribute_conversions',
        workspace_id: workspaceCtx.workspace.id
      })
      .then(() => {
        setLoading(false)
        // refresh workspace
        workspaceCtx.refreshWorkspace().then(() => {
          message.success('The re-attribution task has been launched!')
        })
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  return (
    <Popconfirm
      okText="Re-attribute conversions"
      okButtonProps={{}}
      //   placement="topRight"
      title="Would you like to re-attribute your conversions now?"
      onConfirm={onConfirm}
      disabled={loading}
    >
      <Button type={props.btnType} size={props.btnSize} loading={loading}>
        Re-attribute conversions
      </Button>
    </Popconfirm>
  )
}

export default ButtonReattributeConversions
