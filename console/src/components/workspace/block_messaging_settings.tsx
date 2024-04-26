import { Form, Input, Select, Radio, Button, message, Modal, Alert, InputNumber } from 'antd'
import { CurrentWorkspaceCtxValue, useCurrentWorkspaceCtx } from './context_current_workspace'
import { useState } from 'react'
import Block from 'components/common/block'
import CSS from 'utils/css'
import { EmailProvider } from 'interfaces'

const BlockMessagingSettings = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  const formatProvider = (provider: EmailProvider) => {
    if (provider.provider === 'sparkpost') {
      return (
        <>
          <p>
            <b>Provider</b>: SparkPost
          </p>
          <p>
            <b>API endpoint</b>: {provider.host}
          </p>
        </>
      )
    }
    if (provider.provider === 'smtp') {
      return (
        <>
          <p>
            <b>Provider</b>: SMTP
          </p>
          <p>
            <b>Host</b>: {provider.host}
          </p>
          <p>
            <b>Port</b>: {provider.port}
          </p>
          <p>
            <b>Encryption</b>: {provider.encryption}
          </p>
        </>
      )
    }
    return null
  }

  return (
    <>
      <Block
        title="Transactional email provider"
        extra={
          <>
            {workspaceCtx.workspace.messaging_settings.transactional_email_provider && (
              <EmailProviderSettings
                btnProps={{ size: 'small', type: 'primary', ghost: true }}
                kind="transactional_email_provider"
                workspaceCtx={workspaceCtx}
              >
                <>Edit</>
              </EmailProviderSettings>
            )}
          </>
        }
      >
        <div className={CSS.padding_a_m}>
          {!workspaceCtx.workspace.messaging_settings.transactional_email_provider && (
            <Alert
              message={
                <>
                  No transactional email provider configured.{' '}
                  <EmailProviderSettings
                    btnProps={{ size: 'small', type: 'primary' }}
                    kind="transactional_email_provider"
                    workspaceCtx={workspaceCtx}
                  >
                    <>Setup now</>
                  </EmailProviderSettings>
                </>
              }
              type="warning"
            />
          )}
          {workspaceCtx.workspace.messaging_settings.transactional_email_provider &&
            formatProvider(workspaceCtx.workspace.messaging_settings.transactional_email_provider)}
        </div>
      </Block>

      <Block
        title="Marketing email provider"
        extra={
          <>
            {workspaceCtx.workspace.messaging_settings.transactional_email_provider && (
              <EmailProviderSettings
                btnProps={{ size: 'small', type: 'primary', ghost: true }}
                kind="transactional_email_provider"
                workspaceCtx={workspaceCtx}
              >
                <>Edit</>
              </EmailProviderSettings>
            )}
          </>
        }
      >
        <div className={CSS.padding_a_m}>
          {!workspaceCtx.workspace.messaging_settings.marketing_email_provider && (
            <Alert
              message={
                <>
                  No marketing email provider configured.{' '}
                  <EmailProviderSettings
                    btnProps={{ size: 'small', type: 'primary' }}
                    kind="marketing_email_provider"
                    workspaceCtx={workspaceCtx}
                  >
                    <>Setup now</>
                  </EmailProviderSettings>
                </>
              }
              type="warning"
            />
          )}
          {workspaceCtx.workspace.messaging_settings.marketing_email_provider &&
            formatProvider(workspaceCtx.workspace.messaging_settings.marketing_email_provider)}
        </div>
      </Block>
    </>
  )
}

export default BlockMessagingSettings

const EmailProviderSettings = ({
  btnProps,
  kind,
  workspaceCtx,
  children
}: {
  btnProps: any
  kind: string
  workspaceCtx: CurrentWorkspaceCtxValue
  children: JSX.Element
}) => {
  const [modalVisible, setModalVisible] = useState(false)
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  const onSubmit = (values: any) => {
    if (loading) return

    setLoading(true)

    workspaceCtx
      .apiPOST('/workspace.settings', {
        id: workspaceCtx.workspace.id,
        [kind]: values
      })
      .then(() => {
        workspaceCtx
          .refreshWorkspace()
          .then(() => {
            form.resetFields()
            message.success('The email provider settings have been updated!')
            setLoading(false)
            setModalVisible(false)
          })
          .catch((_) => {
            setLoading(false)
          })
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  const initialValues = Object.assign(
    {
      //   port: 587
    },
    kind === 'marketing_email_provider'
      ? workspaceCtx.workspace.messaging_settings.marketing_email_provider
      : workspaceCtx.workspace.messaging_settings.transactional_email_provider
  )

  //   console.log('kind', kind)
  //   console.log('initialValues', initialValues)

  return (
    <>
      <Button {...btnProps} onClick={() => setModalVisible(true)}>
        {children}
      </Button>

      {modalVisible && (
        <Modal
          title={`Setup ${
            kind === 'marketing_email_provider' ? 'marketing' : 'transactional'
          } email provider`}
          open={true}
          onCancel={() => setModalVisible(false)}
          okText="Save"
          onOk={() => {
            form.validateFields().then(onSubmit).catch(console.error)
          }}
        >
          <Form
            form={form}
            layout="vertical"
            initialValues={initialValues}
            className={CSS.margin_v_xl}
          >
            <Form.Item
              name="provider"
              label="Provider"
              rules={[{ required: true, type: 'string' }]}
            >
              <Select
                options={[
                  { label: 'SparkPost', value: 'sparkpost' },
                  { label: 'SMTP', value: 'smtp' }
                ]}
              />
            </Form.Item>

            <Form.Item noStyle dependencies={['provider']}>
              {({ getFieldValue }: any) => {
                if (getFieldValue('provider') === 'smtp') {
                  // host, port, username, password, encryption
                  return (
                    <>
                      <Form.Item
                        name="host"
                        label="Host"
                        rules={[{ required: true, type: 'string' }]}
                      >
                        <Input />
                      </Form.Item>
                      <Form.Item
                        name="port"
                        label="Port"
                        rules={[{ required: true, type: 'number' }]}
                      >
                        <InputNumber min={0} step={1} />
                      </Form.Item>
                      <Form.Item
                        name="username"
                        label="Username"
                        rules={[{ required: true, type: 'string' }]}
                      >
                        <Input />
                      </Form.Item>
                      <Form.Item
                        name="password"
                        label="Password"
                        rules={[{ required: true, type: 'string' }]}
                      >
                        <Input.Password />
                      </Form.Item>
                      <Form.Item
                        name="encryption"
                        label="Encryption"
                        rules={[{ required: true, type: 'string' }]}
                      >
                        <Radio.Group>
                          <Radio.Button value="none">None</Radio.Button>
                          <Radio.Button value="STARTTLS">STARTTLS</Radio.Button>
                          <Radio.Button value="TLS">TLS</Radio.Button>
                          <Radio.Button value="SSL">SSL</Radio.Button>
                        </Radio.Group>
                      </Form.Item>
                    </>
                  )
                }
                if (getFieldValue('provider') === 'sparkpost') {
                  return (
                    <>
                      <Form.Item
                        name="host"
                        label="API endpint"
                        rules={[{ required: true, type: 'string' }]}
                      >
                        <Select
                          options={[
                            { label: 'EU', value: 'https://api.eu.sparkpost.com' },
                            { label: 'US', value: 'https://api.sparkpost.com' }
                          ]}
                        />
                      </Form.Item>
                      <Form.Item
                        name="password"
                        label="API key"
                        rules={[
                          { required: true, type: 'string', message: 'Please enter your API key' }
                        ]}
                      >
                        <Input />
                      </Form.Item>
                    </>
                  )
                }
                return null
              }}
            </Form.Item>
          </Form>
        </Modal>
      )}
    </>
  )
}

export { EmailProviderSettings }
