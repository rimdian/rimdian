import { useState } from 'react'
import { SelectedBlockButtonsProp } from '../Editor'
import { Tooltip, Popconfirm, Modal, Form, Button, Spin, Input, message, Select } from 'antd'
import { DragOutlined, DeleteOutlined, CopyOutlined, SaveOutlined } from '@ant-design/icons'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'

const SelectedBlockButtons = (props: SelectedBlockButtonsProp) => {
  const [saveVisible, setSaveVisible] = useState(false)
  const [loading, setLoading] = useState(false)
  const workspaceCtx = useCurrentWorkspaceCtx()

  const [form] = Form.useForm()

  const toggleModal = () => {
    // reset on hide
    if (saveVisible === true) {
      form.resetFields()
    }
    setSaveVisible(!saveVisible)
  }

  const onSubmit = () => {
    form
      .validateFields()
      .then((values: any) => {
        setLoading(true)

        const data: any = {
          workspace_id: workspaceCtx.workspace.id,
          block: JSON.stringify(props.block),
          name: values.name
        }

        if (values.operation === 'update') {
          data.id = values.id
        }

        workspaceCtx
          .apiPOST('/emailBlock.' + values.operation, data)
          .then(() => {
            if (values.operation === 'create') {
              message.success('The block has been saved!')
            } else {
              message.success('The block has been updated!')
            }

            // update current workspace
            workspaceCtx.refreshWorkspace().then(() => {
              setLoading(false)
              setSaveVisible(false)
              form.resetFields()
            })
          })
          .catch((e: any) => {
            console.error(e)
            message.error(e.message)
            setLoading(false)
          })
      })
      .catch((e: any) => {
        console.error(e)
      })
  }

  // console.log('props', props)
  const isDraggable = props.blockDefinitions[props.block.kind].isDraggable
  const isDeletable = props.blockDefinitions[props.block.kind].isDeletable

  return (
    <div className="rmdeditor-selected-block-buttons">
      {isDraggable === true && (
        <Tooltip placement="left" title="Move">
          <div className="rmdeditor-selected-block-button">
            <DragOutlined style={{ verticalAlign: 'middle', cursor: 'grab' }} />
          </div>
        </Tooltip>
      )}

      {isDraggable === true && (
        <Tooltip placement="left" title="Clone">
          <div
            className="rmdeditor-selected-block-button"
            onClick={props.cloneBlock.bind(null, props.block)}
          >
            <CopyOutlined style={{ verticalAlign: 'middle' }} />
          </div>
        </Tooltip>
      )}

      {isDraggable === true && (
        <Tooltip placement="left" title="Save">
          <div className="rmdeditor-selected-block-button" onClick={toggleModal}>
            <SaveOutlined style={{ verticalAlign: 'middle' }} />
          </div>
        </Tooltip>
      )}

      {isDeletable === true && (
        <Tooltip placement="left" title="Delete">
          <Popconfirm
            title="Are you sure to delete this element?"
            onConfirm={props.deleteBlock.bind(null, props.block)}
            okText="Yes"
            cancelText="No"
          >
            <div className="rmdeditor-selected-block-button">
              <DeleteOutlined style={{ verticalAlign: 'middle' }} />
            </div>
          </Popconfirm>
        </Tooltip>
      )}

      {saveVisible && (
        <Modal
          title="Save block"
          wrapClassName="vertical-center-modal"
          open={true}
          onCancel={toggleModal}
          footer={[
            <Button key="back" type="ghost" loading={loading} onClick={toggleModal}>
              Cancel
            </Button>,
            <Button key="submit" type="primary" loading={loading} onClick={onSubmit}>
              Confirm
            </Button>
          ]}
        >
          <Spin tip="Loading..." spinning={loading}>
            <Form form={form} initialValues={{ operation: 'create' }} layout="vertical">
              <Form.Item
                name="operation"
                label="Operation"
                rules={[{ required: true, message: 'This field is required!' }]}
              >
                <Select
                  options={[
                    { label: 'Save as new block', value: 'create' },
                    { label: 'Update existing block', value: 'update' }
                  ]}
                />
              </Form.Item>

              <Form.Item noStyle shouldUpdate={true}>
                {({ getFieldValue }: any) => {
                  return (
                    <>
                      {getFieldValue('operation') === 'update' && (
                        <Form.Item
                          name="id"
                          label="Block"
                          rules={[{ required: true, message: 'This field is required!' }]}
                        >
                          <Select
                            onChange={(val: any) => {
                              form.setFieldsValue({
                                name: workspaceCtx.workspace.emailBlocks.find(
                                  (x: any) => x.id === val
                                ).name
                              })
                            }}
                            options={workspaceCtx.workspace.emailBlocks.map((b: any) => {
                              return {
                                label: b.name,
                                value: b.id
                              }
                            })}
                          />
                        </Form.Item>
                      )}

                      {(getFieldValue('operation') === 'create' || getFieldValue('id')) && (
                        <Form.Item
                          name="name"
                          label="Name"
                          rules={[{ required: true, message: 'This field is required!' }]}
                        >
                          <Input />
                        </Form.Item>
                      )}
                    </>
                  )
                }}
              </Form.Item>
            </Form>
          </Spin>
        </Modal>
      )}

      {/* <span style={{fontSize: '10px'}}>kind: {props.block.kind}, {isDraggable && <>draggable into: {props.blockDefinitions[props.block.kind].draggableIntoGroup}</>}</span> */}
    </div>
  )
}

export default SelectedBlockButtons
