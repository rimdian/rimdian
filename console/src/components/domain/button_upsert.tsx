import { useState } from 'react'
import { Radio, Drawer, Button, Input, Select, message, Form, Modal, Table, Tag, Space } from 'antd'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import { ButtonType } from 'antd/lib/button'
import { get, kebabCase } from 'lodash'
import { Domain, DomainHost, DomainKind, Workspace } from 'interfaces'
import Messages from 'utils/formMessages'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faClose } from '@fortawesome/free-solid-svg-icons'
import { QueryObserverResult } from '@tanstack/react-query'
import CSS from 'utils/css'

type UpsertDomainButtonProps = {
  domain?: Domain
  workspaceId: string
  organizationId: string
  btnContent: JSX.Element
  btnType: ButtonType
  btnSize: SizeType
  apiPOST: (endpoint: string, data: any) => Promise<any>
  refreshWorkspace: () => Promise<QueryObserverResult<Workspace, unknown>>
  onComplete?: () => void
}

const UpsertDomainButton = (props: UpsertDomainButtonProps) => {
  const [visible, setVisible] = useState(false)
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  const toggleDrawer = () => {
    if (visible) form.resetFields()
    setVisible(!visible)
  }

  const onSubmit = () => {
    form
      .validateFields()
      .then((values: any) => {
        if (loading) return
        setLoading(true)

        const data = { ...values }
        data.organization_id = props.organizationId
        data.workspace_id = props.workspaceId

        props
          .apiPOST('/domain.upsert', data)
          .then((_res) => {
            props
              .refreshWorkspace()
              .then(() => {
                if (props.domain) message.success('The domain has been updated!')
                else message.success('The domain has been created!')

                setLoading(false)
                toggleDrawer()

                if (props.onComplete) props.onComplete()
              })
              .catch((_) => {
                setLoading(false)
              })
          })
          .catch((_) => {
            setLoading(false)
          })
      })
      .catch((_) => {})
  }

  return (
    <>
      {visible && (
        <Drawer
          title={props.domain ? 'Update domain' : 'Add a new domain'}
          open={true}
          onClose={toggleDrawer}
          width={960}
          extra={
            <Space>
              <Button loading={loading} onClick={toggleDrawer}>
                Cancel
              </Button>
              <Button loading={loading} onClick={onSubmit} type="primary">
                Save
              </Button>
            </Space>
          }
        >
          <Form
            form={form}
            initialValues={props.domain}
            labelCol={{ span: 6 }}
            wrapperCol={{ span: 14 }}
            layout="horizontal"
            className={CSS.margin_a_m + ' ' + CSS.margin_b_xl}
          >
            <Form.Item
              name="type"
              label="Kind"
              rules={[{ required: true, message: Messages.RequiredField }]}
            >
              <Radio.Group disabled={props.domain ? true : false}>
                <Radio.Button value={DomainKind.Web}>Web</Radio.Button>
                <Radio.Button value={DomainKind.Marketplace}>Marketplace</Radio.Button>
                <Radio.Button value={DomainKind.App}>App</Radio.Button>
                <Radio.Button value={DomainKind.Retail}>Retail</Radio.Button>
                <Radio.Button value={DomainKind.Telephone}>Telephone</Radio.Button>
              </Radio.Group>
            </Form.Item>

            <Form.Item
              name="name"
              label="Name"
              rules={[{ required: true, message: Messages.RequiredField }]}
            >
              <Input
                placeholder="i.e: ios, android, paris, amazon, lazada..."
                onChange={(e: any) => {
                  if (!props.domain) {
                    form.setFieldsValue({ id: kebabCase(e.target.value) })
                  }
                }}
              />
            </Form.Item>

            <Form.Item
              name="id"
              label="ID"
              rules={[
                {
                  required: true,
                  type: 'string',
                  pattern: /^[a-z0-9]+(-[a-z0-9]+)*$/,
                  message: Messages.InvalidIdFormat
                }
              ]}
            >
              <Input disabled={props.domain ? true : false} placeholder="i.e: web" />
            </Form.Item>

            <Form.Item noStyle shouldUpdate>
              {(funcs) => {
                const type = funcs.getFieldValue('type')

                return (
                  <>
                    {type === 'web' && (
                      <>
                        <Form.Item
                          name="hosts"
                          label="Hosts"
                          rules={[{ required: true, type: 'array', min: 1 }]}
                          shouldUpdate
                        >
                          <HostsInput />
                        </Form.Item>

                        {/* <Form.Item name="host" label={t('domain_name', "Domain name")} rules={[{ required: true, pattern: /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9-]*[A-Za-z0-9])+(:[0-9]{1,5})?$/, message: t('hostname_invalid', "The value should be a valid host name!") }]}>
      <Input placeholder="i.e: my-website.com / blog.my-website.com / localhost:8080" />
    </Form.Item> */}

                        <Form.Item
                          name="params_whitelist"
                          label="URL parameters to preserve"
                          extra="Press enter to insert a new value."
                          rules={[
                            {
                              required: false,
                              validator: (_rule, value) => {
                                return new Promise((resolve, reject) => {
                                  if (value && value.length) {
                                    value.forEach((param: any) => {
                                      if (!/^([A-Za-z0-9-_~]*)$/.test(param)) {
                                        return reject(new Error(Messages.InvalidURLParamsFormat))
                                      }
                                    })
                                  }
                                  return resolve(undefined)
                                })
                              },
                              message: Messages.InvalidURLParamsFormat
                            }
                          ]}
                        >
                          <Select mode="tags" placeholder="" dropdownStyle={{ display: 'none' }}>
                            {get(props, 'domain.params_whitelist', []).map((param: any) => (
                              <Select.Option key={param} value={param}>
                                {param}
                              </Select.Option>
                            ))}
                          </Select>
                        </Form.Item>

                        {/* <Form.Item valuePropName="checked" name="brandKeywordsAsDirect" label="Brand keywords as Direct Traffic" extra="When your customers already know you and type your brand in a search engine to visit your website (i.e: Adwords...), we recommend you to consider those sessions as Direct Traffic." rules={[{ type: 'boolean', required: false, message: Messages.RequiredField }]}>
                            <Switch />
                        </Form.Item>

                        {brandKeywordsAsDirect && <>
                            <Form.Item name="brandKeywords" label="Brand keywords" extra="Your brand keywords are used to detect Direct Traffic coming from SEA. Press enter to insert a new value." rules={[{ required: true, type: "array", min: 1, message: Messages.InvalidArrayOfStrings }]}>
                                <Select mode="tags" placeholder="Enter your brand keywords" dropdownStyle={{ display: 'none' }}>
                                    {props.domain?.brandKeywords?.map((kw: any) => <Select.Option key={kw} value={kw}>{kw}</Select.Option>)}
                                </Select>
                            </Form.Item>

                            <Form.Item
                                name="homepagePaths"
                                label="Homepage URL paths"
                                extra="These paths are used to detect visits coming from SEO brand keywords to your homepage. Press enter to insert a new value."
                                rules={[{
                                    required: true,
                                    validator: (_rule, value) => {
                                        return new Promise((resolve, reject) => {
                                            if (value && value.length) {
                                                value.forEach((param: any) => {
                                                    if (!/^\/([A-Za-z0-9-_~/]*)$/.test(param)) {
                                                        return reject(new Error(Messages.InvalidPath))
                                                    }
                                                })
                                            } else {
                                                return reject(new Error(Messages.RequiredField))
                                            }
                                            return resolve(undefined)
                                        })
                                    },
                                    message: Messages.InvalidPath
                                }]}
                            >
                                <Select mode="tags" placeholder="" dropdownStyle={{ display: 'none' }} />
                            </Form.Item>
                        </>} */}
                      </>
                    )}
                  </>
                )
              }}
            </Form.Item>
          </Form>
        </Drawer>
      )}
      <Button type={props.btnType} size={props.btnSize} onClick={toggleDrawer}>
        {props.btnContent}
      </Button>
    </>
  )
}

export default UpsertDomainButton

type HostsInputProps = {
  value?: object[]
  onChange?: (data: any) => void
}

const HostsInput: React.FC<HostsInputProps> = ({ value = [], onChange }) => {
  const removeHost = (index: number) => {
    let hosts = value.slice()
    hosts.splice(index, 1)
    onChange?.(hosts)
  }

  return (
    <div>
      {value && value.length > 0 && (
        <Table
          size="middle"
          bordered={false}
          pagination={false}
          rowKey="host"
          showHeader={false}
          className={CSS.margin_b_m}
          columns={[
            {
              title: '',
              key: 'host',
              render: (x: DomainHost) => (
                <>
                  <Tag>{x.host}</Tag>{' '}
                  {x.path_prefix && x.path_prefix !== '' && (
                    <small>
                      rewrite URLs prefix with <Tag>{x.path_prefix}</Tag>
                    </small>
                  )}
                </>
              )
            },
            {
              title: '',
              key: 'remove',
              render: (_text, _record: any, index: number) => {
                return (
                  <div className={CSS.text_right}>
                    <Button type="dashed" size="small" onClick={removeHost.bind(null, index)}>
                      <FontAwesomeIcon icon={faClose} />
                    </Button>
                  </div>
                )
              }
            }
          ]}
          dataSource={value}
        />
      )}

      <AddHostButton
        onComplete={(newHost: any) => {
          let hosts = value ? value.slice() : []
          hosts.push(newHost)
          onChange?.(hosts)
        }}
      />
    </div>
  )
}

const AddHostButton = ({ onComplete }: any) => {
  const [form] = Form.useForm()
  const [modalVisible, setModalVisible] = useState(false)

  const onClicked = () => {
    setModalVisible(true)
  }

  return (
    <>
      <Button type="primary" size="small" block onClick={onClicked}>
        Add
      </Button>
      <Modal
        open={modalVisible}
        title="Add host"
        okText="Confirm"
        cancelText="Cancel"
        onCancel={() => {
          setModalVisible(false)
        }}
        onOk={() => {
          form
            .validateFields()
            .then((values: any) => {
              form.resetFields()
              setModalVisible(false)
              onComplete(values)
            })
            .catch(console.error)
        }}
      >
        <Form
          form={form}
          name="form_add_host"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
          layout="horizontal"
        >
          <Form.Item
            name="host"
            label="Hostname"
            rules={[
              {
                required: true,
                type: 'string',
                pattern:
                  /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9-]*[A-Za-z0-9])+(:[0-9]{1,5})?$/,
                message: Messages.InvalidHostname
              }
            ]}
          >
            <Input placeholder="blog.website.com" />
          </Form.Item>

          <Form.Item
            name="path_prefix"
            label="Rewrite URLs with prefix"
            rules={[
              {
                required: false,
                type: 'string',
                pattern: /^[a-z0-9]+(-[a-z0-9]+)*$/,
                message: Messages.InvalidIdFormat
              }
            ]}
          >
            <Input placeholder="i.e: blog" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
