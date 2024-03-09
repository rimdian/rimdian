import { useState, useMemo } from 'react'
import { Button, Modal, Input, Select, message, Form } from 'antd'
import { Languages } from 'utils/languages'
import { useNavigate } from 'react-router-dom'
import { ErrorNotOwner } from 'components/organization/route_dashboard'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import InfoRadioGroup from 'components/common/input_info_radio_group'
import { CountriesFormOptions, Timezones } from 'utils/countries_timezones'
import { Currencies, Currency } from 'utils/currencies'
import industries from 'utils/industries'
import { WorkspaceKindDemoLead, WorkspaceKindDemoOrder, WorkspaceKindReal } from 'interfaces'
import Messages from 'utils/formMessages'
import dayjs from 'dayjs'
import { ButtonType } from 'antd/es/button'

type CreateWorkspaceButtonProps = {
  imOwner: boolean
  organizationId: string
  text: JSX.Element
  btnType: ButtonType
  btnSize: SizeType
  apiPOST: (endpoint: string, data: any) => Promise<any>
  refreshWorkspaces: () => Promise<any>
}

const CreateWorkspaceButton = (props: CreateWorkspaceButtonProps) => {
  const navigate = useNavigate()
  const [modalVisible, setModalVisible] = useState(false)
  const [workspaceType, setWorkspaceType] = useState<undefined | string>(undefined)
  const [loading, setLoading] = useState(false)
  const [createRealWorkspace, setCreateRealWorkspace] = useState(false)
  const [form] = Form.useForm()

  const toggleModal = () => {
    if (modalVisible) {
      if (createRealWorkspace) form.resetFields()
      setWorkspaceType(undefined)
      setCreateRealWorkspace(false)
      setLoading(false)
    }
    setModalVisible(!modalVisible)
  }

  const createDemo = () => {
    if (loading) return
    setLoading(true)

    props
      .apiPOST('/workspace.createOrResetDemo', {
        organization_id: props.organizationId,
        kind: workspaceType
      })
      .then((res) => {
        props
          .refreshWorkspaces()
          .then(() => {
            message.success('The demo workspace has been created!')
            toggleModal()
            setLoading(false)
            navigate('/orgs/' + props.organizationId + '/workspaces/' + res.id)
          })
          .catch((_) => {})
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  const onCreateWorkspace = () => {
    form
      .validateFields()
      .then((values: any) => {
        if (loading) return
        setLoading(true)

        const data = { ...values }
        data.id = data.id.toLowerCase().replace(/[^0-9a-z]/gi, '')
        data.organization_id = props.organizationId
        data.kind = WorkspaceKindReal

        props
          .apiPOST('/workspace.create', data)
          .then((res) => {
            // console.log('_res', res)
            // waiting for refreshWorkspaces() will trigger a parent rerender and might remove the
            // "create workspace button" when data arrives (=removing "empty-state-ui" components)
            // we should update this component state before refreshWorkspaces() resolves
            message.success('The workspace has been created!')
            toggleModal()
            props
              .refreshWorkspaces()
              .then(() => {})
              .catch((_) => {})
            navigate('/orgs/' + props.organizationId + '/workspaces/' + res.id)
          })
          .catch((_) => {
            setLoading(false)
          })
      })
      .catch(console.error)
  }

  const navigatorTimezone = useMemo(() => {
    return dayjs.tz.guess()
  }, [])

  const timezoneCountry = useMemo(() => {
    return Timezones.find((x) => x.name === navigatorTimezone)?.countries[0] || undefined
  }, [navigatorTimezone])

  const footer = [
    <Button key="back" onClick={toggleModal}>
      Cancel
    </Button>
  ]

  if (workspaceType === WorkspaceKindReal) {
    footer.push(
      <Button key="submit" type="primary" loading={loading} onClick={onCreateWorkspace}>
        Create workspace
      </Button>
    )
  }

  if (workspaceType === WorkspaceKindDemoOrder || workspaceType === WorkspaceKindDemoLead) {
    footer.push(
      <Button key="submit" type="primary" loading={loading} onClick={createDemo}>
        Create demo
      </Button>
    )
  }

  return (<>
    {modalVisible && (
      <Modal
        title="Create a new workspace"
        open={true}
        width={720}
        onCancel={toggleModal}
        footer={footer}
      >
        {!createRealWorkspace && (
          <InfoRadioGroup
            value={workspaceType}
            options={[
              {
                key: WorkspaceKindReal,
                title: <>Real workspace</>,
                content: <>Setup a real workspace for your business.</>
              },
              {
                key: WorkspaceKindDemoOrder,
                title: <>Demo: eCommerce</>,
                content: <>Generate a demo workspace with dummy eCommerce data.</>
              },
              {
                key: WorkspaceKindDemoLead,
                title: <>Demo: lead generation</>,
                content: <>Generate a demo workspace with dummy data for lead generation.</>,
                disabled: true
              }
            ]}
            onChange={(type: string) => {
              setWorkspaceType(type)
              if (type === WorkspaceKindReal) {
                setCreateRealWorkspace(true)
              }
            }}
            layout="horizontal"
          />
        )}

        {createRealWorkspace && (
          <Form
            form={form}
            labelCol={{ span: 10 }}
            wrapperCol={{ span: 12 }}
            initialValues={{
              default_user_country: timezoneCountry,
              default_user_timezone: navigatorTimezone,
              default_user_language: navigator.language.substring(0, 2)
            }}
            layout="horizontal"
          >
            <Form.Item
              name="name"
              label="Workspace name"
              rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
            >
              <Input
                onChange={(e: any) =>
                  form.setFieldsValue({
                    id: e.target.value.toLowerCase().replace(/[^0-9a-z]/gi, '')
                  })
                }
              />
            </Form.Item>

            <Form.Item
              name="id"
              label="Workspace ID"
              rules={[
                {
                  required: true,
                  type: 'string',
                  pattern: /^[a-z0-9]+$/,
                  message: Messages.InvalidWorkspaceIdFormat
                }
              ]}
            >
              <Input
                addonBefore={props.organizationId + '_'}
                onChange={(e) =>
                  form.setFieldsValue({
                    id: e.target.value.toLowerCase().replace(/[^0-9a-z]/gi, '')
                  })
                }
              />
            </Form.Item>

            <Form.Item
              name="website_url"
              initialValue="https://"
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
              name="industry"
              label="Industry"
              rules={[{ required: true, message: Messages.RequiredField }]}
            >
              <Select options={industries} />
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
                    searchText !== '' &&
                    option.name.toLowerCase().includes(searchText.toLowerCase())
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
                    searchText !== '' &&
                    option.name.toLowerCase().includes(searchText.toLowerCase())
                  )
                }}
                options={Languages}
              />
            </Form.Item>
          </Form>
        )}
      </Modal>
    )}
    <Button
      type={props.btnType}
      size={props.btnSize}
      onClick={() => {
        if (!props.imOwner) {
          ErrorNotOwner()
          return
        }
        toggleModal()
      }}
    >
      {props.text}
    </Button>
  </>);
}

export default CreateWorkspaceButton
