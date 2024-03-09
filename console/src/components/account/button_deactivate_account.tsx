import { useState } from 'react'
import { Button, message, Popconfirm, Tooltip } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faTrashCan } from '@fortawesome/free-regular-svg-icons'

type Props = {
  organizationId: string
  deactivateAccountId: string
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
}

const DeactivateAccountButton = (props: Props) => {
  const [loading, setLoading] = useState(false)

  const onConfirm = () => {
    setLoading(true)

    props
      .apiPOST('/organizationAccount.deactivate', {
        organization_id: props.organizationId,
        deactivate_account_id: props.deactivateAccountId
      })
      .then(() => {
        setLoading(false)
        message.success('This account access has been deactivated!')
        props.onComplete()
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  return (
    <Popconfirm
      okText="Deactivate access"
      okButtonProps={{ danger: true }}
      placement="topRight"
      title="Would you like to deactivate this account access?"
      onConfirm={onConfirm}
      disabled={loading}
    >
      <Tooltip title="Deactivate access" placement="bottom">
        <Button type="text" size="small" loading={loading}>
          <FontAwesomeIcon icon={faTrashCan} />
        </Button>
      </Tooltip>
    </Popconfirm>
  )
}

export default DeactivateAccountButton
