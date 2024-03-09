import { useState } from 'react'
import { Button, Spin, Input, Modal, Form, message } from 'antd'
import { ErrorNotOwner } from 'components/organization/route_dashboard'
import CSS from 'utils/css'

type Props = {
  organizationId: string
  isOrganizationOwner: boolean
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
}

const AccountsInviteButton = (props: Props) => {
  const [visible, setVisible] = useState(false)
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  const toggleModal = () => {
    setVisible(!visible)
  }

  const onSubmit = () => {
    form
      .validateFields()
      .then((values: any) => {
        form.resetFields()
        setLoading(true)

        props
          .apiPOST('/organizationInvitation.create', {
            organization_id: props.organizationId,
            email: values.email,
            // for now we give all permissions to the users
            workspaces_scopes: [
              {
                workspace_id: '*',
                scopes: ['*']
              }
            ]
          })
          .then(() => {
            setLoading(false)
            message.success('The invitation has been sent!')
            toggleModal()
            props.onComplete()
          })
          .catch((_) => {
            setLoading(false)
          })
      })
      .catch((_) => {})
  }

  return (
    <>
      <Button
        type="link"
        onClick={() => {
          if (props.isOrganizationOwner) {
            toggleModal()
            return
          }
          ErrorNotOwner()
        }}
        className={CSS.margin_l_m}
      >
        Invite a colleague
      </Button>
      {visible && (
        <Modal
          title="Invite a colleague"
          open={true}
          onCancel={toggleModal}
          footer={[
            <Button key="back" loading={loading} onClick={toggleModal}>
              Cancel
            </Button>,
            <Button
              key="submit"
              type="primary"
              disabled={!props.isOrganizationOwner}
              loading={loading}
              onClick={onSubmit}
            >
              Confirm
            </Button>
          ]}
        >
          <Spin tip="Loading..." spinning={loading}>
            <Form form={form} layout="vertical">
              <Form.Item
                name="email"
                label="Email address"
                hasFeedback={true}
                rules={[{ type: 'email', required: true }]}
              >
                <Input type="email" disabled={!props.isOrganizationOwner} />
              </Form.Item>
            </Form>
          </Spin>
        </Modal>
      )}
    </>
  )
}

export default AccountsInviteButton
