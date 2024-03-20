import { Alert, Button, Form, Input, Modal, message } from 'antd'
import CSS from 'utils/css'
import { useState } from 'react'
import { useForm } from 'antd/lib/form/Form'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'

const ButtonFilesSettings = (props: { children: JSX.Element }) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [loading, setLoading] = useState(false)
  const [form] = useForm()
  const [settingsVisible, setSettingsVisible] = useState(false)

  const toggleSettings = () => {
    setSettingsVisible(!settingsVisible)
  }

  const onFinish = () => {
    form
      .validateFields()
      .then((values: any) => {
        if (loading) return

        setLoading(true)

        const updatedWorkspace = Object.assign({}, workspaceCtx.workspace)
        updatedWorkspace.files_settings = Object.assign(
          {},
          workspaceCtx.workspace.files_settings,
          values
        )

        workspaceCtx
          .apiPOST('/workspace.update', updatedWorkspace)
          .then(() => {
            workspaceCtx
              .refreshWorkspace()
              .then(() => {
                message.success('The workspace settings have been updated!')
                setLoading(false)
                toggleSettings()
              })
              .catch(() => {
                setLoading(false)
              })
          })
          .catch(() => {
            setLoading(false)
          })
      })
      .catch(console.error)
  }

  return (
    <span>
      <span onClick={toggleSettings}>{props.children}</span>
      <Modal
        title="File storage settings"
        open={settingsVisible}
        onCancel={toggleSettings}
        confirmLoading={loading}
        footer={[
          <Button key="cancel" onClick={toggleSettings}>
            Cancel
          </Button>,
          <Button key="submit" type="primary" onClick={onFinish}>
            Save
          </Button>
        ]}
      >
        <Form
          form={form}
          layout="horizontal"
          initialValues={workspaceCtx.workspace.files_settings}
          labelCol={{ span: 6 }}
          wrapperCol={{ span: 18 }}
          className={CSS.margin_t_xl + ' ' + CSS.margin_b_xxl}
          onFinish={onFinish}
        >
          <Alert
            message="Your files can be uploaded to any S3 compatible storage."
            type="info"
            showIcon
            className={CSS.margin_b_l}
          />

          <Form.Item
            label="S3 Endpoint"
            name="endpoint"
            rules={[{ type: 'string', required: true }]}
          >
            <Input placeholder="storage.googleapis.com" prefix="https://" />
          </Form.Item>
          <Form.Item
            label="S3 access key"
            name="access_key"
            rules={[{ type: 'string', required: true }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="S3 secret key"
            name="secret_key"
            rules={[{ type: 'string', required: true }]}
          >
            <Input />
          </Form.Item>
          <Form.Item label="S3 bucket" name="bucket" rules={[{ type: 'string', required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item
            label="S3 location"
            name="location"
            rules={[{ type: 'string', required: false }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="CDN endpoint"
            name="cdn_endpoint"
            help="URL of the CDN that caches your files"
            rules={[{ type: 'url', required: false }]}
          >
            <Input placeholder="https://cdn.yourbusiness.com" />
          </Form.Item>
        </Form>
      </Modal>
    </span>
  )
}

export default ButtonFilesSettings
