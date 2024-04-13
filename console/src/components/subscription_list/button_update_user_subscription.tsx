import { Badge, Button, Form, Input, Modal, Radio, message } from 'antd'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import {
  SubscriptionList,
  SubscriptionListUser,
  SubscriptionListUserActive,
  SubscriptionListUserPaused,
  SubscriptionListUserUnsubscribed,
  User
} from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'

const ButtonUpdateUserSubscription = (props: {
  user: User
  subscriptionList: SubscriptionList
  workspaceCtx: CurrentWorkspaceCtxValue
  subscription?: SubscriptionListUser
  onSuccess?: () => void
}) => {
  const [modalVisible, setModalVisible] = useState(false)
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  return (
    <>
      <Button type="link" size="small" onClick={() => setModalVisible(true)}>
        update
      </Button>
      {modalVisible && (
        <Modal
          title="Update user subscription"
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
                form
                  .validateFields()
                  .then((values) => {
                    if (loading) return

                    // abort if no changes
                    if (props.subscription && values.status === props.subscription.status) {
                      return
                    }

                    setLoading(true)

                    const now = new Date().toISOString()
                    const item = {
                      kind: 'subscription_list_user',
                      subscription_list_user: {
                        subscription_list_id: props.subscriptionList.id,
                        status: values.status,
                        comment: values.comment || null,
                        created_at: props.subscription?.created_at || now,
                        updated_at: now
                      },
                      user: {
                        external_id: props.user.external_id,
                        is_authenticated: props.user.is_authenticated,
                        created_at: props.user.created_at
                      }
                    } as any

                    props.workspaceCtx
                      .collectorPOST(true, {
                        workspace_id: props.workspaceCtx.workspace.id,
                        items: [item]
                      })
                      .then(() => {
                        message.success('Subscription updated')
                        setLoading(false)
                        setModalVisible(false)
                        props.onSuccess && props.onSuccess()
                      })
                      .finally(() => {
                        setLoading(false)
                      })
                  })
                  .catch((info) => {
                    console.log('Validate Failed:', info)
                  })
              }}
            >
              Update subscription
            </Button>
          ]}
          onCancel={() => {
            setModalVisible(false)
          }}
        >
          <div className={CSS.margin_v_xl}>
            <Form
              form={form}
              layout="vertical"
              initialValues={{
                status: props.subscription?.status
              }}
            >
              {/* status select */}
              <Form.Item
                label="Status"
                name="status"
                rules={[{ required: true, type: 'number', message: 'Please select a status' }]}
              >
                <Radio.Group
                  optionType="button"
                  style={{ width: '100%' }}
                  onChange={(e) => {
                    if (e.target.value === 2) {
                      form.setFieldValue('comment', 'paused by admin')
                    }
                    if (e.target.value === 3) {
                      form.setFieldValue('comment', 'unsubscribed by admin')
                    }
                    if (e.target.value === 1) {
                      form.setFieldValue('comment', undefined)
                    }
                  }}
                >
                  <Radio.Button
                    value={SubscriptionListUserActive}
                    style={{ width: '33%', textAlign: 'center' }}
                  >
                    <Badge status="success" text="Active" />
                  </Radio.Button>
                  <Radio.Button
                    value={SubscriptionListUserPaused}
                    style={{ width: '33%', textAlign: 'center' }}
                  >
                    <Badge status="warning" text="Paused" />
                  </Radio.Button>
                  <Radio.Button
                    value={SubscriptionListUserUnsubscribed}
                    style={{ width: '33%', textAlign: 'center' }}
                  >
                    <Badge status="default" text="Unsubscribed" />
                  </Radio.Button>
                </Radio.Group>
              </Form.Item>

              <Form.Item noStyle dependencies={['status']}>
                {() => {
                  return form.getFieldValue('status') ? (
                    <Form.Item name="comment" label="Comment">
                      <Input.TextArea rows={4} placeholder="Comment for the status change" />
                    </Form.Item>
                  ) : null
                }}
              </Form.Item>
            </Form>
          </div>
        </Modal>
      )}
    </>
  )
}

export default ButtonUpdateUserSubscription
