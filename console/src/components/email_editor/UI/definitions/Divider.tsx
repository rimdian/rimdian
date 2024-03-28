import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import { Form, InputNumber, Divider, Radio } from 'antd'
import BorderInputs from '../Widgets/BorderInputs'
import PaddingInputs from '../Widgets/PaddingInputs'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faAlignLeft, faAlignCenter, faAlignRight } from '@fortawesome/free-solid-svg-icons'
import ColorPickerInput from '../Widgets/ColorPicker'

const DividerBlockDefinition: BlockDefinitionInterface = {
  name: 'Divider',
  kind: 'divider',
  containsDraggables: false,
  isDraggable: true,
  draggableIntoGroup: 'column',
  isDeletable: true,
  defaultData: {
    align: 'center',
    paddingControl: 'all', // all, separate
    padding: '20px',
    borderColor: '#B0BEC5',
    borderWidth: '1px',
    borderStyle: 'solid',
    width: '100%'
  },
  menuSettings: {},

  RenderSettings: (props: BlockRenderSettingsProps) => {
    return (
      <div className="rmdeditor-padding-h-l">
        {/* <Form> */}
        <Form.Item
          label="Border"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <BorderInputs
            styles={props.block.data}
            propertyPrefix="border"
            onChange={(updatedStyles: any) => {
              props.block.data = updatedStyles
              props.updateTree(props.block.path, props.block)
            }}
            required={true}
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
            value={props.block.data.width}
            optionType="button"
            size="small"
            onChange={(e) => {
              props.block.data.width = e.target.value
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
              value={props.block.data.width !== '100%' ? props.block.data.width : '200px'}
            >
              <InputNumber
                style={{ width: '100%' }}
                bordered={false}
                value={parseInt(props.block.data.width || '100px')}
                onChange={(value) => {
                  props.block.data.width = value + 'px'
                  props.updateTree(props.block.path, props.block)
                }}
                onClick={() => {
                  // switch focus to px
                  if (props.block.data.width === '100%') {
                    props.block.data.width = '100px'
                    props.updateTree(props.block.path, props.block)
                  }
                }}
                defaultValue={parseInt(props.block.data.width)}
                size="small"
                step={1}
                min={0}
                parser={(value: string | undefined) => {
                  if (value === undefined) return 0
                  return parseInt(value.replace('px', ''))
                }}
                formatter={(value) => value + 'px'}
              />
            </Radio.Button>
          </Radio.Group>
        </Form.Item>

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
              props.block.data.align = e.target.value
              props.updateTree(props.block.path, props.block)
            }}
            value={props.block.data.align}
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

        <Form.Item
          label="Background color"
          labelAlign="left"
          className="rmdeditor-form-item-align-right"
          labelCol={{ span: 10 }}
          wrapperCol={{ span: 14 }}
        >
          <ColorPickerInput
            size="small"
            value={props.block.data.backgroundColor}
            onChange={(newColor) => {
              props.block.data.backgroundColor = newColor
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
                value={parseInt(props.block.data.padding || '0px')}
                onChange={(value) => {
                  props.block.data.padding = value + 'px'
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
                styles={props.block.data}
                onChange={(updatedStyles: any) => {
                  props.block.data = updatedStyles
                  props.updateTree(props.block.path, props.block)
                }}
              />
            </Form.Item>
          </>
        )}
        {/* </Form> */}
      </div>
    )
  },

  renderEditor: (props: BlockEditorRendererProps) => {
    const wrapperStyles: any = {
      lineHeight: '1px',
      fontSize: '1px'
    }
    const paragraphStyles: any = {
      display: 'inline-block',
      lineHeight: '1px',
      fontSize: '1px',
      margin: '0 auto',
      width: props.block.data.width
    }

    wrapperStyles.textAlign = props.block.data.align

    paragraphStyles.borderTop =
      props.block.data.borderWidth +
      ' ' +
      props.block.data.borderStyle +
      ' ' +
      props.block.data.borderColor

    if (props.block.data.width !== '100%') {
      paragraphStyles.width = props.block.data.width
    }

    if (props.block.data.paddingControl === 'all') {
      if (props.block.data.padding && props.block.data.padding !== '0px') {
        wrapperStyles.padding = props.block.data.padding
      }
    }

    if (props.block.data.backgroundColor && props.block.data.backgroundColor !== '') {
      wrapperStyles.backgroundColor = props.block.data.backgroundColor
    }

    if (props.block.data.paddingControl === 'separate') {
      if (props.block.data.paddingTop && props.block.data.paddingTop !== '0px') {
        wrapperStyles.paddingTop = props.block.data.paddingTop
      }
      if (props.block.data.paddingRight && props.block.data.paddingRight !== '0px') {
        wrapperStyles.paddingRight = props.block.data.paddingRight
      }
      if (props.block.data.paddingBottom && props.block.data.paddingBottom !== '0px') {
        wrapperStyles.paddingBottom = props.block.data.paddingBottom
      }
      if (props.block.data.paddingLeft && props.block.data.paddingLeft !== '0px') {
        wrapperStyles.paddingLeft = props.block.data.paddingLeft
      }
    }

    return (
      <div style={wrapperStyles}>
        <p style={paragraphStyles}></p>
      </div>
    )
  },

  // transformer: (block: BlockInterface) => {
  //     return <div>TODO transformer for {block.kind}</div>
  // },

  renderMenu: (_blockDefinition: BlockDefinitionInterface) => {
    return (
      <div className="rmdeditor-ui-block square">
        <div className="rmdeditor-ui-block-icon">
          <div
            style={{ backgroundColor: '#1890ff', height: '2px', margin: '18px 12px 18px 12px' }}
          ></div>
        </div>
        Divider
      </div>
    )
  }
}

export default DividerBlockDefinition
