import { useState } from 'react'
import { Radio, Drawer, Button, message, Form, Space } from 'antd'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import { ButtonType } from 'antd/lib/button'
import { Workspace } from 'interfaces'
import { LeadsStageInput } from 'components/workspace/input_lead_stages'
import { QueryObserverResult } from '@tanstack/react-query'
import CSS from 'utils/css'

type UpsertLeadStagesButtonProps = {
  workspace: Workspace
  organizationId: string
  btnContent: JSX.Element
  btnType: ButtonType
  btnSize: SizeType
  apiPOST: (endpoint: string, data: any) => Promise<any>
  refreshWorkspace: () => Promise<QueryObserverResult<Workspace, unknown>>
  onComplete?: () => void
}

const UpsertLeadStagesButton = (props: UpsertLeadStagesButtonProps) => {
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
        data.workspace_id = props.workspace.id

        props
          .apiPOST('/workspace.update', data)
          .then((_res) => {
            props
              .refreshWorkspace()
              .then(() => {
                message.success('The lead stages have been updated!')

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
          title={'Update lead stages'}
          open={true}
          onClose={toggleDrawer}
          width={760}
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
            initialValues={props.workspace || {}}
            labelCol={{ span: 10 }}
            wrapperCol={{ span: 14 }}
            layout="horizontal"
            className={CSS.margin_a_m + ' ' + CSS.margin_b_xl}
          >
            <Form.Item
              name="lead_stages"
              label="Stages"
              rules={[{ required: true, type: 'array', min: 2 }]}
              shouldUpdate
            >
              <LeadsStageInput />
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

export default UpsertLeadStagesButton
