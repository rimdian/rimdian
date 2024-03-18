import { useMemo, useState } from 'react'
import { Radio, Drawer, Button, Input, Select, message, Form, Space, Table, Modal } from 'antd'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import { ButtonType } from 'antd/lib/button'
import { kebabCase } from 'lodash'
import { AppManifest, AppTable, DataHook, DataHookFor, Workspace } from 'interfaces'
import Messages from 'utils/formMessages'
import { QueryObserverResult } from '@tanstack/react-query'
import CSS from 'utils/css'
import TableTag from 'components/common/partial_table_tag'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faXmark } from '@fortawesome/free-solid-svg-icons'

type UpsertDataHookButtonProps = {
  dataHook?: DataHook
  workspaceId: string
  organizationId: string
  btnContent: JSX.Element
  btnType: ButtonType
  btnSize: SizeType
  apiPOST: (endpoint: string, data: any) => Promise<any>
  refreshWorkspace: () => Promise<QueryObserverResult<Workspace, unknown>>
  onComplete?: () => void
}

const UpsertDataHookButton = (props: UpsertDataHookButtonProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
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
          .apiPOST('/dataHook.upsert', data)
          .then((_res) => {
            props
              .refreshWorkspace()
              .then(() => {
                if (props.dataHook) message.success('The data hook has been updated!')
                else message.success('The data hook has been created!')

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

  const tableOptions = useMemo(() => {
    const tables: any = [
      { label: <TableTag table="user" />, value: 'user' },
      { label: <TableTag table="order" />, value: 'order' },
      { label: <TableTag table="order_item" />, value: 'order_item' },
      { label: <TableTag table="session" />, value: 'session' },
      { label: <TableTag table="pageview" />, value: 'pageview' },
      { label: <TableTag table="custom_event" />, value: 'custom_event' },
      { label: <TableTag table="cart" />, value: 'cart' },
      { label: <TableTag table="cart_item" />, value: 'cart_item' }
    ]
    workspaceCtx.workspace.installed_apps.forEach((app: AppManifest) => {
      if (app.app_tables && app.app_tables.length > 0) {
        app.app_tables.forEach((table: AppTable) => {
          tables.push({
            label: <TableTag table={table.name} />,
            value: table.name
          })
        })
      }
    })
    return tables
  }, [workspaceCtx.workspace.installed_apps])

  return (
    <>
      {visible && (
        <Drawer
          title={props.dataHook ? 'Update hook' : 'Add a new data hook'}
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
            initialValues={
              props.dataHook || {
                for: []
              }
            }
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
              <Radio.Group disabled={props.dataHook ? true : false}>
                <Radio.Button value="on_validation">Before item validation</Radio.Button>
                <Radio.Button value="on_success">On success</Radio.Button>
              </Radio.Group>
            </Form.Item>

            <Form.Item
              name="name"
              label="Name"
              rules={[{ required: true, message: Messages.RequiredField }]}
            >
              <Input
                onChange={(e: any) => {
                  if (!props.dataHook) {
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
              <Input disabled={props.dataHook ? true : false} placeholder="i.e: web" />
            </Form.Item>

            <Form.Item
              name="endpoint"
              label="Webhook URL"
              rules={[{ type: 'string', required: true }]}
            >
              <Input placeholder="https://..." />
            </Form.Item>

            <Form.Item
              name="for"
              label="For events"
              rules={[{ type: 'array', required: true, min: 1 }]}
            >
              <ForInput />
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

const AddForButton = ({ onComplete }: any) => {
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
        title="Add an event kind"
        okText="Add"
        cancelText="Cancel"
        onCancel={() => {
          setModalVisible(false)
        }}
        onOk={() => {
          form
            .validateFields()
            .then((values: any) => {
              // console.log('onComplete', values)
              form.resetFields()
              setModalVisible(false)
              onComplete(values)
            })
            .catch((info) => {
              console.log('Validate Failed:', info)
            })
        }}
      >
        <Form
          form={form}
          labelCol={{ span: 8 }}
          wrapperCol={{ span: 16 }}
          className={CSS.margin_t_xl}
        >
          <Form.Item name="kind" label="Event kind" rules={[{ required: true, type: 'string' }]}>
            <Input placeholder="user, order, session, cart, custom_event, pageview, segment..." />
          </Form.Item>

          <Form.Item name="action" label="Action" rules={[{ required: true, type: 'string' }]}>
            <Select
              options={[
                {
                  label: 'create',
                  value: 'create'
                },
                {
                  label: 'update',
                  value: 'update'
                },
                {
                  label: 'enter (for segment only)',
                  value: 'enter'
                },
                {
                  label: 'exit (for segment only)',
                  value: 'exit'
                }
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

const ForInput = ({ value, onChange }: any) => {
  const removeFor = (index: number) => {
    let entries = value.slice()
    entries.splice(index, 1)
    onChange(entries)
  }

  return (
    <div>
      {value && value.length > 0 && (
        <Table
          size="middle"
          bordered={false}
          pagination={false}
          rowKey={(record) => record.kind + record.action}
          showHeader={false}
          className={CSS.margin_b_m}
          columns={[
            {
              title: '',
              key: 'code',
              render: (item: DataHookFor) => {
                return (
                  <div>
                    <TableTag table={item.kind} /> - {item.action}
                  </div>
                )
              }
            },
            {
              title: '',
              key: 'remove',
              render: (_text, _record: any, index: number) => {
                return (
                  <div className={CSS.text_right}>
                    <Button type="dashed" size="small" onClick={removeFor.bind(null, index)}>
                      <FontAwesomeIcon icon={faXmark} />
                    </Button>
                  </div>
                )
              }
            }
          ]}
          dataSource={value}
        />
      )}

      <AddForButton
        onComplete={(values: any) => {
          let entries = value.slice()
          if (!entries.find((v: any) => v.code === values.code)) {
            entries.push(values)
            onChange(entries)
          }
        }}
      />
    </div>
  )
}

export default UpsertDataHookButton
