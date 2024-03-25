/* eslint-disable no-script-url */
/* eslint-disable jsx-a11y/anchor-is-valid */
import React, { useState, useRef } from 'react'
import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import { Popover, Button, Form, InputNumber, Divider, Radio, Input, Select } from 'antd'
import BorderInputs from '../Widgets/BorderInputs'
import PaddingInputs from '../Widgets/PaddingInputs'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faAlignLeft,
  faAlignCenter,
  faAlignRight,
  faHandPointer
} from '@fortawesome/free-solid-svg-icons'
// import { MobileWidth } from '../Layout'
import ColorPickerInput from '../Widgets/ColorPicker'
import FontStyleInputs from '../Widgets/FontStyleInputs'
import { Fonts } from '../Widgets/ElementForms'

const ButtonBlockDefinition: BlockDefinitionInterface = {
  name: 'Button',
  kind: 'button',
  containsDraggables: false,
  isDraggable: true,
  draggableIntoGroup: 'column',
  isDeletable: true,
  defaultData: {
    wrapper: {
      align: 'center',
      paddingControl: 'all', // all, separate
      padding: '20px'
    },
    button: {
      backgroundColor: '#00BCD4',
      href: 'https://captainmetrics.com',
      text: 'Click me!',
      innerVerticalPadding: '10px',
      innerHorizontalPadding: '25px',
      borderControl: 'all', // all, separate
      borderColor: '#000000',
      borderWidth: '2px',
      borderStyle: 'none',
      borderRadius: '8px',
      width: 'auto',
      color: '#FFFFFF',
      fontFamily: Fonts[2].value,
      fontWeight: 600,
      fontSize: '15px',
      fontStyle: 'normal',
      textTransform: 'uppercase'
      // height: '40px',
    }
  },
  menuSettings: {},

  RenderSettings: (props: BlockRenderSettingsProps) => {
    const textInputRef = useRef<any>(null)
    const [text, setText] = useState(props.block.data.button.text)
    const [textModalVisible, setTextModalVisible] = useState(false)
    // console.log('img block is', block)
    // return <Input
    //     value={props.block.data.url}
    //     onChange={e => {
    //         const newBlock = { ...block }
    //         newBlock.data.url = e.target.value
    //         onUpdate(newBlock)
    //     }}
    // />

    return (
      <div className="cmeditor-padding-h-l">
        <Form.Item
          label="Content"
          className="cmeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Popover
            content={
              <>
                <Input
                  style={{ width: '100%' }}
                  onChange={(e) => setText(e.target.value)}
                  value={text}
                  size="small"
                  ref={textInputRef}
                />
                <Button
                  style={{ marginTop: '12px' }}
                  type="primary"
                  size="small"
                  block
                  onClick={() => {
                    props.block.data.button.text = text
                    props.updateTree(props.block.path, props.block)
                    setTextModalVisible(false)
                  }}
                  disabled={props.block.data.button.text === text}
                >
                  Save changes
                </Button>
              </>
            }
            title="Alternative text"
            trigger="click"
            visible={textModalVisible}
            onVisibleChange={(visible) => {
              setTextModalVisible(visible)
              setTimeout(() => {
                if (visible)
                  textInputRef.current!.focus({
                    cursor: 'start'
                  })
              }, 10)
            }}
          >
            {props.block.data.button.text === '' && (
              <Button type="primary" size="small" block>
                Set value
              </Button>
            )}
            {props.block.data.button.text !== '' && (
              <>
                {props.block.data.button.text} &nbsp;&nbsp;
                <span className="cmeditor-ui-link">update</span>
              </>
            )}
          </Popover>
        </Form.Item>

        <Form.Item
          label="Button URL"
          labelAlign="left"
          className="cmeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Input
            style={{ width: '100%' }}
            onChange={(e) => {
              props.block.data.button.href = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.button.href}
            size="small"
          />
        </Form.Item>

        <Form.Item
          label="Button color"
          labelAlign="left"
          className="cmeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <ColorPickerInput
            size="small"
            value={props.block.data.button.backgroundColor}
            onChange={(newColor) => {
              props.block.data.button.backgroundColor = newColor
              props.updateTree(props.block.path, props.block)
            }}
          />
        </Form.Item>

        <Form.Item
          label="Font family"
          labelAlign="left"
          className="cmeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Select
            size="small"
            options={Fonts}
            value={props.block.data.button.fontFamily}
            onChange={(value) => {
              props.block.data.button.fontFamily = value
              props.updateTree(props.block.path, props.block)
            }}
          />
        </Form.Item>

        <Form.Item
          label="Font style"
          labelAlign="left"
          className="cmeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <FontStyleInputs
            // size="small"
            styles={props.block.data.button}
            onChange={(updatedStyles: any) => {
              props.block.data.button = updatedStyles
              props.updateTree(props.block.path, props.block)
            }}
          />
        </Form.Item>

        <Divider />

        <Form.Item
          label="Width"
          labelAlign="left"
          className="cmeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Radio.Group
            style={{ width: '100%' }}
            value={props.block.data.button.width}
            optionType="button"
            size="small"
            onChange={(e) => {
              props.block.data.button.width = e.target.value
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
                props.block.data.button.width !== 'auto' ? props.block.data.button.width : '200px'
              }
            >
              <InputNumber
                style={{ height: '100%' }}
                bordered={false}
                value={parseInt(
                  props.block.data.button.width === 'auto' ? '200px' : props.block.data.button.width
                )}
                onChange={(value) => {
                  props.block.data.button.width = value + 'px'
                  props.updateTree(props.block.path, props.block)
                }}
                onClick={() => {
                  // switch focus to px
                  if (props.block.data.button.width === 'auto') {
                    props.block.data.button.width = '200px'
                    props.updateTree(props.block.path, props.block)
                  }
                }}
                defaultValue={parseInt(
                  props.block.data.button.width === 'auto' ? '200px' : props.block.data.button.width
                )}
                size="small"
                step={1}
                min={0}
                parser={(value: string | undefined) =>
                  parseInt(value ? value.replace('px', '') : '0')
                }
                formatter={(value) => value + 'px'}
              />
            </Radio.Button>
          </Radio.Group>
        </Form.Item>

        {/* <Form.Item label="Width" labelAlign="left" className="cmeditor-form-item-align-right" labelCol={{ span: 10 }} wrapperCol={{ span: 14 }}>
                <InputNumber
                    style={{ width: '100%' }}
                    value={parseInt(props.block.data.button.width || '100px')}
                    onChange={(value) => {
                        props.block.data.button.width = value + 'px'
                        props.updateTree(props.block.path, props.block)
                    }}
                    defaultValue={parseInt(props.block.data.button.width)}
                    size="small"
                    step={1}
                    min={0}
                    parser={(value: string) => parseInt(value.replace('px', ''))}
                    formatter={value => value + 'px'}
                />
            </Form.Item> */}

        <Form.Item
          label="Inner padding"
          labelAlign="left"
          className="cmeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Input.Group style={{ width: '100%' }} size="small" compact>
            <InputNumber
              style={{ width: '50%' }}
              value={parseInt(props.block.data.button.innerVerticalPadding || '10px')}
              onChange={(value) => {
                props.block.data.button.innerVerticalPadding = value + 'px'
                props.updateTree(props.block.path, props.block)
              }}
              defaultValue={parseInt(props.block.data.button.innerVerticalPadding)}
              size="small"
              step={1}
              min={0}
              parser={(value: string | undefined) => {
                if (value === undefined) return 0
                return parseInt(value.replace('⇅', '').replace('px', ''))
              }}
              formatter={(value) => '⇅  ' + value + 'px'}
            />
            <InputNumber
              style={{ width: '50%' }}
              value={parseInt(props.block.data.button.innerHorizontalPadding || '25px')}
              onChange={(value) => {
                props.block.data.button.innerHorizontalPadding = value + 'px'
                props.updateTree(props.block.path, props.block)
              }}
              defaultValue={parseInt(props.block.data.button.innerHorizontalPadding)}
              size="small"
              step={1}
              min={0}
              parser={(value: string | undefined) => {
                if (value === undefined) return 0
                return parseInt(value.replace('⇆', '').replace('px', ''))
              }}
              formatter={(value) => '⇆  ' + value + 'px'}
            />
          </Input.Group>
        </Form.Item>

        {/* <Form.Item label="Height" labelAlign="left" className="cmeditor-form-item-align-right" labelCol={{ span: 10 }} wrapperCol={{ span: 14 }}>
                <InputNumber
                    style={{ width: '100%' }}
                    value={parseInt(props.block.data.button.height)}
                    onChange={(value) => {
                        props.block.data.button.height = value + 'px'
                        props.updateTree(props.block.path, props.block)
                    }}
                    defaultValue={props.block.data.button.height}
                    size="small"
                    step={1}
                    min={0}
                    parser={(value: string) => parseInt(value.replace('px', ''))}
                    formatter={value => value + 'px'}
                />
            </Form.Item> */}

        <Form.Item
          label="Align"
          labelAlign="left"
          className="cmeditor-form-item-align-right"
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
          label="Padding control"
          labelAlign="left"
          className="cmeditor-form-item-align-right"
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
              className="cmeditor-form-item-align-right"
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
                  if (value === undefined) return 0
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
              className="cmeditor-form-item-align-right"
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

        <Divider />

        <Form.Item
          label="Border control"
          labelAlign="left"
          className="cmeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Radio.Group
            style={{ width: '100%' }}
            onChange={(e) => {
              props.block.data.button.borderControl = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.button.borderControl}
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
          className="cmeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <InputNumber
            style={{ width: '100%' }}
            value={parseInt(props.block.data.button.borderRadius || '0px')}
            onChange={(value) => {
              props.block.data.button.borderRadius = value + 'px'
              props.updateTree(props.block.path, props.block)
            }}
            defaultValue={props.block.data.button.borderRadius}
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

        {props.block.data.button.borderControl === 'all' && (
          <>
            <Form.Item
              label="Borders"
              labelAlign="left"
              className="cmeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.button}
                propertyPrefix="border"
                onChange={(updatedStyles: any) => {
                  props.block.data.button = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
          </>
        )}

        {props.block.data.button.borderControl === 'separate' && (
          <>
            <Form.Item
              label="Border top"
              labelAlign="left"
              className="cmeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.button}
                propertyPrefix="borderTop"
                onChange={(updatedStyles: any) => {
                  props.block.data.button = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
            <Form.Item
              label="Border right"
              labelAlign="left"
              className="cmeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.button}
                propertyPrefix="borderRight"
                onChange={(updatedStyles: any) => {
                  props.block.data.button = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
            <Form.Item
              label="Border bottom"
              labelAlign="left"
              className="cmeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.button}
                propertyPrefix="borderBottom"
                onChange={(updatedStyles: any) => {
                  props.block.data.button = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
            <Form.Item
              label="Border left"
              labelAlign="left"
              className="cmeditor-form-item-align-right"
              labelCol={{ span: 10 }}
              wrapperCol={{ span: 14 }}
            >
              <BorderInputs
                styles={props.block.data.button}
                propertyPrefix="borderLeft"
                onChange={(updatedStyles: any) => {
                  props.block.data.button = updatedStyles
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
    const wrapperStyles: any = {
      textAlign: props.block.data.wrapper.align
    }
    const buttonStyles: any = {
      display: 'inline-block',
      textAlign: 'center',
      width: props.block.data.button.width,
      backgroundColor: props.block.data.button.backgroundColor,
      color: props.block.data.button.color,
      lineHeight: '120%',
      fontFamily: props.block.data.button.fontFamily,
      fontWeight: props.block.data.button.fontWeight,
      fontSize: props.block.data.button.fontSize,
      fontStyle: props.block.data.button.fontStyle,
      textTransform: props.block.data.button.textTransform,
      padding:
        props.block.data.button.innerVerticalPadding +
        ' ' +
        props.block.data.button.innerHorizontalPadding
    }

    if (props.block.data.button.borderControl === 'all') {
      if (
        props.block.data.button.borderStyle !== 'none' &&
        props.block.data.button.borderWidth &&
        props.block.data.button.borderColor
      ) {
        buttonStyles.border =
          props.block.data.button.borderWidth +
          ' ' +
          props.block.data.button.borderStyle +
          ' ' +
          props.block.data.button.borderColor
      }
    }

    // if (props.block.data.button.height && props.block.data.button.height !== '') {
    //     buttonStyles.height = props.block.data.button.height
    // }

    if (props.block.data.button.borderRadius && props.block.data.button.borderRadius !== '0px') {
      buttonStyles.borderRadius = props.block.data.button.borderRadius
    }

    if (props.block.data.button.borderControl === 'separate') {
      if (
        props.block.data.button.borderTopStyle !== 'none' &&
        props.block.data.button.borderTopWidth &&
        props.block.data.button.borderTopColor
      ) {
        buttonStyles.borderTop =
          props.block.data.button.borderTopWidth +
          ' ' +
          props.block.data.button.borderTopStyle +
          ' ' +
          props.block.data.button.borderTopColor
      }

      if (
        props.block.data.button.borderRightStyle !== 'none' &&
        props.block.data.button.borderRightWidth &&
        props.block.data.button.borderRightColor
      ) {
        buttonStyles.borderRight =
          props.block.data.button.borderRightWidth +
          ' ' +
          props.block.data.button.borderRightStyle +
          ' ' +
          props.block.data.button.borderRightColor
      }

      if (
        props.block.data.button.borderBottomStyle !== 'none' &&
        props.block.data.button.borderBottomWidth &&
        props.block.data.button.borderBottomColor
      ) {
        buttonStyles.borderBottom =
          props.block.data.button.borderBottomWidth +
          ' ' +
          props.block.data.button.borderBottomStyle +
          ' ' +
          props.block.data.button.borderBottomColor
      }

      if (
        props.block.data.button.borderLeftStyle !== 'none' &&
        props.block.data.button.borderLeftWidth &&
        props.block.data.button.borderLeftColor
      ) {
        buttonStyles.borderLeft =
          props.block.data.button.borderLeftWidth +
          ' ' +
          props.block.data.button.borderLeftStyle +
          ' ' +
          props.block.data.button.borderLeftColor
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
        <a style={buttonStyles} href="javascript:void(0)">
          {props.block.data.button.text}
        </a>
      </div>
    )
  },

  // transformer: (block: BlockInterface) => {
  //     return <div>TODO transformer for {block.kind}</div>
  // },

  renderMenu: (_blockDefinition: BlockDefinitionInterface) => {
    return (
      <div className="cmeditor-ui-block cmeditor-square">
        <div className="cmeditor-ui-block-icon">
          <div
            style={{
              border: '1px solid #1890ff',
              borderRadius: '4px',
              height: '24px',
              margin: '7px 8px',
              fontSize: '14px'
            }}
          >
            <FontAwesomeIcon icon={faHandPointer} />
          </div>
        </div>
        Button
      </div>
    )
  }
}

export default ButtonBlockDefinition
