import React, { useState, useRef, MutableRefObject } from 'react'
import { BlockDefinitionInterface, BlockInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import {
  Popover,
  Button,
  Form,
  InputNumber,
  Divider,
  Radio,
  Input,
  Switch,
  Modal,
  message
} from 'antd'
import BorderInputs from '../Widgets/BorderInputs'
import PaddingInputs from '../Widgets/PaddingInputs'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faAlignLeft,
  faAlignCenter,
  faAlignRight,
  faImage
} from '@fortawesome/free-solid-svg-icons'
import { MobileWidth } from '../Layout'
import short from 'short-uuid'

interface ImageURLProps {
  block: BlockInterface
  updateTree: any
}

const ImageURL = (props: ImageURLProps) => {
  const altInputRef = useRef<any>(null)
  const [alt, setAlt] = useState(props.block.data.image.alt)
  const [altModalVisible, setAltModalVisible] = useState(false)

  return (
    <Form.Item
      label="Alternative text"
      className="rmdeditor-form-item-align-right"
      labelCol={{ span: 10 }}
      wrapperCol={{ span: 14 }}
    >
      <Popover
        content={
          <>
            <Input
              style={{ width: '100%' }}
              onChange={(e) => setAlt(e.target.value)}
              value={alt}
              size="small"
              ref={altInputRef}
            />
            <Button
              style={{ marginTop: '12px' }}
              type="primary"
              size="small"
              block
              onClick={() => {
                props.block.data.image.alt = alt
                props.updateTree(props.block.path, props.block)
                setAltModalVisible(false)
              }}
              disabled={props.block.data.image.alt === alt}
            >
              Save changes
            </Button>
          </>
        }
        title="Alternative text"
        trigger="click"
        open={altModalVisible}
        onOpenChange={(visible) => {
          setAltModalVisible(visible)
          setTimeout(() => {
            if (visible)
              altInputRef.current!.focus({
                cursor: 'start'
              })
          }, 10)
        }}
      >
        {props.block.data.image.alt === '' && (
          <Button type="primary" size="small" block>
            Set value
          </Button>
        )}
        {props.block.data.image.alt !== '' && (
          <>
            {props.block.data.image.alt} &nbsp;&nbsp;
            <span className="rmdeditor-ui-link">update</span>
          </>
        )}
      </Popover>
    </Form.Item>
  )
}

const ClickURL = (props: ImageURLProps) => {
  const hrefInputRef = useRef<any>(null)
  const [href, setHref] = useState(props.block.data.image.href)
  const [hrefModalVisible, setHrefModalVisible] = useState(false)
  return (
    <Form.Item
      label="Click URL"
      className="rmdeditor-form-item-align-right"
      labelCol={{ span: 10 }}
      wrapperCol={{ span: 14 }}
    >
      <Popover
        content={
          <>
            <Input
              style={{ width: '100%' }}
              onChange={(e) => setHref(e.target.value)}
              value={href}
              size="small"
              ref={hrefInputRef}
              placeholder="https://www..."
            />
            <Button
              style={{ marginTop: '12px' }}
              type="primary"
              size="small"
              block
              onClick={() => {
                props.block.data.image.href = href
                props.updateTree(props.block.path, props.block)
                setHrefModalVisible(false)
              }}
              disabled={props.block.data.image.href === href}
            >
              Save changes
            </Button>
          </>
        }
        title="Click URL"
        trigger="click"
        open={hrefModalVisible}
        onOpenChange={(visible) => {
          setHrefModalVisible(visible)
          setTimeout(() => {
            if (visible)
              hrefInputRef.current!.focus({
                cursor: 'start'
              })
          }, 10)
        }}
      >
        {!props.block.data.image.href && (
          <Button type="primary" size="small" block>
            Set value
          </Button>
        )}
        {props.block.data.image.href && (
          <>
            {props.block.data.image.href} &nbsp;&nbsp;
            <span className="rmdeditor-ui-link">update</span>
          </>
        )}
      </Popover>
    </Form.Item>
  )
}

// the UploadButton useState cant reside directly in RenderSettings()
// because it's not a proper React functional component
// const UploadButton = ({ block, updateTree }) => {
//     const projectCtx: ProjectContextValue = useProjectContext()
//     const [fileManagerVisible, setFileManagerVisible] = useState(false)
//     const [selectedImageURL, setSelectedImageURL] = useState<string | undefined>(undefined)
//     const provider: MutableRefObject<FileProvider> = useRef(new FirestoreFileProvider({
//         FirebaseApp: projectCtx.firebaseApp.current,
//         resolveRootNodeID: (): string => {
//             return projectCtx.currentProject.id
//         },
//         resolveUploadPath: (item: Item): string => {
//             item.id = short.uuid()
//             const uploadPath = '/' + projectCtx.currentProject.id + item.path.replace(new RegExp('~', 'g'), '/') + '/' + item.id
//             return uploadPath
//         },
//         // collectionsPrefix: 'fs_',
//         getNodeMetadata: (_node: TreeNode): any => {
//             return {
//                 userId: projectCtx.firebaseUser.uid,
//                 organizationId: projectCtx.currentOrganization.id,
//                 projectId: projectCtx.currentProject.id,
//             }
//         },
//         // filterKey: 'projectId',
//         // filterValue: 'test',
//     }))

//     const filters: ItemFilter[] = []

//     return <>
//         {fileManagerVisible && <Modal
//             visible={true}
//             title="Select or upload"
//             width={1100}
//             bodyStyle={{ background: '#F3F6FC' }}
//             onOk={() => {
//                 if (selectedImageURL && selectedImageURL !== block.data.image.src) {
//                     block.data.image.src = selectedImageURL
//                     setSelectedImageURL(undefined)
//                     setFileManagerVisible(false)
//                     updateTree(block.path, block)
//                 }
//             }}
//             onCancel={() => setFileManagerVisible(false)}
//             destroyOnClose={true}
//             okText="Use image"
//             okButtonProps={{
//                 disabled: !selectedImageURL,
//             }}
//         >
//             <div style={{ height: 500 }}>
//                 <FileManager
//                     itemFilters={filters}
//                     onError={(error) => {
//                         console.error(error)
//                         message.error(error)
//                     }}
//                     fileProvider={provider.current}
//                 >
//                 <Layout {...{
//                     visible: true,
//                     height: 500,
//                     acceptFileType: 'images/*',
//                     onError: (e) => message.error(e),
//                     onSelect: (items: Item[]) => {
//                         // console.log('onSelect', items)
//                         if (items && items.length) {
//                             setSelectedImageURL(items[0].url)
//                         } else {
//                             setSelectedImageURL(undefined)
//                         }
//                     },
//                     withSelection: true,
//                     multiple: false,
//                     acceptItem: (item: Item) => {
//                         return item.contentType.includes('image')
//                     }
//                 }} />
//             </FileManager>
//         </div>
//         </Modal>
// }
// <Button type="primary" size="small" block onClick={() => setFileManagerVisible(true)}>Select or upload</Button>
//    </>
// }
const ImageBlockDefinition: BlockDefinitionInterface = {
  name: 'Image',
  kind: 'image',
  containsDraggables: false,
  isDraggable: true,
  draggableIntoGroup: 'column',
  isDeletable: true,
  defaultData: {
    wrapper: {
      align: 'center',
      paddingControl: 'separate', // all, separate
      paddingTop: '20px',
      paddingBottom: '20px'
    },
    image: {
      borderControl: 'all', // all, separate
      borderColor: '#000000',
      borderWidth: '2px',
      borderStyle: 'none',
      fullWidthOnMobile: false,
      src: 'https://images.unsplash.com/photo-1432889490240-84df33d47091?ixid=MnwxMjA3fDB8MHxzZWFyY2h8MTZ8fHRyb3BpY2FsfGVufDB8fDB8fA%3D%3D&ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=60',
      alt: '',
      href: '',
      width: '100%',
      height: 'auto'
    }
  },
  menuSettings: {},

  RenderSettings: (props: BlockRenderSettingsProps) => {
    // console.log('img block is', props.block)

    return (
      <div className="rmdeditor-padding-h-l">
        <Form.Item
          label="Image"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          {/* <UploadButton block={props.block} updateTree={props.updateTree} /> */}
          TODO
        </Form.Item>

        <ImageURL block={props.block} updateTree={props.updateTree} />

        <ClickURL block={props.block} updateTree={props.updateTree} />

        <Form.Item
          label="Align"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Radio.Group
            style={{ width: '100%' }}
            onChange={(e) => {
              props.block.data.wrapper.align = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.wrapper.align}
            optionType="button"
            size="small"
          >
            <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="left">
              <FontAwesomeIcon icon={faAlignLeft} />
            </Radio.Button>
            <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="center">
              <FontAwesomeIcon icon={faAlignCenter} />
            </Radio.Button>
            <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="right">
              <FontAwesomeIcon icon={faAlignRight} />
            </Radio.Button>
          </Radio.Group>
        </Form.Item>

        <Divider />

        <Form.Item
          valuePropName="checked"
          label="Full width on mobile"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Switch
            onChange={(value) => {
              props.block.data.image.fullWidthOnMobile = value
              props.updateTree(props.block.path, props.block)
            }}
            checked={props.block.data.image.fullWidthOnMobile || false}
            size="small"
          />
        </Form.Item>

        <Form.Item
          label="Width"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Radio.Group
            style={{ width: '100%' }}
            value={props.block.data.image.width}
            optionType="button"
            size="small"
            onChange={(e) => {
              props.block.data.image.width = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
          >
            <Radio.Button value="100%" style={{ width: '40%', textAlign: 'center' }}>
              100%
            </Radio.Button>
            <label
              style={{
                display: 'inline-block',
                height: '24px',
                lineHeight: '22px',
                width: '20%',
                textAlign: 'center'
              }}
            >
              or
            </label>
            <Radio.Button
              style={{ width: '40%' }}
              value={
                props.block.data.image.width !== '100%' ? props.block.data.image.width : '200px'
              }
            >
              <InputNumber
                style={{ width: '100%' }}
                bordered={false}
                value={parseInt(props.block.data.image.width || '100px')}
                onChange={(value) => {
                  props.block.data.image.width = value + 'px'
                  props.updateTree(props.block.path, props.block)
                }}
                onClick={() => {
                  // switch focus to px
                  if (props.block.data.image.width === '100%') {
                    props.block.data.image.width = '100px'
                    props.updateTree(props.block.path, props.block)
                  }
                }}
                defaultValue={parseInt(props.block.data.image.width)}
                size="small"
                step={1}
                min={0}
                parser={(value: string | undefined) => {
                  return value ? parseInt(value.replace('px', '')) : 0
                }}
                formatter={(value) => value + 'px'}
              />
            </Radio.Button>
          </Radio.Group>
        </Form.Item>

        <Form.Item
          label="Height"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Radio.Group
            style={{ width: '100%' }}
            value={props.block.data.image.height}
            optionType="button"
            size="small"
            onChange={(e) => {
              props.block.data.image.height = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
          >
            <Radio.Button value="auto" style={{ width: '40%', textAlign: 'center' }}>
              auto
            </Radio.Button>
            <label
              style={{
                display: 'inline-block',
                height: '24px',
                lineHeight: '22px',
                width: '20%',
                textAlign: 'center'
              }}
            >
              or
            </label>
            <Radio.Button
              style={{ width: '40%' }}
              value={
                props.block.data.image.height !== 'auto' ? props.block.data.image.height : '100px'
              }
            >
              <InputNumber
                style={{ height: '100%' }}
                bordered={false}
                value={parseInt(props.block.data.image.height || '100px')}
                onChange={(value) => {
                  props.block.data.image.height = value + 'px'
                  props.updateTree(props.block.path, props.block)
                }}
                onClick={() => {
                  // switch focus to px
                  if (props.block.data.image.height === 'auto') {
                    props.block.data.image.height = '100px'
                    props.updateTree(props.block.path, props.block)
                  }
                }}
                defaultValue={parseInt(
                  props.block.data.image.height === 'auto' ? '100px' : props.block.data.image.height
                )}
                size="small"
                step={1}
                min={0}
                parser={(value: string | undefined) => {
                  if (!value) {
                    return 0
                  }
                  return parseInt(value.replace('px', ''))
                }}
                formatter={(value) => value + 'px'}
              />
            </Radio.Button>
          </Radio.Group>
        </Form.Item>

        <Divider />

        <Form.Item
          label="Border control"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Radio.Group
            style={{ width: '100%' }}
            onChange={(e) => {
              props.block.data.image.borderControl = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.image.borderControl}
            optionType="button"
            size="small"
            // buttonStyle="solid"
          >
            <Radio.Button style={{ width: '50%', textAlign: 'center' }} value="all">
              All
            </Radio.Button>
            <Radio.Button style={{ width: '50%', textAlign: 'center' }} value="separate">
              Separate
            </Radio.Button>
          </Radio.Group>
        </Form.Item>

        <Form.Item
          label="Border radius"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <InputNumber
            style={{ width: '100%' }}
            value={parseInt(props.block.data.image.borderRadius || '0px')}
            onChange={(value) => {
              props.block.data.image.borderRadius = value + 'px'
              props.updateTree(props.block.path, props.block)
            }}
            defaultValue={props.block.data.image.borderRadius}
            size="small"
            step={1}
            min={0}
            parser={(value: string | undefined) => {
              // if (['▭'].indexOf(value)) {
              //     value = value.substring(1)
              // }
              if (!value) {
                return 0
              }
              return parseInt(value.replace('px', ''))
            }}
            // formatter={value => '▭  ' + value + 'px'}
            formatter={(value) => value + 'px'}
          />
        </Form.Item>

        {props.block.data.image.borderControl === 'all' && (
          <>
            <Form.Item
              label="Borders"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.image}
                propertyPrefix="border"
                onChange={(updatedStyles: any) => {
                  props.block.data.image = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
          </>
        )}

        {props.block.data.image.borderControl === 'separate' && (
          <>
            <Form.Item
              label="Border top"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.image}
                propertyPrefix="borderTop"
                onChange={(updatedStyles: any) => {
                  props.block.data.image = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
            <Form.Item
              label="Border right"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.image}
                propertyPrefix="borderRight"
                onChange={(updatedStyles: any) => {
                  props.block.data.image = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
            <Form.Item
              label="Border bottom"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.image}
                propertyPrefix="borderBottom"
                onChange={(updatedStyles: any) => {
                  props.block.data.image = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
            <Form.Item
              label="Border left"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.image}
                propertyPrefix="borderLeft"
                onChange={(updatedStyles: any) => {
                  props.block.data.image = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
          </>
        )}

        <Divider />

        <Form.Item
          label="Padding control"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Radio.Group
            style={{ width: '100%' }}
            onChange={(e) => {
              props.block.data.wrapper.paddingControl = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.wrapper.paddingControl}
            optionType="button"
            size="small"
            // buttonStyle="solid"
          >
            <Radio.Button style={{ width: '50%', textAlign: 'center' }} value="all">
              All
            </Radio.Button>
            <Radio.Button style={{ width: '50%', textAlign: 'center' }} value="separate">
              Separate
            </Radio.Button>
          </Radio.Group>
        </Form.Item>

        {props.block.data.wrapper.paddingControl === 'all' && (
          <>
            <Form.Item
              label="Paddings"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <InputNumber
                style={{ width: '100%' }}
                value={parseInt(props.block.data.wrapper.padding || '0px')}
                onChange={(value) => {
                  props.block.data.wrapper.padding = value + 'px'
                  props.updateTree(props.block.path, props.block)
                }}
                size="small"
                step={1}
                min={0}
                parser={(value: string | undefined) => {
                  // if (['▭'].indexOf(value)) {
                  //     value = value.substring(1)
                  // }
                  if (!value) {
                    return 0
                  }
                  return parseInt(value.replace('px', ''))
                }}
                // formatter={value => '▭  ' + value + 'px'}
                formatter={(value) => value + 'px'}
              />
            </Form.Item>
          </>
        )}

        {props.block.data.wrapper.paddingControl === 'separate' && (
          <>
            <Form.Item
              label="Paddings"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <PaddingInputs
                styles={props.block.data.wrapper}
                onChange={(updatedStyles: any) => {
                  props.block.data.wrapper = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
          </>
        )}
      </div>
    )
  },

  renderEditor: (props: BlockEditorRendererProps) => {
    const wrapperStyles: any = {}
    const imageStyles: any = {
      width: props.block.data.image.width,
      height: props.block.data.image.height
    }

    wrapperStyles.textAlign = props.block.data.wrapper.align

    if (props.block.data.image.borderControl === 'all') {
      if (
        props.block.data.image.borderStyle !== 'none' &&
        props.block.data.image.borderWidth &&
        props.block.data.image.borderColor
      ) {
        imageStyles.border =
          props.block.data.image.borderWidth +
          ' ' +
          props.block.data.image.borderStyle +
          ' ' +
          props.block.data.image.borderColor
      }
    }

    if (props.block.data.image.width !== '100%') {
      imageStyles.width = props.block.data.image.width
    }

    if (props.block.data.image.height !== 'auto') {
      imageStyles.height = props.block.data.image.height
    }

    if (props.block.data.image.fullWidthOnMobile === true && props.deviceWidth <= MobileWidth) {
      imageStyles.width = '100%'
      imageStyles.height = 'auto'
    }

    if (props.block.data.image.borderRadius && props.block.data.image.borderRadius !== '0px') {
      imageStyles.borderRadius = props.block.data.image.borderRadius
    }

    if (props.block.data.image.borderControl === 'separate') {
      if (
        props.block.data.image.borderTopStyle !== 'none' &&
        props.block.data.image.borderTopWidth &&
        props.block.data.image.borderTopColor
      ) {
        imageStyles.borderTop =
          props.block.data.image.borderTopWidth +
          ' ' +
          props.block.data.image.borderTopStyle +
          ' ' +
          props.block.data.image.borderTopColor
      }

      if (
        props.block.data.image.borderRightStyle !== 'none' &&
        props.block.data.image.borderRightWidth &&
        props.block.data.image.borderRightColor
      ) {
        imageStyles.borderRight =
          props.block.data.image.borderRightWidth +
          ' ' +
          props.block.data.image.borderRightStyle +
          ' ' +
          props.block.data.image.borderRightColor
      }

      if (
        props.block.data.image.borderBottomStyle !== 'none' &&
        props.block.data.image.borderBottomWidth &&
        props.block.data.image.borderBottomColor
      ) {
        imageStyles.borderBottom =
          props.block.data.image.borderBottomWidth +
          ' ' +
          props.block.data.image.borderBottomStyle +
          ' ' +
          props.block.data.image.borderBottomColor
      }

      if (
        props.block.data.image.borderLeftStyle !== 'none' &&
        props.block.data.image.borderLeftWidth &&
        props.block.data.image.borderLeftColor
      ) {
        imageStyles.borderLeft =
          props.block.data.image.borderLeftWidth +
          ' ' +
          props.block.data.image.borderLeftStyle +
          ' ' +
          props.block.data.image.borderLeftColor
      }
    }

    if (props.block.data.wrapper.paddingControl === 'all') {
      if (props.block.data.wrapper.padding && props.block.data.wrapper.padding !== '0px') {
        wrapperStyles.padding = props.block.data.wrapper.padding
      }
    }

    if (props.block.data.wrapper.paddingControl === 'separate') {
      if (props.block.data.wrapper.paddingTop && props.block.data.wrapper.paddingTop !== '0px') {
        wrapperStyles.paddingTop = props.block.data.wrapper.paddingTop
      }
      if (
        props.block.data.wrapper.paddingRight &&
        props.block.data.wrapper.paddingRight !== '0px'
      ) {
        wrapperStyles.paddingRight = props.block.data.wrapper.paddingRight
      }
      if (
        props.block.data.wrapper.paddingBottom &&
        props.block.data.wrapper.paddingBottom !== '0px'
      ) {
        wrapperStyles.paddingBottom = props.block.data.wrapper.paddingBottom
      }
      if (props.block.data.wrapper.paddingLeft && props.block.data.wrapper.paddingLeft !== '0px') {
        wrapperStyles.paddingLeft = props.block.data.wrapper.paddingLeft
      }
    }

    return (
      <div style={wrapperStyles}>
        <img
          style={imageStyles}
          src={props.block.data.image.src}
          alt={props.block.data.image.alt}
        />
      </div>
    )
  },

  // transformer: (block: BlockInterface) => {
  //     return <div>TODO transformer for {block.kind}</div>
  // },

  renderMenu: (_blockDefinition: BlockDefinitionInterface) => {
    return (
      <div className="rmdeditor-ui-block rmdeditor-square">
        <div className="rmdeditor-ui-block-icon">
          <FontAwesomeIcon icon={faImage} />
        </div>
        Image
      </div>
    )
  }

  // deserialize (json: any) {

  //     // children can contain other definitions
  //     // they are deserialized at top level
  //     // block.children = json.children

  //     const block: BlockInterface = {
  //         kind: this.kind,
  //         id: json.id,
  //         path: json.path,
  //         data: json.data,
  //         children: [],
  //     }

  //     return block
  // }

  // serialize (block: BlockInterface) {
  //     // children can contain other definitions
  //     // they are deserialized at top level
  //     // block.children = json.children
  //     return {
  //         kind: block.kind,
  //         id: block.id,
  //         path: block.path,
  //         data: block.data,
  //     }
  // }
}

export default ImageBlockDefinition
