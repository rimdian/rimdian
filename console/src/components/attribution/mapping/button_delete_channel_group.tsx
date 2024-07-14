import { useState } from 'react'
import { Button, message, Popconfirm, Tooltip } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faTrashCan } from '@fortawesome/free-regular-svg-icons'
import { SizeType } from 'antd/lib/config-provider/SizeContext'

type Props = {
  channelGroupId: string
  workspaceId: string
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
  btnSize?: SizeType
  btnType?: 'text' | 'link' | 'default' | 'dashed' | 'primary' | undefined
}

const DeleteChannelGroupButton = (props: Props) => {
  const [loading, setLoading] = useState(false)

  const onConfirm = () => {
    setLoading(true)

    props
      .apiPOST('/channelGroup.delete', {
        workspace_id: props.workspaceId,
        id: props.channelGroupId
      })
      .then(() => {
        setLoading(false)
        message.success('This channel group has been deleted!')
        props.onComplete()
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  return (
    <Popconfirm
      okText="Delete channel group"
      okButtonProps={{ danger: true }}
      placement="topRight"
      title="Would you like to delete this channel group?"
      onConfirm={onConfirm}
      disabled={loading}
    >
      <Tooltip title="Delete channel group" placement="bottom">
        <Button size={props.btnSize} type={props.btnType} loading={loading}>
          <FontAwesomeIcon icon={faTrashCan} />
        </Button>
      </Tooltip>
    </Popconfirm>
  )
}

export default DeleteChannelGroupButton
