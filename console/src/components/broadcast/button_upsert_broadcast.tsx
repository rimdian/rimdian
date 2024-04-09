import { Button, Col, DatePicker, Drawer, Form, Input, Radio, Row, Space, message } from 'antd'
import { useState } from 'react'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { BroadcastCampaign } from 'interfaces'
import { kebabCase } from 'lodash'
import Messages from 'utils/formMessages'
import extractTLD from 'utils/tld'
import InputSubscriptionLists from './input_subscription_lists'
import InputCampaignEmailTemplates from './input_campaign_email_templates'

interface ButtonUpsertCampaignProps {
  campaign?: BroadcastCampaign
  onSuccess?: () => void
  btnProps: any
  children: React.ReactNode
}

const ButtonUpsertCampaign = (props: ButtonUpsertCampaignProps) => {
  const [drawserVisible, setDrawserVisible] = useState(false)
  return (
    <>
      {drawserVisible && (
        <DrawerCampaign
          broadcastCampaign={props.campaign}
          setDrawserVisible={setDrawserVisible}
          onSuccess={props.onSuccess}
        />
      )}
      <Button {...props.btnProps} onClick={() => setDrawserVisible(true)}>
        {props.children}
      </Button>
    </>
  )
}

const DrawerCampaign = (props: {
  broadcastCampaign?: BroadcastCampaign
  setDrawserVisible: any
  onSuccess?: () => void
}) => {
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()
  const workspaceCtx = useCurrentWorkspaceCtx()

  const submitForm = (values: any) => {
    if (loading) return
    setLoading(true)

    const data = { ...values }
    data.workspace_id = workspaceCtx.workspace.id
    // data.channel = 'email'

    if (props.broadcastCampaign) {
      data.id = props.broadcastCampaign.id
    }

    workspaceCtx
      .apiPOST('/broadcastCampaign.upsert', data)
      .then(() => {
        message.success('The campaign has been saved!')
        setLoading(false)
        props.onSuccess && props.onSuccess()
        props.setDrawserVisible(false)
      })
      .catch(() => {
        setLoading(false)
      })
  }

  const initialValues = Object.assign(
    {
      channel: 'email',
      utm_source: extractTLD(workspaceCtx.workspace.website_url),
      utm_medium: 'email'
    },
    props.broadcastCampaign
  )

  return (
    <Drawer
      title={<>{props.broadcastCampaign ? 'Edit campaign' : 'Create a new campaign'}</>}
      closable={true}
      keyboard={false}
      maskClosable={false}
      width={700}
      open={true}
      onClose={() => props.setDrawserVisible(false)}
      extra={
        <div style={{ textAlign: 'right' }}>
          <Space>
            <Button type="link" loading={loading} onClick={() => props.setDrawserVisible(false)}>
              Cancel
            </Button>

            <Button
              loading={loading}
              onClick={() => {
                form.validateFields().then((values: any) => {
                  submitForm(values)
                })
              }}
              type="primary"
            >
              Save
            </Button>
          </Space>
        </div>
      }
    >
      <Form form={form} layout="vertical" initialValues={initialValues}>
        <Row gutter={24}>
          <Col span={12}>
            <Form.Item name="name" label="Campaign name" rules={[{ required: true }]}>
              <Input
                placeholder="i.e: Newsletter ABC"
                onChange={(e: any) => {
                  if (!props.broadcastCampaign) {
                    const id = kebabCase(e.target.value)
                    form.setFieldsValue({ id: id })
                  }
                }}
              />
            </Form.Item>

            {/* utm_source */}
            <Form.Item
              name="utm_source"
              label="Source (utm_source)"
              rules={[{ required: true, type: 'string' }]}
            >
              <Input placeholder="i.e: business.com" />
            </Form.Item>

            <Form.Item name="channel" label="Channel" rules={[{ required: true, type: 'string' }]}>
              <Radio.Group style={{ width: '100%' }}>
                <Radio.Button value="email" style={{ width: '50%', textAlign: 'center' }}>
                  Email
                </Radio.Button>
                <Radio.Button value="sms" disabled style={{ width: '50%', textAlign: 'center' }}>
                  SMS (soon)
                </Radio.Button>
              </Radio.Group>
            </Form.Item>
          </Col>

          <Col span={12}>
            <Form.Item
              name="id"
              label="Campaign ID (utm_campaign)"
              rules={[
                {
                  required: true,
                  type: 'string',
                  pattern: /^[a-z0-9]+(-[a-z0-9]+)*$/,
                  message: Messages.InvalidIdFormat
                }
              ]}
            >
              <Input
                disabled={props.broadcastCampaign ? true : false}
                placeholder="i.e: newsletter-abc"
              />
            </Form.Item>

            {/* medium */}
            <Form.Item
              name="utm_medium"
              label="Medium (medium)"
              rules={[{ required: true, type: 'string' }]}
            >
              <Input placeholder="i.e: email" />
            </Form.Item>

            {/* scheduled at */}
            <Form.Item
              name="scheduled_at"
              label="Scheduled at"
              rules={[{ required: true, type: 'date' }]}
            >
              <DatePicker showTime style={{ width: '100%' }} />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item noStyle dependencies={['channel']}>
          {() => {
            const channel = form.getFieldValue('channel')
            return (
              <>
                <Form.Item
                  name="subscription_lists"
                  label="Subscription lists (recipients)"
                  rules={[{ required: true, type: 'array', min: 1 }]}
                >
                  <InputSubscriptionLists channel={channel} />
                </Form.Item>

                {channel === 'email' && (
                  <Form.Item
                    name="message_templates"
                    label="Message templates"
                    rules={[
                      {
                        required: true,
                        type: 'array',
                        min: 1,
                        validator: (_rule, value, callback) => {
                          // total of object.percentage should be 100
                          const totalPercentage = value.reduce(
                            (acc: number, x: any) => acc + x.percentage,
                            0
                          )
                          if (totalPercentage !== 100) {
                            callback('The total of percentage should be 100')
                          }

                          // an object.percentage cannot be 0
                          const hasZeroPercentage = value.some((x: any) => x.percentage === 0)
                          if (hasZeroPercentage) {
                            callback('A percentage cannot be 0')
                          }
                          callback()
                        }
                      }
                    ]}
                  >
                    <InputCampaignEmailTemplates />
                  </Form.Item>
                )}
              </>
            )
          }}
        </Form.Item>
      </Form>
    </Drawer>
  )
}

export default ButtonUpsertCampaign
