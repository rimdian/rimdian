import { useState } from 'react'
import { Button, message, Popconfirm, Tooltip } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faTrashCan } from '@fortawesome/free-regular-svg-icons'

type Props = {
  organizationId: string
  email: string
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
}

const CancelInvitationButton = (props: Props) => {
  const [loading, setLoading] = useState(false)

  const onConfirm = () => {
    setLoading(true)

    props
      .apiPOST('/organizationInvitation.cancel', {
        organization_id: props.organizationId,
        email: props.email
      })
      .then(() => {
        setLoading(false)
        message.success('This invitation has been cancelled!')
        props.onComplete()
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  return (
    <Popconfirm
      okText="Cancel invitation"
      okButtonProps={{ danger: true }}
      placement="topRight"
      title="Would you like to cancel this invitation?"
      onConfirm={onConfirm}
      disabled={loading}
    >
      <Tooltip title="Cancel invitation" placement="bottom">
        <Button type="text" size="small" loading={loading}>
          <FontAwesomeIcon icon={faTrashCan} />
        </Button>
      </Tooltip>
    </Popconfirm>
  )
}

export default CancelInvitationButton
