import { ReactNode } from 'react'
import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import ColorPickerInput from '../Widgets/ColorPicker'
import BorderInputs from '../Widgets/BorderInputs'
import PaddingInputs from '../Widgets/PaddingInputs'
import { Form, InputNumber, Divider, Radio } from 'antd'

const ColumnBlockDefinition: BlockDefinitionInterface = {
  name: 'Column',
  kind: 'column',
  containsDraggables: true,
  containerGroup: 'column',
  isDraggable: false,
  isDeletable: false,
  defaultData: {
    paddingControl: 'all', // all, separate
    borderControl: 'all', // all, separate
    styles: {
      verticalAlign: 'top',
      minHeight: '30px'
    }
  },
  menuSettings: {},

  RenderSettings: (props: BlockRenderSettingsProps) => {
    // console.log('render settings', block.data)

    return (
      <div className="rmdeditor-padding-h-l">
        <Form.Item
          label="Vertical align"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <Radio.Group
            style={{ width: '100%' }}
            onChange={(e) => {
              props.block.data.styles.verticalAlign = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.styles.verticalAlign}
            optionType="button"
            size="small"
          >
            <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="top">
              Top
            </Radio.Button>
            <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="middle">
              Middle
            </Radio.Button>
            <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="bottom">
              Bottom
            </Radio.Button>
          </Radio.Group>
        </Form.Item>

        <Divider />

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
    const styles: any = {
      height: '100%' // content level, should match other columns heights
    }

    if (props.block.data.styles.verticalAlign)
      styles.verticalAlign = props.block.data.styles.verticalAlign

    if (props.block.data.styles.backgroundColor)
      styles.backgroundColor = props.block.data.styles.backgroundColor

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
          props.block.data.styles.borderTopColor
      }
    }

    return <div style={styles}>{content}</div>
  }
}

export default ColumnBlockDefinition
