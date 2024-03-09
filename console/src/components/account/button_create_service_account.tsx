import { useState, useMemo } from 'react'
import { Button, Spin, Input, Modal, Form, message, Alert, Tooltip } from 'antd'
import { camelCase } from 'lodash'
import nanoid from 'utils/nanoid'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCopy } from '@fortawesome/free-regular-svg-icons'
import { ErrorNotOwner } from 'components/organization/route_dashboard'
import Messages from 'utils/formMessages'
import CSS from 'utils/css'

type Props = {
  organizationId: string
  isOrganizationOwner: boolean
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
}

const CreateServiceAccountButton = (props: Props) => {
  const [visible, setVisible] = useState(false)
  const [loading, setLoading] = useState(false)
  const [apiKey, setApiKey] = useState<string>(nanoid(32))
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
          .apiPOST('/organizationAccount.createServiceAccount', {
            organization_id: props.organizationId,
            name: values.name,
            email_id: values.email_id,
            password: values.password,
            // for now we give all permissions to the service account
            workspaces_scopes: [
              {
                workspace_id: '*',
                scopes: ['*']
              }
            ]
          })
          .then(() => {
            setLoading(false)
            message.success('The Service Account has been created!')
            toggleModal()
            setApiKey(nanoid(32)) // generate a new api key for subsequent usage
            props.onComplete()
          })
          .catch((_) => {
            setLoading(false)
          })
      })
      .catch((_) => {})
  }

  const apiDomain = useMemo(() => {
    const url = new URL(window.Config.API_ENDPOINT)
    return url.hostname
  }, [])

  return (
    <>
      <Button
        // size="small"
        type="link"
        onClick={() => {
          if (props.isOrganizationOwner) {
            toggleModal()
            return
          }
          ErrorNotOwner()
        }}
      >
        New Service Account
      </Button>
      {visible && (
        <Modal
          title="Create a Service Account"
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
            <Form form={form} initialValues={{ password: apiKey }} layout="vertical">
              <Form.Item
                name="name"
                label="Service Account name"
                rules={[{ required: true, message: Messages.RequiredField }]}
              >
                <Input
                  disabled={!props.isOrganizationOwner}
                  onChange={(e: any) => {
                    const name = e.target.value
                    if (name) {
                      let newId = camelCase(name).toLowerCase()
                      if (newId !== '') {
                        form.setFieldsValue({ emailId: newId })
                      }
                    }
                  }}
                />
              </Form.Item>

              <Form.Item
                name="email_id"
                label="Login"
                rules={[
                  {
                    type: 'string',
                    required: true,
                    pattern: /^[a-z0-9]+$/,
                    message: Messages.RequiredField
                  }
                ]}
                shouldUpdate
              >
                <Input
                  disabled={!props.isOrganizationOwner}
                  addonAfter={
                    <span className={CSS.font_size_xs}>
                      .{props.organizationId}@{apiDomain}
                    </span>
                  }
                />
              </Form.Item>

              <Form.Item
                name="password"
                label="Password"
                rules={[
                  {
                    type: 'string',
                    required: true,
                    min: 16,
                    message: Messages.ServiceAccountPasswordInvalidFormat
                  }
                ]}
              >
                <Input
                  disabled={!props.isOrganizationOwner}
                  suffix={
                    <CopyToClipboard
                      text={form.getFieldValue('password')}
                      onCopy={() => message.success('Password copied to clipboard.')}
                    >
                      <Tooltip title="Copy to clipboard">
                        <FontAwesomeIcon icon={faCopy} style={{ cursor: 'pointer' }} />
                      </Tooltip>
                    </CopyToClipboard>
                  }
                />
              </Form.Item>

              {/* <Form.Item
                name="workspaces_scopes"
                label="Workspaces scopes"
                rules={[
                  {
                    type: 'string',
                    required: true,
                    min: 16,
                    message: Messages.ServiceAccountPasswordInvalidFormat
                  }
                ]}
              >
              </Form.Item> */}

              <Alert
                type="warning"
                message="Save this password in a secure place, it cannot be changed or recovered if you loose it!"
                showIcon
              />
            </Form>
          </Spin>
        </Modal>
      )}
    </>
  )
}

export default CreateServiceAccountButton
