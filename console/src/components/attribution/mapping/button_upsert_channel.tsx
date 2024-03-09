import { useState } from 'react'
import {
  Table,
  Alert,
  Drawer,
  Form,
  Input,
  Button,
  Select,
  Divider,
  Tag,
  Modal,
  message,
  Space
} from 'antd'
import { find, kebabCase } from 'lodash'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faXmark } from '@fortawesome/free-solid-svg-icons'
import { ButtonType } from 'antd/lib/button'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import { Channel, ChannelGroup, Origin, VoucherCode } from 'interfaces'
import { VoucherCodePopover } from './tab_mapping'
import CSS from 'utils/css'

const { Option } = Select

export const FindChannelFromOrigin = (
  allChannels: Channel[],
  utmSource: string,
  utmMedium: string,
  utmCampaign?: string
): Channel | undefined => {
  utmSource = utmSource.trim()
  utmMedium = utmMedium.trim()
  if (utmCampaign) utmCampaign = utmCampaign.trim()

  return allChannels.find((ch) => {
    return ch.origins.find((o) => {
      if (o.utm_source === utmSource && o.utm_medium === utmMedium) {
        // matches without campaign specified
        if ((!o.utm_campaign || o.utm_campaign === '') && (!utmCampaign || utmCampaign === '')) {
          return ch
        }

        // matches without campaign specified
        if (utmCampaign && utmCampaign !== '' && o.utm_campaign && o.utm_campaign === utmCampaign) {
          return ch
        }
      }
      return undefined
    })
  })
}

type AddOriginButtonProps = {
  channels: Channel[]
  onComplete: (origin: Origin) => void
}

const AddOriginButton = (props: AddOriginButtonProps) => {
  const [form] = Form.useForm()
  const [modalVisible, setModalVisible] = useState(false)

  const onClicked = () => {
    setModalVisible(true)
  }

  const mappingValidator = (_rule: any, value: any) => {
    const source = form.getFieldValue('utm_source')
    const medium = form.getFieldValue('utm_medium')
    let campaign = form.getFieldValue('utm_campaign')

    if (!source || !medium) return Promise.resolve(undefined)

    // check if this source/medium is already in another channel
    const foundChannel = FindChannelFromOrigin(props.channels, source, medium, campaign)

    if (foundChannel) {
      return Promise.reject(
        'This origin is already mapped in the channel ' + foundChannel.name + '!'
      )
    }

    return Promise.resolve(undefined)
  }

  return (
    <>
      <Button type="primary" size="small" block onClick={onClicked}>
        Add
      </Button>
      <Modal
        open={modalVisible}
        title="Add an origin"
        okText="Confirm"
        width={600}
        cancelText="Cancel"
        onCancel={() => {
          setModalVisible(false)
        }}
        onOk={() => {
          form
            .validateFields()
            .then((values: Origin) => {
              // console.log('onComplete', values);
              form.resetFields()
              setModalVisible(false)
              values.match_operator = 'equals'
              values.id = values.utm_source.trim() + ' / ' + values.utm_medium.trim()
              if (values.utm_campaign && values.utm_campaign.trim() !== '')
                values.id += ' / ' + values.utm_campaign.trim()
              props.onComplete(values)
            })
            .catch(console.error)
        }}
      >
        <Form form={form} name="form_add_origin" labelCol={{ span: 8 }} wrapperCol={{ span: 14 }}>
          <Form.Item
            label="utm_source"
            name="utm_source"
            rules={[{ required: true, type: 'string' }, { validator: mappingValidator }]}
          >
            <Input />
          </Form.Item>

          <Form.Item
            label="utm_medium"
            name="utm_medium"
            rules={[{ required: true, type: 'string' }, { validator: mappingValidator }]}
          >
            <Input />
          </Form.Item>

          <Form.Item
            label="utm_campaign - optional"
            name="utm_campaign"
            rules={[{ required: false, type: 'string' }, { validator: mappingValidator }]}
          >
            <Input />
          </Form.Item>

          {/* <Form.Item shouldUpdate>
                  {(_funcs) => {
                      // console.log('values', funcs.getFieldsValue());

                      // const values = funcs.getFieldsValue()

                      return <>
                          <Form.Item noStyle label="Source / Medium" name="value" rules={[
                              { required: true, type: 'string' },
                              {
                                  validator: (_rule, value) => {
                                      if (!value) return Promise.resolve(undefined);

                                      const parts = value.split(' / ')
                                      if (parts.length < 2 || parts[1] === '') {
                                          return Promise.reject('The "medium" field is required!')
                                      }
                                      if (parts[0] === '') {
                                          return Promise.reject('The "source" field is required!')
                                      }

                                      // check if this source/medium is already in another channel
                                      const foundChannel = props.channels.find(ch => ch.origins?.find((p: any) => p.value === value))

                                      if (foundChannel) {
                                          return Promise.reject('This "source / medium" is already mapped in the channel ' + foundChannel.name + '!')
                                      }

                                      return Promise.resolve(undefined);
                                  }
                              },
                          ]}>
                              <SourceMediumCampaignInput />
                          </Form.Item>
                      </>
                  }}
              </Form.Item> */}
        </Form>
      </Modal>
    </>
  )
}

const AddVoucherButton = ({ origins, onComplete }: any) => {
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
        title="Add voucher code"
        okText="Confirm"
        cancelText="Cancel"
        onCancel={() => {
          setModalVisible(false)
        }}
        onOk={() => {
          form
            .validateFields()
            .then((values: any) => {
              // console.log('onComplete', values);
              form.resetFields()
              setModalVisible(false)
              onComplete(values)
            })
            .catch((info) => {
              console.log('Validate Failed:', info)
            })
        }}
      >
        <Form form={form} name="form_add_voucher" labelCol={{ span: 8 }} wrapperCol={{ span: 16 }}>
          <Form.Item label="Voucher code" name="code" rules={[{ required: true, type: 'string' }]}>
            <Input type="text" />
          </Form.Item>

          <Form.Item
            label="Attribute to"
            name="origin_id"
            rules={[{ required: true, type: 'string' }]}
          >
            <Select>
              {origins.map((path: Origin) => (
                <Option key={path.id} id={path.id}>
                  {path.id}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            label="utm_campaign"
            name="set_utm_campaign"
            rules={[{ required: false, type: 'string' }]}
          >
            <Input type="text" />
          </Form.Item>

          <Form.Item
            label="utm_content"
            name="set_utm_content"
            rules={[{ required: false, type: 'string' }]}
          >
            <Input type="text" />
          </Form.Item>

          <Form.Item
            label="Description"
            name="description"
            rules={[{ required: false, type: 'string' }]}
          >
            <Input type="text" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

// type SourceMediumCampaignInputProps = {
//     onChange?: (path: string) => void
// }

// const SourceMediumCampaignInput = (props: SourceMediumCampaignInputProps) => {

//     const [source, setSource] = useState<string | undefined>()
//     const [medium, setMedium] = useState<string | undefined>()
//     const [campaign, setCampaign] = useState<string | undefined>()

//     const updateField = (fieldName: string, e: any) => {
//         const value = e.target.value.trim()
//         if (fieldName === 'source') {
//             setSource(value)
//             props.onChange?.(value + ' / ' + medium + ((campaign && campaign.length > 0) ? ' / ' + campaign : ''))
//         }
//         if (fieldName === 'medium') {
//             setMedium(value)
//             props.onChange?.(source + ' / ' + value + ((campaign && campaign.length > 0) ? ' / ' + campaign : ''))
//         }
//         if (fieldName === 'campaign') {
//             setCampaign(value)
//             if (value !== '') {
//                 props.onChange?.(source + ' / ' + medium + ' / ' + value)
//             } else {
//                 props.onChange?.(source + ' / ' + medium)
//             }
//         }
//     }

//     const spacer: CSSProperties = { width: '8%', borderLeft: 0, textAlign: 'center', pointerEvents: 'none', backgroundColor: '#fff' }

//     return <Input.Group compact style={{ display: 'inline-block', width: '100%' }}>
//         <Input required type="text" style={{ width: '28%' }} placeholder="source" value={source} onChange={updateField.bind(null, 'source')} />
//         <Input style={spacer} placeholder="/" disabled />
//         <Input required type="text" style={{ width: '28%', borderLeft: 0 }} placeholder="medium" value={medium} onChange={updateField.bind(null, 'medium')} />
//         <Input style={spacer} placeholder="/" disabled />
//         <Input type="text" style={{ width: '28%', borderLeft: 0 }} placeholder="optional campaign" value={campaign} onChange={updateField.bind(null, 'campaign')} />
//     </Input.Group>
// }

const VoucherCodesInput = ({ origins, value, onChange }: any) => {
  const removeVoucher = (index: number) => {
    let vouchers = value.slice()
    vouchers.splice(index, 1)
    onChange(vouchers)
  }

  return (
    <div>
      {value && value.length > 0 && (
        <Table
          size="middle"
          bordered={false}
          pagination={false}
          rowKey="code"
          showHeader={false}
          className={CSS.margin_b_m}
          columns={[
            {
              title: '',
              key: 'code',
              render: (item: VoucherCode) => <VoucherCodePopover voucherCode={item} />
            },
            {
              title: '',
              key: 'remove',
              render: (_text, _record: any, index: number) => {
                return (
                  <div className={CSS.text_right}>
                    <Button type="dashed" size="small" onClick={removeVoucher.bind(null, index)}>
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

      <AddVoucherButton
        origins={origins}
        onComplete={(values: any) => {
          let vouchers = value.slice()
          if (!vouchers.find((v: any) => v.code === values.code)) {
            vouchers.push(values)
            onChange(vouchers)
          }
        }}
      />
    </div>
  )
}

type OriginsInputProps = {
  channels: Channel[]
  onChange?: (origins: Origin[]) => void
  value?: Origin[]
}

const OriginsInput = (props: OriginsInputProps) => {
  const removeOrigin = (index: number) => {
    const origins = props.value ? props.value.slice() : []
    origins.splice(index, 1)
    props.onChange?.(origins)
  }

  return (
    <div>
      {props.value && props.value.length > 0 && (
        <Table
          size="middle"
          bordered={false}
          pagination={false}
          rowKey="id"
          showHeader={false}
          className={CSS.margin_b_m}
          columns={[
            {
              title: '',
              key: 'path',
              render: (item: any) => item.id
            },
            {
              title: '',
              key: 'remove',
              render: (_text, _record: any, index: number) => {
                return (
                  <div className={CSS.text_right}>
                    <Button type="dashed" size="small" onClick={removeOrigin.bind(null, index)}>
                      <FontAwesomeIcon icon={faXmark} />
                    </Button>
                  </div>
                )
              }
            }
          ]}
          dataSource={props.value}
        />
      )}

      <AddOriginButton
        channels={props.channels}
        onComplete={(values: any) => {
          const origins = props.value ? props.value.slice() : []
          origins.push(values)
          props.onChange?.(origins)
        }}
      />
    </div>
  )
}

type UpsertChannelButtonProps = {
  channel?: Channel
  workspaceId: string
  channels: Channel[]
  channelGroups: ChannelGroup[]
  btnContent: JSX.Element
  apiPOST: (endpoint: string, data: any) => Promise<any>
  onComplete: () => void
  btnType?: ButtonType
  btnSize?: SizeType
}

const UpsertChannelButton = (props: UpsertChannelButtonProps) => {
  const [form] = Form.useForm()
  const [drawerVisible, setDrawerVisible] = useState(false)
  const [loading, setLoading] = useState(false)

  const closeDrawer = () => {
    setDrawerVisible(false)
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

    if (props.channel) {
      values.id = props.channel.id
    } else {
      values.id = generateId(values.name, props.channels, 1)
    }

    values.workspace_id = props.workspaceId

    props
      .apiPOST(props.channel ? '/channel.update' : '/channel.create', values)
      .then(() => {
        if (props.channel) {
          message.success('The channel has successfully been updated.')
        } else {
          message.success('The channel has successfully been created.')
          form.resetFields()
        }

        setLoading(false)
        setDrawerVisible(false)
        props.onComplete()
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  // console.log('commissionGroup', commissionGroup);

  const initialValues = Object.assign({ origins: [], voucher_codes: [] }, props.channel)

  if (!initialValues.origins) initialValues.origins = []
  if (!initialValues.voucher_codes) initialValues.voucher_codes = []

  // console.log('initialValues', initialValues);

  return (
    <>
      <Button type={props.btnType} size={props.btnSize} onClick={() => setDrawerVisible(true)}>
        {props.btnContent}
      </Button>
      {drawerVisible && (
        <Drawer
          title={props.channel ? 'Update channel' : 'Create a new channel'}
          width={900}
          open={true}
          onClose={closeDrawer}
          bodyStyle={{ paddingBottom: 80 }}
          extra={
            <Space>
              <Button loading={loading} onClick={closeDrawer}>
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
            <Form.Item name="name" label="Name" rules={[{ required: true, type: 'string' }]}>
              <Input />
            </Form.Item>

            <Form.Item name="group_id" label="Group" rules={[{ required: true, type: 'string' }]}>
              <Select>
                {props.channelGroups
                  .filter((g: ChannelGroup) => g.id !== '_id' && g.id !== 'not-mapped')
                  .sort((a: ChannelGroup, b: ChannelGroup) => {
                    if (a.created_at) return -1
                    if (b.created_at) return -1
                    return 0
                  })
                  .map((group: any) => (
                    <Select.Option key={group.id} value={group.id}>
                      <Tag color={group.color} style={{ cursor: 'inherit' }}>
                        {group.name}
                      </Tag>
                    </Select.Option>
                  ))}
              </Select>
            </Form.Item>

            <Divider plain>Traffic mapping</Divider>

            <Alert
              className={CSS.margin_b_m}
              message="Traffic coming from these sources will be mapped to this channel."
              type="info"
            />

            <Form.Item
              name="origins"
              label="Origins"
              rules={[{ required: true, type: 'array', min: 1 }]}
              shouldUpdate
            >
              <OriginsInput channels={props.channels} />
            </Form.Item>

            <Divider plain>Voucher code attribution</Divider>

            <Alert
              className={CSS.margin_b_m}
              message="If a conversion contains one of the following voucher code, its last session origin will be attributed to this channel. This mecanism is used to attribute influencers voucher codes to conversions."
              type="info"
            />

            <Form.Item shouldUpdate noStyle>
              {(funcs) => {
                return (
                  <Form.Item
                    name="voucher_codes"
                    label="Voucher codes"
                    rules={[{ required: false, type: 'array', min: 0 }]}
                    shouldUpdate
                  >
                    <VoucherCodesInput origins={funcs.getFieldValue('origins')} />
                  </Form.Item>
                )
              }}
            </Form.Item>
          </Form>
        </Drawer>
      )}
    </>
  )
}

export default UpsertChannelButton
