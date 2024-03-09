import { Form, Input, Select, Divider, Radio, Col, Row, Button, message, Modal } from 'antd'
import {
  WorkspaceUserIdSigningAll,
  WorkspaceUserIdSigningAuthenticated,
  WorkspaceUserIdSigningNone
} from 'interfaces'
import { useCurrentWorkspaceCtx } from './context_current_workspace'
import { Languages } from 'utils/languages'
import { CountriesFormOptions, Timezones } from 'utils/countries_timezones'
import Messages from 'utils/formMessages'
import { useState } from 'react'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faKey } from '@fortawesome/free-solid-svg-icons'
import { Currencies, Currency } from 'utils/currencies'
import industries from 'utils/industries'
import Block from 'components/common/block'
import CSS from 'utils/css'

const BlockWorkspaceSettings = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  const [loading, setLoading] = useState(false)
  const [secretLoading, setSecretLoading] = useState(false)
  const [form] = Form.useForm()
  const [modal, contextHolder] = Modal.useModal()

  const onFinish = () => {
    form
      .validateFields()
      .then((values: any) => {
        if (loading) return

        values.session_timeout = parseInt(values.session_timeout, 10)
        values.data_retention = parseInt(values.data_retention, 10)

        setLoading(true)

        workspaceCtx
          .apiPOST('/workspace.update', values)
          .then((_res) => {
            workspaceCtx
              .refreshWorkspace()
              .then(() => {
                message.success('The workspace settings have been updated!')
                setLoading(false)
              })
              .catch((_) => {
                setLoading(false)
              })
          })
          .catch((_) => {
            setLoading(false)
          })
      })
      .catch(console.error)
  }

  const viewSecretKey = () => {
    if (secretLoading) return
    setSecretLoading(true)

    workspaceCtx
      .apiPOST('/workspace.getSecretKey', { id: workspaceCtx.workspace.id })
      .then((res) => {
        modal.info({
          open: true,
          title: 'Secret key',
          content: (
            <>
              <p>
                Use the following secret key to compute the user ID HMAC256, required to sign data
                collected by the JS SDK:
              </p>
              <p>
                <b>{res.secret_key}</b>
              </p>
            </>
          )
        })
        setSecretLoading(false)
      })
      .catch((_) => {
        setSecretLoading(false)
      })
  }

  return (
    <>
      {contextHolder}
      <Block classNames={[CSS.padding_t_xl, CSS.padding_b_l]}>
        <Form
          form={form}
          labelCol={{ span: 8 }}
          wrapperCol={{ span: 14 }}
          initialValues={workspaceCtx.workspace}
          layout="horizontal"
          onFinish={onFinish}
        >
          <Form.Item
            name="id"
            label="Workspace ID"
            rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
          >
            <Input disabled />
          </Form.Item>

          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
          >
            <Input />
          </Form.Item>

          <Form.Item
            name="website_url"
            label="Website URL"
            rules={[{ required: true, type: 'url', message: Messages.ValidURLRequired }]}
          >
            <Input />
          </Form.Item>

          <Form.Item
            name="privacy_policy_url"
            label="URL of your Privacy Policy"
            extra="Your workspace might be suspended if your Privacy Policy doesn't comply with global regulations (i.e: GDPR...)."
            rules={[{ required: true, type: 'url', message: Messages.ValidURLRequired }]}
          >
            <Input />
          </Form.Item>

          <Form.Item
            name="industry"
            label="Industry"
            rules={[{ required: true, message: Messages.RequiredField }]}
          >
            <Select options={industries} />
          </Form.Item>

          <Divider plain className={CSS.padding_v_xl}>
            General settings
          </Divider>

          <Form.Item
            name="currency"
            label="Main currency"
            rules={[{ required: true, message: Messages.RequiredField }]}
          >
            <Select
              showSearch
              placeholder="Select a currency"
              optionFilterProp="children"
              filterOption={(input: any, option: any) =>
                option.value.toLowerCase().includes(input.toLowerCase())
              }
              options={Currencies.map((c: Currency) => {
                return { value: c.code, label: c.code + ' - ' + c.currency }
              })}
            />
          </Form.Item>

          <Form.Item
            name="default_user_timezone"
            label="Default user timezone"
            rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
          >
            <Select
              placeholder="Select a time zone"
              allowClear={false}
              showSearch={true}
              filterOption={(searchText: any, option: any) => {
                return (
                  searchText !== '' && option.name.toLowerCase().includes(searchText.toLowerCase())
                )
              }}
              options={Timezones}
              fieldNames={{
                label: 'name',
                value: 'name'
              }}
            />
          </Form.Item>

          <Form.Item
            name="default_user_language"
            label="Default user language"
            rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
          >
            <Select
              placeholder="Select a value"
              allowClear={false}
              showSearch={true}
              filterOption={(searchText: any, option: any) => {
                return (
                  searchText !== '' && option.name.toLowerCase().includes(searchText.toLowerCase())
                )
              }}
              options={Languages}
            />
          </Form.Item>

          <Form.Item
            name="default_user_country"
            label="Default user country"
            rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
          >
            <Select
              showSearch
              placeholder="Select a country"
              filterOption={(input: any, option: any) =>
                option.label.toLowerCase().includes(input.toLowerCase())
              }
              options={CountriesFormOptions}
            />
          </Form.Item>

          <Form.Item
            name="user_reconciliation_keys"
            label="User reconciliation keys"
            extra="Columns used to reconciliate &amp; merge user identities."
            rules={[{ required: true, type: 'array', min: 1, message: Messages.RequiredField }]}
          >
            <Select
              showSearch
              placeholder="Select or enter user fields"
              filterOption={(input: any, option: any) =>
                option.label.toLowerCase().includes(input.toLowerCase())
              }
              options={[
                { value: 'email', label: 'email' },
                { value: 'email_md5', label: 'email_md5' },
                { value: 'email_sha1', label: 'email_sha1' },
                { value: 'email_sha256', label: 'email_sha256' },
                { value: 'telephone', label: 'telephone' }
              ]}
              mode="tags"
            />
          </Form.Item>

          <Divider plain className={CSS.padding_v_xl}>
            Web tracking settings
          </Divider>

          <Form.Item
            name="session_timeout"
            label="Session timeout"
            extra="Everytime you have a new visit on your website/app, a new session starts. By default, a session expires after 30 minutes (= 1800 secs) of inactivity, or if the visitor comes back from another source of traffic."
            rules={[
              {
                required: true,
                type: 'integer',
                transform: (value: any) => parseInt(value, 10),
                message: Messages.RequiredField
              }
            ]}
          >
            <Input type="number" placeholder="1800" addonAfter="secs" style={{ width: '150px' }} />
          </Form.Item>

          <Form.Item
            name="user_id_signing"
            label="Secure web hits"
            help="Sign the user IDs (with HMAC256) sent by the web agent to avoid malicious data corruption."
            rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
          >
            <Radio.Group style={{ width: '100%' }}>
              <Radio.Button
                style={{ width: '33.33%', textAlign: 'center' }}
                value={WorkspaceUserIdSigningNone}
              >
                None
              </Radio.Button>
              <Radio.Button
                style={{ width: '33.33%', textAlign: 'center' }}
                value={WorkspaceUserIdSigningAuthenticated}
              >
                Authenticated users
              </Radio.Button>
              <Radio.Button
                style={{ width: '33.33%', textAlign: 'center' }}
                value={WorkspaceUserIdSigningAll}
              >
                All users
              </Radio.Button>
            </Radio.Group>
          </Form.Item>

          <Form.Item noStyle shouldUpdate>
            {(funcs) => {
              if (funcs.getFieldValue('user_id_signing') !== 'none') {
                return (
                  <Form.Item label="Secret key" className={CSS.margin_t_m}>
                    <Button loading={secretLoading} onClick={viewSecretKey} type="primary" ghost>
                      <FontAwesomeIcon icon={faKey} />
                      &nbsp; View secret key
                    </Button>
                  </Form.Item>
                )
              }
            }}
          </Form.Item>

          <Divider plain className={CSS.padding_v_xl}>
            License
          </Divider>

          <Form.Item
            label="License key"
            help="Without license, the Community Edition quotas & restrictions will apply."
            name="license_key"
          >
            <Input />
          </Form.Item>

          <Row>
            <Col xs={22} sm={22} className={CSS.text_right + ' ' + CSS.padding_t_l}>
              <Button type="primary" loading={loading} htmlType="submit">
                Save changes
              </Button>
            </Col>
          </Row>
        </Form>
      </Block>
    </>
  )
}

export default BlockWorkspaceSettings
