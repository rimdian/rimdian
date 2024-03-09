import { useMemo, useState } from 'react'
import { Radio, Drawer, Button, Input, Select, message, Form, Space } from 'antd'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import { ButtonType } from 'antd/lib/button'
import { kebabCase } from 'lodash'
import { AppManifest, AppTable, DataHook, Workspace } from 'interfaces'
import Messages from 'utils/formMessages'
import { QueryObserverResult } from '@tanstack/react-query'
import CSS from 'utils/css'
import TableTag from 'components/common/partial_table_tag'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'

type UpsertDataHookButtonProps = {
  domain?: DataHook
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
                if (props.domain) message.success('The data hook has been updated!')
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
          title={props.domain ? 'Update hook' : 'Add a new data hook'}
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

            <Form.Item
              name="only_for_events"
              label="Only for these events"
              rules={[{ type: 'array' }]}
            >
              <Select mode="multiple" placeholder="Select events" options={tableOptions} />
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

export default UpsertDataHookButton
