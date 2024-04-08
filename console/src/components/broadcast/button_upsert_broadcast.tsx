import { Button, Drawer, Form, Input, Space, Tag, message } from 'antd'
import { useState } from 'react'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import CSS from 'utils/css'
import { BroadcastCampaign } from 'interfaces'
import { kebabCase } from 'lodash'
import Messages from 'utils/formMessages'

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
      //   utm_source: extractTLD(workspaceCtx.workspace.website_url),
      //   utm_medium: 'email',
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
                  console.log('values', values)
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

        <Form.Item name="channel" label="Channel" rules={[{ required: true, type: 'string' }]}>
          TODO
        </Form.Item>
      </Form>
    </Drawer>
  )
}

export default ButtonUpsertCampaign
