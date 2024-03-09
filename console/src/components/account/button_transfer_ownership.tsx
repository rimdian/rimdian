import { useState } from 'react'
import { Button, message, Popconfirm, Tooltip } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCrown } from '@fortawesome/free-solid-svg-icons'

type Props = {
  organizationId: string
  toAccountId: string
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
}

const TransferOwnershipButton = (props: Props) => {
  const [loading, setLoading] = useState(false)

  const onConfirm = () => {
    setLoading(true)

    props
      .apiPOST('/organizationAccount.transferOwnership', {
        organization_id: props.organizationId,
        to_account_id: props.toAccountId
      })
      .then(() => {
        setLoading(false)
        message.success('Organization ownership has been transfered!')
        props.onComplete()
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  return (
    <Popconfirm
      okText="Transfer ownership"
      okButtonProps={{ danger: true }}
      placement="topRight"
      title="Would you like to transfer ownership to this account?"
      onConfirm={onConfirm}
      disabled={loading}
    >
      <Tooltip title="Transfer ownership" placement="bottom">
        <Button type="text" size="small" loading={loading}>
          <FontAwesomeIcon icon={faCrown} style={{ color: '#FFC107' }} />
        </Button>
      </Tooltip>
    </Popconfirm>
  )
}

export default TransferOwnershipButton
