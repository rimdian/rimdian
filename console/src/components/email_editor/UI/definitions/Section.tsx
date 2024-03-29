import { ReactNode } from 'react'
import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import ColorPickerInput from '../Widgets/ColorPicker'
import BorderInputs from '../Widgets/BorderInputs'
import PaddingInputs from '../Widgets/PaddingInputs'
import { MobileWidth } from '../Layout'
import { Form, InputNumber, Divider, Radio, Input } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faAlignLeft,
  faAlignCenter,
  faAlignRight,
  faAlignJustify
} from '@fortawesome/free-solid-svg-icons'

const SectionBlockDefinition: BlockDefinitionInterface = {
  name: 'Section',
  kind: 'section',
  containsDraggables: false,
  isDraggable: true,
  draggableIntoGroup: 'root',
  isDeletable: true,
  defaultData: {
    columnsOnMobile: false,
    stackColumnsAtWidth: MobileWidth,
    backgroundType: 'color', // color / image
    paddingControl: 'all', // all, separate
    borderControl: 'all', // all, separate
    styles: {
      textAlign: 'center',
      backgroundRepeat: 'repeat', // MJML default
      padding: '30px',
      borderWidth: '0px',
      borderStyle: 'none',
      borderColor: '#000000'
      // backgroundImage: 'https://images.unsplash.com/photo-1507525428034-b723cf961d3e?ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&ixlib=rb-1.2.1&auto=format&fit=crop&w=1353&q=80',
    }
  },
  menuSettings: {},
  children: [],

  RenderSettings: (props: BlockRenderSettingsProps) => {
    // console.log('render settings', block.data)

    return (
      <div className="rmdeditor-padding-h-l">
        <Form.Item
          label="Text align"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Radio.Group
            style={{ width: '100%' }}
            onChange={(e) => {
              props.block.data.styles.textAlign = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.styles.textAlign}
            optionType="button"
            size="small"
            // buttonStyle="solid"
          >
            <Radio.Button style={{ width: '25%', textAlign: 'center' }} value="left">
              <FontAwesomeIcon icon={faAlignLeft} />
            </Radio.Button>
            <Radio.Button style={{ width: '25%', textAlign: 'center' }} value="center">
              <FontAwesomeIcon icon={faAlignCenter} />
            </Radio.Button>
            <Radio.Button style={{ width: '25%', textAlign: 'center' }} value="right">
              <FontAwesomeIcon icon={faAlignRight} />
            </Radio.Button>
            <Radio.Button style={{ width: '25%', textAlign: 'center' }} value="justify">
              <FontAwesomeIcon icon={faAlignJustify} />
            </Radio.Button>
          </Radio.Group>
        </Form.Item>

        {props.block.children.length > 0 && (
          <Form.Item
            label="Columns on mobile"
            labelAlign="left"
            className="rmdeditor-form-item-align-right"
            labelCol={{ span: 10 }}
            wrapperCol={{ span: 14 }}
          >
            <Radio.Group
              style={{ width: '100%', textAlign: 'center' }}
              size="small"
              value={props.block.data.columnsOnMobile}
              onChange={(e) => {
                props.block.data.columnsOnMobile = e.target.value
                props.updateTree(props.block.path, props.block)
              }}
            >
              <Radio.Button style={{ width: '50%' }} value={false}>
                Stack
              </Radio.Button>
              <Radio.Button style={{ width: '50%' }} value={true}>
                Keep
              </Radio.Button>
            </Radio.Group>
          </Form.Item>
        )}

        <Divider />

        <Form.Item
          label="Background type"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Radio.Group
            style={{ width: '100%' }}
            onChange={(e) => {
              props.block.data.backgroundType = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.backgroundType}
            optionType="button"
            size="small"
            // buttonStyle="solid"
          >
            <Radio.Button style={{ width: '50%', textAlign: 'center' }} value="color">
              Color
            </Radio.Button>
            <Radio.Button style={{ width: '50%', textAlign: 'center' }} value="image">
              Image
            </Radio.Button>
          </Radio.Group>
        </Form.Item>

        {props.block.data.backgroundType === 'color' && (
          <>
            <Form.Item
              label="Background color"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <ColorPickerInput
                size="small"
                value={props.block.data.styles.backgroundColor}
                onChange={(newColor) => {
                  props.block.data.styles.backgroundColor = newColor
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
          </>
        )}

        {/* 
            <Form.Item label="Background image" labelAlign="left" className="rmdeditor-form-item-align-right" labelCol={{ span: 10 }} wrapperCol={{ span: 14 }}>
                <Switch checked={props.block.data.backgroundType} onChange={(checked) => {
                    props.block.data.backgroundType = checked
                    props.updateTree(props.block.path, props.block)
                }} />
            </Form.Item> */}

        {props.block.data.backgroundType === 'image' && (
          <>
            <Form.Item
              label="Image URL"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <Input
                allowClear={true}
                size="small"
                placeholder="https://..."
                value={props.block.data.styles.backgroundImage}
                onChange={(e) => {
                  props.block.data.styles.backgroundImage = e.target.value
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>

            <Form.Item
              label="Background size"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <Radio.Group
                style={{ width: '100%' }}
                onChange={(e) => {
                  props.block.data.styles.backgroundSize = e.target.value
                  props.updateTree(props.block.path, props.block)
                }}
                value={props.block.data.styles.backgroundSize}
                optionType="button"
                size="small"
              >
                <Radio.Button
                  style={{ width: '50%', textAlign: 'center' }}
                  value="cover"
                  onClick={() => {
                    // reset field if was selected
                    if (props.block.data.styles.backgroundSize === 'cover') {
                      // requeue event to process after onChange
                      setTimeout(() => {
                        props.block.data.styles.backgroundSize = undefined
                        props.updateTree(props.block.path, props.block)
                      }, 0)
                    }
                  }}
                >
                  Cover
                </Radio.Button>
                <Radio.Button
                  style={{ width: '50%', textAlign: 'center' }}
                  value="contain"
                  onClick={() => {
                    // reset field if was selected
                    if (props.block.data.styles.backgroundSize === 'contain') {
                      // requeue event to process after onChange
                      setTimeout(() => {
                        props.block.data.styles.backgroundSize = undefined
                        props.updateTree(props.block.path, props.block)
                      }, 0)
                    }
                  }}
                >
                  Contain
                </Radio.Button>
              </Radio.Group>
            </Form.Item>

            <Form.Item
              label="Background repeat"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <Radio.Group
                style={{ width: '100%' }}
                onChange={(e) => {
                  props.block.data.styles.backgroundRepeat = e.target.value
                  props.updateTree(props.block.path, props.block)
                }}
                value={props.block.data.styles.backgroundRepeat}
                optionType="button"
                size="small"
              >
                <Radio.Button
                  style={{ width: '50%', textAlign: 'center' }}
                  value="no-repeat"
                  onClick={() => {
                    // reset field if was selected
                    if (props.block.data.styles.backgroundRepeat === 'no-repeat') {
                      // requeue event to process after onChange
                      setTimeout(() => {
                        props.block.data.styles.backgroundRepeat = undefined
                        props.updateTree(props.block.path, props.block)
                      }, 0)
                    }
                  }}
                >
                  None
                </Radio.Button>

                <Radio.Button
                  style={{ width: '25%', textAlign: 'center' }}
                  value="repeat-x"
                  onClick={() => {
                    // reset field if was selected
                    if (props.block.data.styles.backgroundRepeat === 'repeat-x') {
                      // requeue event to process after onChange
                      setTimeout(() => {
                        props.block.data.styles.backgroundRepeat = undefined
                        props.updateTree(props.block.path, props.block)
                      }, 0)
                    }
                  }}
                >
                  &rarr;
                </Radio.Button>

                <Radio.Button
                  style={{ width: '25%', textAlign: 'center' }}
                  value="repeat-y"
                  onClick={() => {
                    // reset field if was selected
                    if (props.block.data.styles.backgroundRepeat === 'repeat-y') {
                      // requeue event to process after onChange
                      setTimeout(() => {
                        props.block.data.styles.backgroundRepeat = undefined
                        props.updateTree(props.block.path, props.block)
                      }, 0)
                    }
                  }}
                >
                  &darr;
                </Radio.Button>
              </Radio.Group>
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
              props.block.data.paddingControl = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.paddingControl}
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

        {props.block.data.paddingControl === 'all' && (
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
                value={parseInt(props.block.data.styles.padding || '0px')}
                onChange={(value) => {
                  props.block.data.styles.padding = value + 'px'
                  props.updateTree(props.block.path, props.block)
                }}
                defaultValue={props.block.data.styles.padding}
                size="small"
                step={1}
                min={0}
                parser={(value: string | undefined) => {
                  // if (['▭'].indexOf(value)) {
                  //     value = value.substring(1)
                  // }
                  if (value === undefined) return 0
                  return parseInt(value.replace('px', ''))
                }}
                // formatter={value => '▭  ' + value + 'px'}
                formatter={(value) => value + 'px'}
              />
            </Form.Item>
          </>
        )}

        {props.block.data.paddingControl === 'separate' && (
          <>
            <Form.Item
              label="Paddings"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <PaddingInputs
                styles={props.block.data.styles}
                onChange={(updatedStyles: any) => {
                  props.block.data.styles = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
          </>
        )}

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
              props.block.data.borderControl = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.borderControl}
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
            value={parseInt(props.block.data.styles.borderRadius || '0px')}
            onChange={(value) => {
              props.block.data.styles.borderRadius = value + 'px'
              props.updateTree(props.block.path, props.block)
            }}
            defaultValue={props.block.data.styles.borderRadius}
            size="small"
            step={1}
            min={0}
            parser={(value: string | undefined) => {
              // if (['▭'].indexOf(value)) {
              //     value = value.substring(1)
              // }
              if (value === undefined) return 0
              return parseInt(value.replace('px', ''))
            }}
            // formatter={value => '▭  ' + value + 'px'}
            formatter={(value) => value + 'px'}
          />
        </Form.Item>

        {props.block.data.borderControl === 'all' && (
          <>
            <Form.Item
              label="Borders"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.styles}
                propertyPrefix="border"
                onChange={(updatedStyles: any) => {
                  props.block.data.styles = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
          </>
        )}

        {props.block.data.borderControl === 'separate' && (
          <>
            <Form.Item
              label="Border top"
              labelAlign="left"
              className="rmdeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.styles}
                propertyPrefix="borderTop"
                onChange={(updatedStyles: any) => {
                  props.block.data.styles = updatedStyles
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
                styles={props.block.data.styles}
                propertyPrefix="borderRight"
                onChange={(updatedStyles: any) => {
                  props.block.data.styles = updatedStyles
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
                styles={props.block.data.styles}
                propertyPrefix="borderBottom"
                onChange={(updatedStyles: any) => {
                  props.block.data.styles = updatedStyles
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
                styles={props.block.data.styles}
                propertyPrefix="borderLeft"
                onChange={(updatedStyles: any) => {
                  props.block.data.styles = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
          </>
        )}
      </div>
    )
  },

  renderEditor: (props: BlockEditorRendererProps, content: ReactNode) => {
    const styles: any = {}

    if (props.block.data.styles.textAlign) styles.textAlign = props.block.data.styles.textAlign

    if (props.block.data.backgroundType === 'color') {
      if (props.block.data.styles.backgroundColor)
        styles.backgroundColor = props.block.data.styles.backgroundColor
    }

    if (props.block.data.backgroundType === 'image') {
      if (props.block.data.styles.backgroundImage)
        styles.backgroundImage = 'url("' + props.block.data.styles.backgroundImage + '")'

      if (props.block.data.styles.backgroundSize) {
        styles.backgroundSize = props.block.data.styles.backgroundSize
      }
      if (props.block.data.styles.backgroundRepeat) {
        styles.backgroundRepeat = props.block.data.styles.backgroundRepeat
      }
    }

    if (props.block.data.borderControl === 'all') {
      if (
        props.block.data.styles.borderStyle !== 'none' &&
        props.block.data.styles.borderWidth &&
        props.block.data.styles.borderColor
      ) {
        styles.border =
          props.block.data.styles.borderWidth +
          ' ' +
          props.block.data.styles.borderStyle +
          ' ' +
          props.block.data.styles.borderColor
      }
    }

    if (props.block.data.styles.borderRadius && props.block.data.styles.borderRadius !== '0px') {
      styles.borderRadius = props.block.data.styles.borderRadius
    }

    if (props.block.data.borderControl === 'separate') {
      if (
        props.block.data.styles.borderTopStyle !== 'none' &&
        props.block.data.styles.borderTopWidth &&
        props.block.data.styles.borderTopColor
      ) {
        styles.borderTop =
          props.block.data.styles.borderTopWidth +
          ' ' +
          props.block.data.styles.borderTopStyle +
          ' ' +
          props.block.data.styles.borderTopColor
      }

      if (
        props.block.data.styles.borderRightStyle !== 'none' &&
        props.block.data.styles.borderRightWidth &&
        props.block.data.styles.borderRightColor
      ) {
        styles.borderRight =
          props.block.data.styles.borderRightWidth +
          ' ' +
          props.block.data.styles.borderRightStyle +
          ' ' +
          props.block.data.styles.borderRightColor
      }

      if (
        props.block.data.styles.borderBottomStyle !== 'none' &&
        props.block.data.styles.borderBottomWidth &&
        props.block.data.styles.borderBottomColor
      ) {
        styles.borderBottom =
          props.block.data.styles.borderBottomWidth +
          ' ' +
          props.block.data.styles.borderBottomStyle +
          ' ' +
          props.block.data.styles.borderBottomColor
      }

      if (
        props.block.data.styles.borderLeftStyle !== 'none' &&
        props.block.data.styles.borderLeftWidth &&
        props.block.data.styles.borderLeftColor
      ) {
        styles.borderLeft =
          props.block.data.styles.borderLeftWidth +
          ' ' +
          props.block.data.styles.borderLeftStyle +
          ' ' +
          props.block.data.styles.borderLeftColor
      }
    }

    if (props.block.data.paddingControl === 'all') {
      if (props.block.data.styles.padding && props.block.data.styles.padding !== '0px') {
        styles.padding = props.block.data.styles.padding
      }
    }

    if (props.block.data.paddingControl === 'separate') {
      if (props.block.data.styles.paddingTop && props.block.data.styles.paddingTop !== '0px') {
        styles.paddingTop = props.block.data.styles.paddingTop
      }
      if (props.block.data.styles.paddingRight && props.block.data.styles.paddingRight !== '0px') {
        styles.paddingRight = props.block.data.styles.paddingRight
      }
      if (
        props.block.data.styles.paddingBottom &&
        props.block.data.styles.paddingBottom !== '0px'
      ) {
        styles.paddingBottom = props.block.data.styles.paddingBottom
      }
      if (props.block.data.styles.paddingLeft && props.block.data.styles.paddingLeft !== '0px') {
        styles.paddingLeft = props.block.data.styles.paddingLeft
      }
    }

    return <div style={styles}>{content}</div>
  },

  renderMenu: (blockDefinition: BlockDefinitionInterface) => {
    return (
      <div className="rmdeditor-menu-definition">
        {blockDefinition.name} - {blockDefinition.draggableIntoGroup}
      </div>
    )
  }
}

export default SectionBlockDefinition
