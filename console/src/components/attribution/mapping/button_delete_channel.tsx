import { useState } from 'react'
import { Button, message, Popconfirm, Tooltip } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faTrashCan } from '@fortawesome/free-regular-svg-icons'
import { SizeType } from 'antd/lib/config-provider/SizeContext'

type Props = {
  channelId: string
  workspaceId: string
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
  btnSize?: SizeType
  btnType?: 'link' | 'text' | 'ghost' | 'default' | 'primary' | 'dashed' | undefined
}

const DeleteChannelButton = (props: Props) => {
  const [loading, setLoading] = useState(false)

  const onConfirm = () => {
    setLoading(true)

    props
      .apiPOST('/channel.delete', {
        workspace_id: props.workspaceId,
        id: props.channelId
      })
      .then(() => {
        setLoading(false)
        message.success('This channel has been deleted!')
        props.onComplete()
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  return (
    <Popconfirm
      okText="Delete channel"
      okButtonProps={{ danger: true }}
      placement="topRight"
      title="Would you like to delete this channel?"
      onConfirm={onConfirm}
      disabled={loading}
    >
      <Tooltip title="Delete channel" placement="bottom">
        <Button type={props.btnType} size={props.btnSize} loading={loading}>
          <FontAwesomeIcon icon={faTrashCan} />
        </Button>
      </Tooltip>
    </Popconfirm>
  )
}

export default DeleteChannelButton
