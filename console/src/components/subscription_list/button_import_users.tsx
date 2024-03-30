import { Alert, Button, ButtonProps, Form, Modal, Select, Tag } from 'antd'
import { Segment } from 'components/segment/interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { SubscriptionList } from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'

const ButtonImportSubscriptionListUsers = (props: {
  btnProps?: ButtonProps
  onSuccess?: () => void
  subscriptionList: SubscriptionList
  segments: Segment[]
}) => {
  const { btnProps, onSuccess } = props
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [modalVisible, setModalVisible] = useState(false)
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  return (
    <>
      <Button {...btnProps} onClick={() => setModalVisible(true)}>
        Import users from segment
      </Button>
      {modalVisible && (
        <Modal
          title="Import users to subscription list"
          open={modalVisible}
          footer={[
            <Button loading={loading} key="cancel" onClick={() => setModalVisible(false)}>
              Cancel
            </Button>,
            <Button
              key="submit"
              type="primary"
              loading={loading}
              onClick={() => {
                form.validateFields().then((values) => {
                  setLoading(true)
                  workspaceCtx
                    .apiPOST('/task.run', {
                      id: 'system_import_users_to_subscription_list',
                      workspace_id: workspaceCtx.workspace.id,
                      main_worker_state: {
                        subscription_list_id: props.subscriptionList.id,
                        source: 'segment',
                        segment_id: values.segment_id
                      },
                      multiple_exec_key: props.subscriptionList.id
                    })
                    .then(() => {
                      setLoading(false)
                      setModalVisible(false)
                      onSuccess && onSuccess()
                    })
                    .catch(() => {
                      setLoading(false)
                    })
                })
              }}
            >
              Import users
            </Button>
          ]}
          onCancel={() => {
            setModalVisible(false)
          }}
        >
          <Form form={form} layout="vertical" className={CSS.margin_v_xl}>
            <Alert
              message={
                <>
                  This action will create a background task to import all users from the selected
                  segment to the subscription list: <b>{props.subscriptionList.name}</b>
                </>
              }
              type="info"
              className={CSS.margin_b_xl}
            />

            <Form.Item name="segment_id" label="User segment" rules={[{ required: true }]}>
              <Select
                options={props.segments
                  .filter((segment) => segment.id !== 'anonymous')
                  .map((segment: Segment) => {
                    return {
                      label: <Tag color={segment.color}>{segment.name}</Tag>,
                      value: segment.id
                    }
                  })}
              />
            </Form.Item>
          </Form>
        </Modal>
      )}
    </>
  )
}

export default ButtonImportSubscriptionListUsers
