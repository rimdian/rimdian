import {
  Button,
  Col,
  Drawer,
  Form,
  Input,
  Row,
  Select,
  Space,
  Tag,
  Progress,
  Popover,
  message,
  Modal,
  Alert
} from 'antd'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { useMemo, useState } from 'react'
import { TreeNodeInput, HasLeaf } from './input'
import { Segment } from './interfaces'
import { Timezones } from 'utils/countries_timezones'
import CSS from 'utils/css'
import { forEach, size } from 'lodash'

const ButtonUpsertSegment = (props: { segment?: Segment }) => {
  const [drawserVisible, setDrawserVisible] = useState(false)
  const workspaceCtx = useCurrentWorkspaceCtx()

  const segmentsCount = useMemo(() => {
    return size(workspaceCtx.segmentsMap)
  }, [workspaceCtx.segmentsMap])

  const max = workspaceCtx.workspace.license_info.usq
  const onNewSegment = () => {
    if (segmentsCount >= max) {
      Modal.warning({
        title: 'Quota reached',
        content: (
          <Alert
            description={`You have reached your user segments quota of ${segmentsCount}/${max}. Please upgrade your license to create more segments.`}
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
      Edit segment
    </Button>
  ) : (
    <Button type="primary" block ghost onClick={onNewSegment}>
      New segment
    </Button>
  )

  // but the drawer in a separate component to make sure the
  // form is reset when the drawer is closed
  return (
    <>
      {button}
      {drawserVisible && (
        <DrawerSegment segment={props.segment} setDrawserVisible={setDrawserVisible} />
      )}
    </>
  )
}

const DrawerSegment = (props: { segment?: Segment; setDrawserVisible: any }) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [loadingPreview, setLoadingPreview] = useState(false)
  const [previewedData, setPreviewedData] = useState<string | undefined>() // track the tree hash to avoid re-render
  const [previewResponse, setPreviewResponse] = useState<any>()

  const preview = () => {
    if (loadingPreview) return
    setLoadingPreview(true)

    const values = form.getFieldsValue()
    const data = {
      workspace_id: workspaceCtx.workspace.id,
      parent_segment_id: values.parent_segment_id,
      tree: values.tree,
      timezone: values.timezone
    }

    // compute data hash
    setPreviewedData(JSON.stringify(data))

    workspaceCtx
      .apiPOST('/segment.preview', data)
      .then((res) => {
        // console.log('res', res)
        setPreviewResponse(res)
        setLoadingPreview(false)
      })
      .catch(() => {
        setLoadingPreview(false)
      })
  }

  const initialValues = Object.assign(
    {
      color: 'blue',
      parent_segment_id: 'authenticated',
      timezone: workspaceCtx.workspace.default_user_timezone,
      tree: {
        kind: 'branch',
        branch: {
          operator: 'and',
          leaves: []
        }
      }
    },
    props.segment
  )

  // console.log('workspaceCtx', workspaceCtx)

  const onFinish = (values: any) => {
    // console.log('values', values);
    if (loading) return

    setLoading(true)

    const data = { ...values }
    data.workspace_id = workspaceCtx.workspace.id

    if (props.segment) {
      data.id = props.segment.id
    }

    workspaceCtx
      .apiPOST('/segment.' + (props.segment ? 'update' : 'create'), data)
      .then((_res) => {
        workspaceCtx
          .refetchSegments()
          .then(() => {
            if (props.segment) message.success('The segment has been updated!')
            else message.success('The segment has been created!')

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
  // compute graph network of tables
  const schemas = useMemo(() => {
    // console.log('workspaceCtx.cubeSchemasMap', workspaceCtx.cubeSchemasMap)

    const results = {
      user: workspaceCtx.cubeSchemasMap['User'],
      order: workspaceCtx.cubeSchemasMap['Order'],
      order_item: workspaceCtx.cubeSchemasMap['Order_item'],
      custom_event: workspaceCtx.cubeSchemasMap['Custom_event'],
      session: workspaceCtx.cubeSchemasMap['Session'],
      pageview: workspaceCtx.cubeSchemasMap['Pageview']
      // cart: workspaceCtx.cubeSchemasMap['Cart']
    } as any

    // console.log('results', results)
    // add cube schemas that join with user
    workspaceCtx.workspace.installed_apps.forEach((app) => {
      forEach(app.cube_schemas, (schema, cubeName) => {
        forEach(schema.joins, (_join, joinTable) => {
          if (joinTable === 'User') {
            results[cubeName] = schema
          }
        })
      })
    })

    return results
  }, [workspaceCtx])

  return (
    <Drawer
      title={props.segment ? 'Update segment' : 'New segment'}
      open={true}
      width={'90%'}
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
          <Row gutter={24}>
            <Col span={18}>
              <Form.Item name="name" label="Name" rules={[{ required: true, type: 'string' }]}>
                <Input
                  placeholder="i.e: Big spenders..."
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

              <Form.Item
                name="parent_segment_id"
                label="Only in"
                rules={[{ required: true, type: 'string' }]}
              >
                <Select
                  options={[
                    { value: 'authenticated', label: <Tag color="blue">Authenticated</Tag> },
                    { value: '_all', label: <Tag>All</Tag> }
                  ]}
                  suffixIcon={<>users</>}
                />
              </Form.Item>

              <Form.Item
                name="timezone"
                label="Timezone used for dates"
                rules={[{ required: true, type: 'string' }]}
                className={CSS.margin_b_xxl}
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
                  fieldNames={{ label: 'name', value: 'name' }}
                />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item noStyle dependencies={['tree', 'parent_segment_id', 'timezone']}>
                {() => {
                  if (loadingPreview) {
                    return (
                      <Progress
                        format={() => (
                          <Button type="primary" ghost loading={true}>
                            Preview
                          </Button>
                        )}
                        type="circle"
                        percent={0}
                        width={150}
                      />
                    )
                  }

                  // check if tree has changed
                  const values = form.getFieldsValue()
                  let shouldPreview = false

                  if (values.tree) {
                    const data = {
                      workspace_id: workspaceCtx.workspace.id,
                      parent_segment_id: values.parent_segment_id,
                      tree: values.tree,
                      timezone: values.timezone
                    }

                    // compute data hash
                    const dataHash = JSON.stringify(data)

                    if (!previewedData || previewedData !== dataHash) {
                      shouldPreview = true
                    }
                  }

                  if (shouldPreview) {
                    return (
                      <Progress
                        format={() => (
                          <Button
                            type="primary"
                            ghost
                            onClick={preview}
                            disabled={HasLeaf(values.tree) ? false : true}
                          >
                            Preview
                          </Button>
                        )}
                        type="circle"
                        percent={0}
                        width={150}
                      />
                    )
                  } else if (previewResponse && previewResponse.count >= 0) {
                    let content = (
                      <span className={CSS.font_size_m}>{previewResponse.count} users</span>
                    )
                    let percent = 0

                    if (previewResponse.count === 0) {
                      content = <>0 users</>
                    } else {
                      if (values.parent_segment_id === '_all') {
                        const total =
                          workspaceCtx.segmentsMap['authenticated'].users_count +
                          workspaceCtx.segmentsMap['anonymous'].users_count
                        percent = (previewResponse.count * 100) / total
                      } else {
                        percent =
                          (previewResponse.count * 100) /
                          workspaceCtx.segmentsMap[values.parent_segment_id].users_count
                      }
                    }

                    return (
                      <Popover
                        title="SQL conditions generated"
                        placement="left"
                        content={
                          <div style={{ width: 500 }}>
                            <p>{previewResponse.sql}</p>
                            Args: {JSON.stringify(previewResponse.args, null, 2)}
                          </div>
                        }
                      >
                        <Progress
                          format={() => content}
                          type="circle"
                          percent={percent}
                          width={150}
                          status="normal"
                          strokeColor={{
                            '0%': '#4e6cff',
                            '100%': '#8E2DE2'
                          }}
                        />
                      </Popover>
                    )
                  }

                  return 'No preview available...'
                }}
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="tree"
            noStyle
            rules={[
              {
                required: true,
                validator: (_rule, value) => {
                  // console.log('value', value)
                  return new Promise((resolve, reject) => {
                    if (HasLeaf(value)) {
                      return resolve(undefined)
                    }
                    return reject(new Error('A tree is required'))
                  })
                }
                // message: Messages.RequiredField
              }
            ]}
          >
            <TreeNodeInput schemas={schemas} />
          </Form.Item>
        </Form>
      </>
    </Drawer>
  )
}

export default ButtonUpsertSegment
