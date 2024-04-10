import {
  Alert,
  Button,
  Col,
  Divider,
  Drawer,
  Form,
  Input,
  Row,
  Select,
  Space,
  Tabs,
  Tag,
  message
} from 'antd'
import { MessageTemplate } from './interfaces'
import { useState } from 'react'
import { cloneDeep, kebabCase } from 'lodash'
import Messages from 'utils/formMessages'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCode, faPalette } from '@fortawesome/free-solid-svg-icons'
import RootBlockDefinition from 'components/email_editor/UI/definitions/Root'
import OneColumnBlockDefinition from 'components/email_editor/UI/definitions/OneColumn'
import Columns168BlockDefinition from 'components/email_editor/UI/definitions/Columns168'
import Columns204BlockDefinition from 'components/email_editor/UI/definitions/Columns204'
import Columns420BlockDefinition from 'components/email_editor/UI/definitions/Columns420'
import Columns816BlockDefinition from 'components/email_editor/UI/definitions/Columns816'
import Columns888BlockDefinition from 'components/email_editor/UI/definitions/Columns888'
import Columns1212BlockDefinition from 'components/email_editor/UI/definitions/Columns1212'
import Columns6666BlockDefinition from 'components/email_editor/UI/definitions/Columns6666'
import ImageBlockDefinition from 'components/email_editor/UI/definitions/Image'
import DividerBlockDefinition from 'components/email_editor/UI/definitions/Divider'
import OpenTrackingBlockDefinition from 'components/email_editor/UI/definitions/OpenTracking'
import ButtonBlockDefinition from 'components/email_editor/UI/definitions/Button'
import TextBlockDefinition from 'components/email_editor/UI/definitions/Text'
import HeadingBlockDefinition from 'components/email_editor/UI/definitions/Heading'
import { Editor, SelectedBlockButtonsProp } from 'components/email_editor/Editor'
import { BlockInterface, BlockDefinitionInterface } from 'components/email_editor/Block'
import ColumnBlockDefinition from 'components/email_editor/UI/definitions/Column'
import { ExportHTML } from 'components/email_editor/UI/Preview'
import { Layout, DesktopWidth } from 'components/email_editor/UI/Layout'
import SelectedBlockButtons from 'components/email_editor/UI/SelectedBlockButtons'
import uuid from 'short-uuid'
import { css } from '@emotion/css'
import InfoRadioGroup from 'components/common/input_info_radio_group'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Nunjucks from 'nunjucks'
// import AceInput from 'components/common/input_ace'
// import IframeSandbox from 'components/email_editor/UI/Widgets/Iframe'
import CSS from 'utils/css'
import extractTLD from 'utils/tld'

const generateBlockFromDefinition = (blockDefinition: BlockDefinitionInterface) => {
  const id = uuid.generate()

  const block: BlockInterface = {
    id: id,
    kind: blockDefinition.kind,
    path: '', // path is set when rendering
    children: blockDefinition.children
      ? blockDefinition.children.map((child: BlockDefinitionInterface) => {
          return generateBlockFromDefinition(child)
        })
      : [],
    data: cloneDeep(blockDefinition.defaultData)
  }

  return block
}

// const image = generateBlockFromDefinition(ImageBlockDefinition)
// const button = generateBlockFromDefinition(ButtonBlockDefinition)
const text = generateBlockFromDefinition(TextBlockDefinition)
const heading = generateBlockFromDefinition(HeadingBlockDefinition)
const logo = generateBlockFromDefinition(ImageBlockDefinition)
const image = generateBlockFromDefinition(ImageBlockDefinition)
const divider = generateBlockFromDefinition(DividerBlockDefinition)
const openTracking = generateBlockFromDefinition(OpenTrackingBlockDefinition)
const btn = generateBlockFromDefinition(ButtonBlockDefinition)
const column = generateBlockFromDefinition(OneColumnBlockDefinition)

// column.data.paddingControl = 'separate'
// column.data.styles.paddingTop = '50px'
// column.data.styles.backgroundColor = '#FFFFFF'
// column.data.styles.borderRadius = '10px'

logo.data.image.src = 'https://cdn-eu.rimdian.com/images/rimdian-email.png'
logo.data.image.alt = 'Rimdian'
logo.data.image.href = 'https://www.rimdian.com'
logo.data.image.width = '50px'

// heading.data.align = 'center'
heading.data.paddingControl = 'separate'
heading.data.paddingTop = '40px'
heading.data.paddingBottom = '40px'
heading.data.editorData[0].children[0].text = 'Hi {{ user.first_name }} 👋'

divider.data.paddingControl = 'separate'
divider.data.paddingTop = '40px'
divider.data.paddingBottom = '20px'
divider.data.paddingLeft = '200px'
divider.data.paddingRight = '200px'

text.data.editorData[0].children[0].text = 'Welcome to the email editor!'

btn.data.button.backgroundColor = '#4e6cff'
btn.data.button.text = '👉 Click me'

column.children[0].children.push(logo)
column.children[0].children.push(heading)
column.children[0].children.push(text)
column.children[0].children.push(divider)
column.children[0].children.push(image)
column.children[0].children.push(btn)
column.children[0].children.push(openTracking)

const rootData = cloneDeep(RootBlockDefinition.defaultData)
// rootData.styles.body.backgroundColor = '#1A237E'

const rootBlock: BlockInterface = {
  id: 'root',
  kind: 'root',
  path: '',
  children: [column],
  data: rootData
}

const tabsInHeader = css({
  position: 'absolute',
  top: 0,
  height: '65px',
  width: '40%',
  right: '30%',
  left: '30%',
  margin: '0 auto',
  textAlign: 'center',
  '& .ant-tabs-tab': {
    lineHeight: '41px'
  }
})

interface ButtonUpsertEmailTemplateProps {
  template?: MessageTemplate
  onSuccess?: (id: string) => void
  btnProps: any
  category?: string
  children: React.ReactNode
  name?: string
  utmSource?: string
  utmMedium?: string
  utmCampaign?: string
}

const ButtonUpsertEmailTemplate = (props: ButtonUpsertEmailTemplateProps) => {
  const [drawserVisible, setDrawserVisible] = useState(false)
  return (
    <>
      {drawserVisible && (
        <DrawerEmailTemplate
          template={props.template}
          setDrawserVisible={setDrawserVisible}
          onSuccess={props.onSuccess}
          name={props.name}
          category={props.category}
          utmSource={props.utmSource}
          utmMedium={props.utmMedium}
          utmCampaign={props.utmCampaign}
        />
      )}
      <Button {...props.btnProps} onClick={() => setDrawserVisible(true)}>
        {props.children}
      </Button>
    </>
  )
}

const DrawerEmailTemplate = (props: {
  template?: MessageTemplate
  setDrawserVisible: any
  onSuccess?: (id: string) => void
  name?: string
  category?: string
  utmSource?: string
  utmMedium?: string
  utmCampaign?: string
}) => {
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()
  const [tab, setTab] = useState('settings')
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [contentDecorated, setContentDecorated] = useState('')
  const [contentField, setContentField] = useState('content')

  const macros: any[] = [] // TODO: fetch template macros

  const submitForm = (values: any) => {
    if (loading) return
    setLoading(true)

    const data = { ...values }
    data.workspace_id = workspaceCtx.workspace.id
    data.channel = 'email'

    if (props.template) {
      data.id = props.template.id
    }

    workspaceCtx
      .apiPOST('/messageTemplate.upsert', data)
      .then(() => {
        message.success('The template has been saved!')
        setLoading(false)
        props.onSuccess && props.onSuccess(data.id)
        props.setDrawserVisible(false)
      })
      .catch(() => {
        setLoading(false)
      })
  }

  const initialValues = Object.assign(
    {
      name: props.name,
      id: props.name ? kebabCase(props.name) : undefined,
      utm_source: props.utmSource || extractTLD(workspaceCtx.workspace.website_url),
      utm_medium: props.utmMedium || 'email',
      utm_campaign: props.utmCampaign || undefined,
      category: props.category,
      engine: 'visual',
      email: {
        visual_editor_tree: rootBlock
      },
      test_data: `{
  "user": {
    "external_id": "user_id",
    "created_at": "2022-10-30T22:21:48.020Z",
    "is_authenticated": true,
    "signed_up_at": "2022-10-30T22:21:48.020Z",
    "timezone": "Europe/Paris",
    "language": "fr",
    "country": "FR",
    "consent_all": true,
    "consent_personalization": true,
    "consent_marketing": true,
    "latitude": 123.123,
    "longitude": 123.123,
    "first_name": "John",
    "last_name": "Doe",
    "gender": "male",
    "birthday": "1980-01-01",
    "photo_url": "https://photo-url.com/photo.jpg",
    "email": "john@doe.com",
    "email_md5": "xxxxx",
    "email_sha1": "xxxxx",
    "email_sha256": "xxxxx",
    "telephone": "+33601010101",
    "address_line_1": "abc",
    "address_line_2": "abc",
    "city": "Paris",
    "region": "Ile de France",
    "postal_code": "75000",
    "state": "abc"
  }
}`
    },
    props.template
  )

  const goNext = () => {
    setTab('template')
  }

  const decorateContent = (content: any, data: any, template_macro_id: any) => {
    if (!content) {
      return
    }

    let parsedData = {}
    try {
      parsedData = JSON.parse(data)
    } catch (e: any) {}

    let macroContent = ''

    if (template_macro_id) {
      macros.forEach((x: any) => {
        if (x.id === template_macro_id) {
          macroContent = x.content
        }
      })
    }

    // console.log('template_macro_id', template_macro_id);
    // console.log('macroContent', macroContent);
    // console.log('parsedData', parsedData);

    Nunjucks.renderString(macroContent + ' ' + content, parsedData, (error, result) => {
      if (error) {
        setContentDecorated('Templating syntax is not valid...')
        return
      }
      // trim result to remove linebreaks introduced by macros sets
      if (result) setContentDecorated(result.trim())
    })
  }

  // layout height

  const doc = document.querySelector('.ant-drawer')
  const topbarHeight = 65
  const contentHeight = doc ? parseInt(window.getComputedStyle(doc).height) - topbarHeight : 0

  return (
    <Drawer
      title={<>{props.template ? 'Edit email template' : 'Create an email template'}</>}
      closable={true}
      keyboard={false}
      maskClosable={false}
      width={tab === 'settings' ? 960 : '95%'}
      open={true}
      onClose={() => props.setDrawserVisible(false)}
      className={CSS.drawerBodyNoPadding}
      // rootClassName={CSS.drawerNoTransition}
      extra={
        <div style={{ textAlign: 'right' }}>
          <Space>
            <Button type="link" loading={loading} onClick={() => props.setDrawserVisible(false)}>
              Cancel
            </Button>

            {tab === 'settings' && (
              <Button type="primary" onClick={goNext}>
                Next
              </Button>
            )}
            {tab === 'template' && (
              <Button type="primary" ghost onClick={() => setTab('settings')}>
                Previous
              </Button>
            )}

            {tab === 'template' && (
              <Button
                loading={loading}
                onClick={() => {
                  form
                    .validateFields()
                    .then((values: any) => {
                      console.log('values', values)
                      // compile html
                      if (values.engine === 'visual') {
                        const urlParams = {
                          utm_source: values.utm_source,
                          utm_medium: values.utm_medium,
                          utm_campaign: values.utm_campaign,
                          utm_content: values.id,
                          // replaced by the backend
                          // the utm_id contains the unique message.id sent to the user
                          utm_id: '{{ rmd_utm_id }}'
                        }

                        const result = ExportHTML(values.email.visual_editor_tree, urlParams)

                        if (result.errors && result.errors.length > 0) {
                          message.error(result.errors[0].formattedMessage)
                          return
                        }

                        values.email.content = result.html
                      } else {
                        values.email.visual_editor_tree = undefined
                      }

                      submitForm(values)
                    })
                    .catch((info) => {
                      // console.log('Validate Failed:', info)
                      if (info.errorFields) {
                        info.errorFields.forEach((field: any) => {
                          if (
                            [
                              'name',
                              'id',
                              'engine',
                              'email.from_address',
                              'email.from_name',
                              'email.subject',
                              'email.reply_to'
                            ].indexOf(field.name['0']) !== -1
                          ) {
                            setTab('settings')
                          }
                        })
                      }
                    })
                }}
                type="primary"
              >
                Save
              </Button>
            )}
          </Space>
        </div>
      }
    >
      <div className={tabsInHeader}>
        <Tabs
          activeKey={tab}
          centered
          onChange={(k) => setTab(k)}
          style={{ display: 'inline-block' }}
          items={[
            {
              key: 'settings',
              label: '1. Settings'
            },
            {
              key: 'template',
              label: '2. Template'
            }
          ]}
        />
      </div>
      <div style={{ position: 'relative' }}>
        <Form form={form} layout="vertical" initialValues={initialValues}>
          <Tabs
            activeKey={tab}
            centered
            onChange={(k) => setTab(k)}
            tabBarStyle={{ display: 'none' }}
            items={[
              {
                key: 'settings',
                label: '1. Settings',
                children: (
                  <div className={CSS.padding_a_l}>
                    <Row gutter={24}>
                      <Col span={8}>
                        <Form.Item name="name" label="Template name" rules={[{ required: true }]}>
                          <Input
                            placeholder="i.e: Newsletter ABC"
                            onChange={(e: any) => {
                              if (!props.template) {
                                const id = kebabCase(e.target.value)
                                form.setFieldsValue({ id: id })
                              }
                            }}
                          />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name="id"
                          label="Template ID (utm_content)"
                          rules={[
                            {
                              required: true,
                              type: 'string',
                              pattern: /^[a-z0-9]+(-[a-z0-9]+)*$/,
                              message: Messages.InvalidIdFormat
                            }
                          ]}
                        >
                          <Input
                            disabled={props.template ? true : false}
                            placeholder="i.e: newsletter-abc"
                          />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name="category"
                          label="Category"
                          rules={[{ required: true, type: 'string' }]}
                        >
                          <Select
                            placeholder="Select category"
                            disabled={props.category ? true : false}
                            options={[
                              {
                                value: 'transactional',
                                label: <Tag color="green">Transactional</Tag>
                              },
                              {
                                value: 'campaign',
                                label: <Tag color="purple">Campaign</Tag>
                              },
                              {
                                value: 'automation',
                                label: <Tag color="cyan">Automation</Tag>
                              },
                              {
                                value: 'other',
                                label: <Tag color="magenta">Other...</Tag>
                              }
                            ]}
                          />
                        </Form.Item>
                      </Col>
                    </Row>

                    <Form.Item
                      name="engine"
                      label="Editor"
                      rules={[{ required: true, type: 'string' }]}
                    >
                      <InfoRadioGroup
                        layout="vertical"
                        span={12}
                        options={[
                          {
                            key: 'visual',
                            title: <>Drag'n drop</>,
                            icon: <FontAwesomeIcon icon={faPalette} />,
                            content: (
                              <span>Visual drag'n drop editor, for non-technical users.</span>
                            )
                          },
                          {
                            key: 'code',
                            title: (
                              <>
                                Code{' '}
                                <Tag color="cyan" className={CSS.margin_l_m}>
                                  Coming soon
                                </Tag>
                              </>
                            ),
                            icon: <FontAwesomeIcon icon={faCode} />,
                            content: (
                              <span>
                                Code editor with templating engine, for maximum control over HTML.
                              </span>
                            ),
                            disabled: true
                          }
                        ]}
                        onChange={() => {
                          // triggers are unique for each kind, reset them when switching
                          // form.setFieldsValue({ triggers: [] })
                        }}
                      />
                    </Form.Item>

                    <Row gutter={24}>
                      <Col span={8}>
                        <Form.Item
                          name={['email', 'from_address']}
                          label="Sender email address"
                          rules={[{ required: true, type: 'email' }]}
                        >
                          <Input />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={['email', 'from_name']}
                          label="Sender name"
                          rules={[{ required: true, type: 'string' }]}
                        >
                          <Input />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={['email', 'reply_to']}
                          label="Reply to"
                          rules={[{ required: false, type: 'email' }]}
                        >
                          <Input />
                        </Form.Item>
                      </Col>
                    </Row>

                    <Form.Item
                      name={['email', 'subject']}
                      label="Email subject"
                      rules={[{ required: true, type: 'string' }]}
                    >
                      <Input placeholder="Templating markup allowed" />
                    </Form.Item>

                    <Divider plain className={CSS.padding_v_l}>
                      URL Tracking
                    </Divider>

                    {props.utmCampaign && (
                      <Alert
                        type="info"
                        showIcon
                        className={CSS.margin_b_l}
                        message="The utm_source / medium / campaign parameters are already defined at the Campaign level."
                      />
                    )}
                    <Row gutter={24}>
                      <Col span={8}>
                        <Form.Item
                          name="utm_source"
                          label="utm_source"
                          rules={[{ required: false, type: 'string' }]}
                        >
                          <Input
                            placeholder="business.com"
                            disabled={props.utmSource ? true : false}
                          />
                        </Form.Item>
                      </Col>

                      <Col span={8}>
                        <Form.Item
                          name="utm_medium"
                          label="utm_medium"
                          rules={[{ required: false, type: 'string' }]}
                        >
                          <Input placeholder="email" disabled={props.utmMedium ? true : false} />
                        </Form.Item>
                      </Col>

                      <Col span={8}>
                        <Form.Item
                          name="utm_campaign"
                          label="utm_campaign"
                          rules={[{ required: false, type: 'string' }]}
                        >
                          <Input disabled={props.utmCampaign ? true : false} />
                        </Form.Item>
                      </Col>
                    </Row>
                  </div>
                )
              },
              {
                key: 'template',
                label: '2. Template',
                children: (
                  <>
                    <Form.Item noStyle shouldUpdate={true}>
                      {({ getFieldValue }: any) => {
                        if (getFieldValue('engine') === 'visual') {
                          return (
                            <div>
                              <Form.Item noStyle name={['email', 'visual_editor_tree']}>
                                <Editor
                                  blockDefinitions={{
                                    root: RootBlockDefinition,
                                    column: ColumnBlockDefinition,
                                    oneColumn: OneColumnBlockDefinition,
                                    columns168: Columns168BlockDefinition,
                                    columns204: Columns204BlockDefinition,
                                    columns420: Columns420BlockDefinition,
                                    columns816: Columns816BlockDefinition,
                                    columns888: Columns888BlockDefinition,
                                    columns1212: Columns1212BlockDefinition,
                                    columns6666: Columns6666BlockDefinition,
                                    image: ImageBlockDefinition,
                                    divider: DividerBlockDefinition,
                                    openTracking: OpenTrackingBlockDefinition,
                                    button: ButtonBlockDefinition,
                                    text: TextBlockDefinition,
                                    heading: HeadingBlockDefinition
                                  }}
                                  savedBlocks={
                                    workspaceCtx.workspace.messaging_settings
                                      .email_template_blocks || []
                                  }
                                  templateDataValue={getFieldValue('test_data')}
                                  selectedBlockId={divider.id}
                                  value={rootBlock}
                                  onChange={() => {
                                    // console.log('new tree', newTree)
                                  }}
                                  renderSelectedBlockButtons={(props: SelectedBlockButtonsProp) => (
                                    <SelectedBlockButtons {...props} />
                                  )}
                                  deviceWidth={DesktopWidth}
                                  urlParams={{
                                    utm_source: getFieldValue('utm_source'),
                                    utm_medium: getFieldValue('utm_medium'),
                                    utm_campaign: getFieldValue('utm_campaign'),
                                    utm_content: getFieldValue('id')
                                    // utm_id: getFieldValue('id')
                                  }}
                                >
                                  <Layout form={form} macros={macros} height={contentHeight} />
                                </Editor>
                              </Form.Item>
                              {/*  hidden test_data field */}
                              <Form.Item
                                name="test_data"
                                style={{ display: 'none' }}
                                rules={[{ required: false, type: 'string' }]}
                              >
                                <Input />
                              </Form.Item>
                              <Form.Item
                                name="template_macro_id"
                                style={{ display: 'none' }}
                                rules={[{ required: false, type: 'string' }]}
                              >
                                <Input />
                              </Form.Item>
                            </div>
                          )
                        }

                        // if (getFieldValue('engine') === 'code') {
                        //   return (
                        //     <>
                        //       <Row gutter={24} className={CSS.margin_a_m}>
                        //         <Col span={12}>
                        //           <Tabs
                        //             tabBarExtraContent={
                        //               <div style={{ width: '250px' }}>
                        //                 <Tooltip title="Macros page">
                        //                   <div>
                        //                     <Form.Item noStyle name="template_macro_id">
                        //                       <Select
                        //                         style={{ width: '100%' }}
                        //                         dropdownMatchSelectWidth={false}
                        //                         allowClear={true}
                        //                         size="small"
                        //                         placeholder="Select macros page"
                        //                         options={macros.map((x: any) => {
                        //                           return { label: x.name, value: x.id }
                        //                         })}
                        //                       />
                        //                     </Form.Item>
                        //                   </div>
                        //                 </Tooltip>
                        //               </div>
                        //             }
                        //             defaultActiveKey={contentField}
                        //             onChange={(value: any) => {
                        //               setContentField(value)
                        //               decorateContent(
                        //                 form.getFieldValue(value),
                        //                 form.getFieldValue('test_data'),
                        //                 form.getFieldValue('template_macro_id')
                        //               )
                        //             }}
                        //           >
                        //             <Tabs.TabPane tab="HTML" key="content">
                        //               <Form.Item
                        //                 name={['email', 'content']}
                        //                 rules={[{ required: false, type: 'string' }]}
                        //               >
                        //                 <AceInput
                        //                   id="widgetContent"
                        //                   width="600px"
                        //                   height="300px"
                        //                   mode="nunjucks"
                        //                 />
                        //               </Form.Item>
                        //             </Tabs.TabPane>

                        //             <Tabs.TabPane tab="Text" key="text">
                        //               <Form.Item
                        //                 name={['email', 'text']}
                        //                 rules={[{ required: false, type: 'string' }]}
                        //               >
                        //                 <AceInput
                        //                   id="textContent"
                        //                   width="600px"
                        //                   height="300px"
                        //                   mode="nunjucks"
                        //                 />
                        //               </Form.Item>
                        //             </Tabs.TabPane>
                        //           </Tabs>

                        //           <Form.Item
                        //             label="Test data"
                        //             name="test_data"
                        //             validateFirst={true}
                        //             rules={[
                        //               {
                        //                 validator: (_xxx, value) => {
                        //                   // check if data is valid json
                        //                   try {
                        //                     if (JSON.parse(value)) {
                        //                     }
                        //                     return Promise.resolve(undefined)
                        //                   } catch (e: any) {
                        //                     return Promise.reject(
                        //                       'Your test variables is not a valid JSON object!'
                        //                     )
                        //                   }
                        //                 }
                        //               },
                        //               {
                        //                 required: false,
                        //                 type: 'object',
                        //                 transform: (value: any) => {
                        //                   try {
                        //                     const parsed = JSON.parse(value)
                        //                     return parsed
                        //                   } catch (e: any) {
                        //                     return value
                        //                   }
                        //                 }
                        //               }
                        //             ]}
                        //           >
                        //             <AceInput
                        //               id="test_data"
                        //               width="600px"
                        //               height="150px"
                        //               mode="json"
                        //             />
                        //           </Form.Item>
                        //           {/* <Form.Item
                        //           name="email.css_inlining"
                        //           label="CSS inlining"
                        //           rules={[{ required: false, type: 'boolean' }]}
                        //           valuePropName="checked"
                        //         >
                        //           <Switch />
                        //         </Form.Item> */}
                        //         </Col>

                        //         <Col span={12} className={CSS.borderLeft.solid1}>
                        //           <p className={CSS.padding_t_s}>Preview</p>

                        //           <Form.Item
                        //             noStyle
                        //             dependencies={[
                        //               ['email', 'content'],
                        //               ['email', 'text'],
                        //               'test_data',
                        //               'template_macro_id'
                        //             ]}
                        //           >
                        //             {({ getFieldValue }: any) => {
                        //               decorateContent(
                        //                 getFieldValue(contentField),
                        //                 getFieldValue('test_data'),
                        //                 getFieldValue('template_macro_id')
                        //               )
                        //               return (
                        //                 <IframeSandbox
                        //                   content={contentDecorated}
                        //                   sizeId="previewContent"
                        //                   id="templateCompiled"
                        //                 />
                        //               )
                        //             }}
                        //           </Form.Item>
                        //         </Col>
                        //       </Row>
                        //     </>
                        //   )
                        // }
                      }}
                    </Form.Item>
                  </>
                )
              }
            ]}
          />
        </Form>
      </div>
    </Drawer>
  )
}

export default ButtonUpsertEmailTemplate
