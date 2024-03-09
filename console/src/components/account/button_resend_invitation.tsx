import { useState } from 'react'
import { Button, message, Popconfirm, Tooltip } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faRotateRight } from '@fortawesome/free-solid-svg-icons'

type Props = {
  organizationId: string
  email: string
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
}

const ResendInvitationButton = (props: Props) => {
  const [loading, setLoading] = useState(false)

  const onConfirm = () => {
    setLoading(true)

    props
      .apiPOST('/organizationInvitation.create', {
        organization_id: props.organizationId,
        email: props.email
      })
      .then(() => {
        setLoading(false)
        message.success('This invitation has been resent!')
        props.onComplete()
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  return (
    <Popconfirm
      okText="Resend invitation"
      placement="topRight"
      title="Would you like to resend this invitation?"
      onConfirm={onConfirm}
      disabled={loading}
    >
      <Tooltip title="Resend invitation" placement="bottom">
        <Button type="text" size="small" loading={loading}>
          <FontAwesomeIcon icon={faRotateRight} />
        </Button>
      </Tooltip>
    </Popconfirm>
  )
}

export default ResendInvitationButton
