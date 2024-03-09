import { useState } from 'react'
import { Modal, Form, Input, Button, Select, Tag, message, Space } from 'antd'
import { find, kebabCase } from 'lodash'
import { ButtonType } from 'antd/lib/button'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import { ChannelGroup } from 'interfaces'
import Messages from 'utils/formMessages'

type UpsertChannelGroupButtonProps = {
  channelGroup?: ChannelGroup
  workspaceId: string
  channelGroups: ChannelGroup[]
  btnContent: JSX.Element
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
  btnType?: ButtonType
  btnSize?: SizeType
}

const UpsertChannelGroupButton = (props: UpsertChannelGroupButtonProps) => {
  const [form] = Form.useForm()
  const [modalVisible, setModalVisible] = useState(false)
  const [loading, setLoading] = useState(false)

  const closeModal = () => {
    setModalVisible(false)
  }

  const generateId = (name: string, existingChannels: any, increment: number): string => {
    let id: string = kebabCase(name)

    if (find(existingChannels, (x) => x.id === id)) {
      increment = increment + 1

      id = name + increment

      // check with new increment
      if (find(existingChannels, (x) => x.id === id)) {
        return generateId(name, existingChannels, increment)
      }
    }

    return id
  }

  const onFinish = (values: any) => {
    // console.log('values', values);

    if (loading) return

    setLoading(true)

    if (props.channelGroup) {
      values.id = props.channelGroup.id
    } else {
      values.id = generateId(values.name, props.channelGroups, 1)
    }

    values.workspace_id = props.workspaceId

    props
      .apiPOST('/channelGroup.upsert', values)
      .then((res) => {
        if (props.channelGroup) {
          message.success('The group has successfully been updated.')
        } else {
          message.success('The group has successfully been created.')
          form.resetFields()
        }

        setLoading(false)
        setModalVisible(false)
        props.onComplete()
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  // console.log('commissionGroup', commissionGroup);

  const initialValues = Object.assign({ color: 'magenta' }, props.channelGroup)

  // console.log('initialValues', initialValues);

  return (
    <>
      <Button type={props.btnType} size={props.btnSize} onClick={() => setModalVisible(true)}>
        {props.btnContent}
      </Button>
      {modalVisible && (
        <Modal
          title={props.channelGroup ? 'Update group' : 'Create a new group'}
          width={500}
          open={true}
          footer={
            <Space>
              <Button loading={loading} onClick={closeModal}>
                Cancel
              </Button>
              <Button
                loading={loading}
                onClick={() => {
                  form.submit()
                }}
                type="primary"
              >
                Confirm
              </Button>
            </Space>
          }
        >
          <Form
            form={form}
            initialValues={initialValues}
            labelCol={{ span: 6 }}
            wrapperCol={{ span: 18 }}
            name="groupForm"
            onFinish={onFinish}
          >
            <Form.Item
              name="name"
              label="Name"
              rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
            >
              <Input
                addonAfter={
                  <Form.Item noStyle name="color">
                    <Select style={{ width: 120 }}>
                      <Select.Option value="magenta">
                        <Tag color="magenta">magenta</Tag>
                      </Select.Option>
                      <Select.Option value="red">
                        <Tag color="red">red</Tag>
                      </Select.Option>
                      <Select.Option value="volcano">
                        <Tag color="volcano">volcano</Tag>
                      </Select.Option>
                      <Select.Option value="orange">
                        <Tag color="orange">orange</Tag>
                      </Select.Option>
                      <Select.Option value="gold">
                        <Tag color="gold">gold</Tag>
                      </Select.Option>
                      <Select.Option value="lime">
                        <Tag color="lime">lime</Tag>
                      </Select.Option>
                      <Select.Option value="green">
                        <Tag color="green">green</Tag>
                      </Select.Option>
                      <Select.Option value="cyan">
                        <Tag color="cyan">cyan</Tag>
                      </Select.Option>
                      <Select.Option value="blue">
                        <Tag color="blue">blue</Tag>
                      </Select.Option>
                      <Select.Option value="geekblue">
                        <Tag color="geekblue">geekblue</Tag>
                      </Select.Option>
                      <Select.Option value="purple">
                        <Tag color="purple">purple</Tag>
                      </Select.Option>
                      <Select.Option value="grey">
                        <Tag>grey</Tag>
                      </Select.Option>
                    </Select>
                  </Form.Item>
                }
              />
            </Form.Item>
          </Form>
        </Modal>
      )}
    </>
  )
}

export default UpsertChannelGroupButton
