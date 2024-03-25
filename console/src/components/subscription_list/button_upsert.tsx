import {
  Button,
  Drawer,
  Form,
  Input,
  Select,
  Space,
  Tag,
  message,
  Modal,
  Alert,
  Switch
} from 'antd'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { useMemo, useState } from 'react'
import { size } from 'lodash'
import { SubscriptionList } from 'interfaces'
import EmailTemplateInput from 'components/message_template/input_email'

const ButtonUpsertSubscriptionList = (props: { segment?: SubscriptionList }) => {
  const [drawserVisible, setDrawserVisible] = useState(false)
  const workspaceCtx = useCurrentWorkspaceCtx()

  const listsCount = useMemo(() => {
    return size(workspaceCtx.segmentsMap)
  }, [workspaceCtx.segmentsMap])

  const max = workspaceCtx.workspace.license_info.uslq
  const onNewSubscriptionList = () => {
    if (listsCount >= max) {
      Modal.warning({
        title: 'Quota reached',
        content: (
          <Alert
            description={`You have reached your user subscription list quota of ${listsCount}/${max}. Please upgrade your license to create more segments.`}
            type="warning"
          />
        )
      })
    } else {
      setDrawserVisible(true)
    }
  }

  const button = props.segment ? (
    <Button type="primary" size="small" ghost onClick={() => setDrawserVisible(!drawserVisible)}>
      Edit list
    </Button>
  ) : (
    <Button type="primary" block ghost onClick={onNewSubscriptionList}>
      New list
    </Button>
  )

  // but the drawer in a separate component to make sure the
  // form is reset when the drawer is closed
  return (
    <>
      {button}
      {drawserVisible && (
        <DrawerSubscriptionList list={props.segment} setDrawserVisible={setDrawserVisible} />
      )}
    </>
  )
}

const DrawerSubscriptionList = (props: { list?: SubscriptionList; setDrawserVisible: any }) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)

  const initialValues = Object.assign(
    {
      color: 'blue'
    },
    props.list
  )

  // console.log('workspaceCtx', workspaceCtx)

  const onFinish = (values: any) => {
    // console.log('values', values);
    if (loading) return

    setLoading(true)

    const data = { ...values }
    data.workspace_id = workspaceCtx.workspace.id

    if (props.list) {
      data.id = props.list.id
    }

    workspaceCtx
      .apiPOST('/subscriptionList.' + (props.list ? 'update' : 'create'), data)
      .then((_res) => {
        workspaceCtx
          .refetchSubscriptionLists()
          .then(() => {
            if (props.list) message.success('The list has been updated!')
            else message.success('The list has been created!')

            form.resetFields()
            setLoading(false)
            props.setDrawserVisible(false)

            // if (props.onComplete) props.onComplete()
          })
          .catch((_) => {
            setLoading(false)
          })
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  return (
    <Drawer
      title={props.list ? 'Update subscription list' : 'New subscription list'}
      open={true}
      width={800}
      onClose={() => props.setDrawserVisible(false)}
      bodyStyle={{ paddingBottom: 80 }}
      extra={
        <Space>
          <Button loading={loading} onClick={() => props.setDrawserVisible(false)}>
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
      <>
        <Form
          form={form}
          initialValues={initialValues}
          labelCol={{ span: 8 }}
          wrapperCol={{ span: 12 }}
          name="groupForm"
          onFinish={onFinish}
        >
          <Form.Item name="id" label="ID" rules={[{ required: true, type: 'string' }]}>
            <Input placeholder="i.e: newsletter" />
          </Form.Item>

          <Form.Item name="name" label="Name" rules={[{ required: true, type: 'string' }]}>
            <Input
              placeholder="i.e: Newsletter"
              addonAfter={
                <Form.Item noStyle name="color">
                  <Select
                    style={{ width: 150 }}
                    options={[
                      { label: <Tag color="magenta">magenta</Tag>, value: 'magenta' },
                      { label: <Tag color="red">red</Tag>, value: 'red' },
                      { label: <Tag color="volcano">volcano</Tag>, value: 'volcano' },
                      { label: <Tag color="orange">orange</Tag>, value: 'orange' },
                      { label: <Tag color="gold">gold</Tag>, value: 'gold' },
                      { label: <Tag color="lime">lime</Tag>, value: 'lime' },
                      { label: <Tag color="green">green</Tag>, value: 'green' },
                      { label: <Tag color="cyan">cyan</Tag>, value: 'cyan' },
                      { label: <Tag color="blue">blue</Tag>, value: 'blue' },
                      { label: <Tag color="geekblue">geekblue</Tag>, value: 'geekblue' },
                      { label: <Tag color="purple">purple</Tag>, value: 'purple' },
                      { label: <Tag color="grey">grey</Tag>, value: 'grey' }
                    ]}
                  ></Select>
                </Form.Item>
              }
            />
          </Form.Item>
          {/*  channel */}
          <Form.Item name="channel" label="Channel" rules={[{ required: true, type: 'string' }]}>
            <Select
              options={[
                { label: 'Email', value: 'email' }
                // { label: 'Push', value: 'push' },
                // { label: 'SMS', value: 'sms' },
              ]}
            />
          </Form.Item>
          {/* double opt in if channel is email */}
          <Form.Item noStyle dependencies={['channel']}>
            {() => {
              return form.getFieldValue('channel') === 'email' ? (
                <Form.Item name="double_opt_in" label="Double opt-in" valuePropName="checked">
                  <Switch />
                </Form.Item>
              ) : null
            }}
          </Form.Item>
          <Form.Item noStyle dependencies={['double_opt_in']}>
            {() => {
              return form.getFieldValue('double_opt_in') ? (
                <Form.Item name="email_template_id" label="Email template">
                  <EmailTemplateInput />
                </Form.Item>
              ) : null
            }}
          </Form.Item>
        </Form>
      </>
    </Drawer>
  )
}

export default ButtonUpsertSubscriptionList
