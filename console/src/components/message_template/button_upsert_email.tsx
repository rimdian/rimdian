import {
  Button,
  Col,
  Divider,
  Drawer,
  Form,
  Input,
  Row,
  Select,
  Space,
  Switch,
  Tabs,
  Tooltip,
  message
} from 'antd'
import { MessageTemplate } from './interfaces'
import { useState } from 'react'
import { cloneDeep, kebabCase } from 'lodash'
import Messages from 'utils/formMessages'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCode, faPalette } from '@fortawesome/free-solid-svg-icons'
import RootBlockDefinition from '../email_editor/UI/definitions/Root'
import OneColumnBlockDefinition from '../email_editor/UI/definitions/OneColumn'
import Columns168BlockDefinition from '../email_editor/UI/definitions/Columns168'
import Columns204BlockDefinition from '../email_editor/UI/definitions/Columns204'
import Columns420BlockDefinition from '../email_editor/UI/definitions/Columns420'
import Columns816BlockDefinition from '../email_editor/UI/definitions/Columns816'
import Columns888BlockDefinition from '../email_editor/UI/definitions/Columns888'
import Columns1212BlockDefinition from '../email_editor/UI/definitions/Columns1212'
import Columns6666BlockDefinition from '../email_editor/UI/definitions/Columns6666'
import ImageBlockDefinition from '../email_editor/UI/definitions/Image'
import DividerBlockDefinition from '../email_editor/UI/definitions/Divider'
import ButtonBlockDefinition from '../email_editor/UI/definitions/Button'
import TextBlockDefinition from '../email_editor/UI/definitions/Text'
import HeadingBlockDefinition from '../email_editor/UI/definitions/Heading'
import { Editor, SelectedBlockButtonsProp } from '../email_editor/Editor'
import { BlockInterface, BlockDefinitionInterface } from '../email_editor/Block'
import ColumnBlockDefinition from '../email_editor/UI/definitions/Column'
import { ExportHTML } from '../email_editor/UI/Preview'
import { Layout, DesktopWidth } from '../email_editor/UI/Layout'
import SelectedBlockButtons from '../email_editor/UI/SelectedBlockButtons'
import uuid from 'short-uuid'
import { css } from '@emotion/css'
import InfoRadioGroup from 'components/common/input_info_radio_group'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Nunjucks from 'nunjucks'
import AceInput from 'components/common/input_ace'
import IframeSandbox from 'components/email_editor/UI/Widgets/Iframe'
import CSS from 'utils/css'

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

interface ButtonUpsertTemplateProps {
  template?: MessageTemplate
  onSuccess?: () => void
  btnProps: any
  children: React.ReactNode
}

const ButtonUpsertTemplate = (props: ButtonUpsertTemplateProps) => {
  const [drawserVisible, setDrawserVisible] = useState(false)
  return (
    <>
      {drawserVisible && (
        <DrawerEmailTemplate
          template={props.template}
          setDrawserVisible={setDrawserVisible}
          onSuccess={props.onSuccess}
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
  onSuccess?: () => void
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

    props.setDrawserVisible(false)
  }

  // extract TLD from URL
  const extractTLD = (url: string) => {
    const hostname = new URL(url).hostname
    return hostname.split('.').slice(-2).join('.')
  }

  const initialValues = Object.assign(
    {
      utm_source: extractTLD(workspaceCtx.workspace.website_url),
      utm_medium: 'email',
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
    form
      .validateFields([
        // TODO
        // 'name',
        // 'kind',
        // 'segmentId',
        // 'notificationTopicId',
        // 'triggers',
        // 'limitExecPerUser',
      ])
      .then((values: any) => {
        // console.log('next values', values)
        setTab('template')
      })
      .catch(() => {})
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
      title={<>{props.template ? 'Edit template' : 'Create a template'}</>}
      closable={true}
      maskClosable={false}
      width={tab === 'settings' ? 900 : '95%'}
      open={true}
      onClose={() => props.setDrawserVisible(false)}
      className={CSS.drawerBodyNoPadding}
      // rootClassName={CSS.drawerNoTransition}
      extra={
        <div style={{ textAlign: 'right' }}>
          <Space>
            <Button loading={loading} onClick={() => props.setDrawserVisible(false)}>
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
                      // compile html
                      if (values.engine === 'visual') {
                        const result = ExportHTML(values.editorData)
                        if (result.errors && result.errors.length > 0) {
                          message.error(result.errors[0].formattedMessage)
                          return
                        }
                        values.content = result.html
                      } else {
                        values.editorData = undefined
                      }

                      submitForm(values)
                    })
                    .catch((info) => {
                      console.log('Validate Failed:', info)
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
                      <Col span={12}>
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
                      <Col span={12}>
                        <Form.Item
                          name="id"
                          label="Template ID"
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
                            title: <>Code</>,
                            icon: <FontAwesomeIcon icon={faCode} />,
                            content: (
                              <span>
                                Code editor with templating engine, for maximum control over HTML.
                              </span>
                            )
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

                    <Row gutter={24}>
                      <Col span={12}>
                        <Form.Item
                          name="utm_source"
                          label="utm_source"
                          rules={[{ required: false, type: 'string' }]}
                        >
                          <Input placeholder="business.com" />
                        </Form.Item>
                        <Form.Item
                          name="utm_medium"
                          label="utm_medium"
                          rules={[{ required: false, type: 'string' }]}
                        >
                          <Input placeholder="email" />
                        </Form.Item>
                      </Col>

                      <Col span={12}>
                        <Form.Item
                          name="utm_campaign"
                          label="utm_campaign"
                          rules={[{ required: false, type: 'string' }]}
                        >
                          <Input />
                        </Form.Item>
                        <Form.Item
                          name="utm_content"
                          label="utm_content"
                          rules={[{ required: false, type: 'string' }]}
                        >
                          <Input />
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
                                    button: ButtonBlockDefinition,
                                    text: TextBlockDefinition,
                                    heading: HeadingBlockDefinition
                                  }}
                                  savedBlocks={workspaceCtx.workspace.emailBlocks || []}
                                  templateData={getFieldValue('test_data')}
                                  selectedBlockId={divider.id}
                                  value={rootBlock}
                                  onChange={(_newTree) => {
                                    // console.log('new tree', newTree)
                                  }}
                                  renderSelectedBlockButtons={(props: SelectedBlockButtonsProp) => (
                                    <SelectedBlockButtons {...props} />
                                  )}
                                  deviceWidth={DesktopWidth}
                                >
                                  <Layout form={form} macros={macros} height={contentHeight} />
                                </Editor>
                              </Form.Item>
                            </div>
                          )
                        }

                        if (getFieldValue('engine') === 'code') {
                          return (
                            <>
                              <Row gutter={24} className={CSS.margin_a_m}>
                                <Col span={12}>
                                  <Tabs
                                    tabBarExtraContent={
                                      <div style={{ width: '250px' }}>
                                        <Tooltip title="Macros page">
                                          <div>
                                            <Form.Item noStyle name="template_macro_id">
                                              <Select
                                                style={{ width: '100%' }}
                                                dropdownMatchSelectWidth={false}
                                                allowClear={true}
                                                size="small"
                                                placeholder="Select macros page"
                                                options={macros.map((x: any) => {
                                                  return { label: x.name, value: x.id }
                                                })}
                                              />
                                            </Form.Item>
                                          </div>
                                        </Tooltip>
                                      </div>
                                    }
                                    defaultActiveKey={contentField}
                                    onChange={(value: any) => {
                                      setContentField(value)
                                      decorateContent(
                                        form.getFieldValue(value),
                                        form.getFieldValue('test_data'),
                                        form.getFieldValue('template_macro_id')
                                      )
                                    }}
                                  >
                                    <Tabs.TabPane tab="HTML" key="content">
                                      <Form.Item
                                        name={['email', 'content']}
                                        rules={[{ required: false, type: 'string' }]}
                                      >
                                        <AceInput
                                          id="widgetContent"
                                          width="600px"
                                          height="300px"
                                          mode="nunjucks"
                                        />
                                      </Form.Item>
                                    </Tabs.TabPane>

                                    <Tabs.TabPane tab="Text" key="text">
                                      <Form.Item
                                        name={['email', 'text']}
                                        rules={[{ required: false, type: 'string' }]}
                                      >
                                        <AceInput
                                          id="textContent"
                                          width="600px"
                                          height="300px"
                                          mode="nunjucks"
                                        />
                                      </Form.Item>
                                    </Tabs.TabPane>
                                  </Tabs>

                                  <Form.Item
                                    label="Test data"
                                    name="test_data"
                                    validateFirst={true}
                                    rules={[
                                      {
                                        validator: (_xxx, value) => {
                                          // check if data is valid json
                                          try {
                                            if (JSON.parse(value)) {
                                            }
                                            return Promise.resolve(undefined)
                                          } catch (e: any) {
                                            return Promise.reject(
                                              'Your test variables is not a valid JSON object!'
                                            )
                                          }
                                        }
                                      },
                                      {
                                        required: false,
                                        type: 'object',
                                        transform: (value: any) => {
                                          try {
                                            const parsed = JSON.parse(value)
                                            return parsed
                                          } catch (e: any) {
                                            return value
                                          }
                                        }
                                      }
                                    ]}
                                  >
                                    <AceInput
                                      id="test_data"
                                      width="600px"
                                      height="150px"
                                      mode="json"
                                    />
                                  </Form.Item>
                                  {/* <Form.Item
                                  name="email.css_inlining"
                                  label="CSS inlining"
                                  rules={[{ required: false, type: 'boolean' }]}
                                  valuePropName="checked"
                                >
                                  <Switch />
                                </Form.Item> */}
                                </Col>

                                <Col span={12} className={CSS.borderLeft.solid1}>
                                  <p className={CSS.padding_t_s}>Preview</p>

                                  <Form.Item
                                    noStyle
                                    dependencies={[
                                      ['email', 'content'],
                                      ['email', 'text'],
                                      'test_data',
                                      'template_macro_id'
                                    ]}
                                  >
                                    {({ getFieldValue }: any) => {
                                      decorateContent(
                                        getFieldValue(contentField),
                                        getFieldValue('test_data'),
                                        getFieldValue('template_macro_id')
                                      )
                                      return (
                                        <IframeSandbox
                                          content={contentDecorated}
                                          sizeId="previewContent"
                                          id="templateCompiled"
                                        />
                                      )
                                    }}
                                  </Form.Item>
                                </Col>
                              </Row>
                            </>
                          )
                        }
                      }}
                    </Form.Item>
                  </>
                )
              }
            ]}
          />
        </Form>
      </div>
      {/* <Form
        form={form}
        initialValues={initialValues}
        labelCol={{ span: 6 }}
        wrapperCol={{ span: 14 }}
        layout="horizontal"
        className={CSS.margin_a_m + ' ' + CSS.margin_b_xl}
      >
       
        <Form.Item noStyle shouldUpdate>
          {(funcs) => {
            const type = funcs.getFieldValue('type')

            return (
              <>
                {type === 'web' && (
                  <>
                    
                  </>
                )}
              </>
            )
          }}
        </Form.Item>
      </Form> */}
    </Drawer>
  )
}

export default ButtonUpsertTemplate
