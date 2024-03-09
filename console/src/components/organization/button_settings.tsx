import { useState } from 'react'
import { Select, Drawer, Button, Input, Form, message, Tooltip, Space } from 'antd'
import { HandleAxiosError } from 'utils/errors'
import { Currencies } from 'utils/currencies'
import { Organization } from 'interfaces'
import { ErrorNotOwner } from 'components/organization/route_dashboard'
import Messages from 'utils/formMessages'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faEdit } from '@fortawesome/free-regular-svg-icons'

type ButtonOrganizationSettingsProps = {
  apiPOST: (endpoint: string, data: any) => Promise<any>
  organization: Organization
  organizations: Organization[]
  updateOrganization: (org: Organization) => void
}

const ButtonOrganizationSettings = (props: ButtonOrganizationSettingsProps) => {
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()
  const [settingsVisible, setSettingsVisible] = useState(false)

  const toggleSettings = () => {
    setSettingsVisible(!settingsVisible)
  }

  const onFinish = (values: any) => {
    setLoading(true)

    props
      .apiPOST('/organization.setProfile', {
        id: props.organization.id,
        name: values.name,
        currency: values.currency
      })
      .then((res) => {
        // console.log('res', res)
        setLoading(false)
        message.success('Your organization has been updated!')
        props.updateOrganization(res.organization)
        toggleSettings()
      })
      .catch((e) => {
        HandleAxiosError(e)
      })
  }

  const initialValues = {
    name: props.organization.name,
    currency: props.organization.currency
  }

  return (
    <>
      <Tooltip title="Edit settings">
        <Button
          type="text"
          onClick={() => {
            if (props.organization.im_owner) {
              toggleSettings()
              return
            }
            ErrorNotOwner()
          }}
        >
          <FontAwesomeIcon icon={faEdit} />
        </Button>
      </Tooltip>

      {settingsVisible && (
        <Drawer
          title="Organization settings"
          open={true}
          onClose={toggleSettings}
          width={600}
          extra={
            <Space>
              <Button loading={loading} onClick={toggleSettings}>
                Cancel
              </Button>
              <Button
                loading={loading}
                onClick={() => {
                  form
                    .validateFields()
                    .then(onFinish)
                    .catch(() => {})
                }}
                type="primary"
              >
                Save
              </Button>
            </Space>
          }
        >
          <Form form={form} initialValues={initialValues} onFinish={onFinish} layout="vertical">
            <Form.Item
              name="name"
              label="Name"
              rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
            >
              <Input />
            </Form.Item>

            <Form.Item
              name="currency"
              label="Currency"
              rules={[{ required: true, type: 'string', message: Messages.InvalidTimezone }]}
            >
              <Select
                placeholder="Select a currency"
                allowClear={false}
                showSearch={true}
                filterOption={(searchText: any, option: any) => {
                  return (
                    searchText !== '' &&
                    option.label.toLowerCase().includes(searchText.toLowerCase())
                  )
                }}
                options={Currencies}
              />
            </Form.Item>
          </Form>
        </Drawer>
      )}
    </>
  )
}

export default ButtonOrganizationSettings
