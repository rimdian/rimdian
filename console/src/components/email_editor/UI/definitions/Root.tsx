import { ReactNode } from 'react'
import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import { Form, InputNumber } from 'antd'
import ColorPickerInput from '../Widgets/ColorPicker'
import { Fonts } from '../Widgets/ElementForms'

const defaultWidth = '600px'

const RootBlockDefinition: BlockDefinitionInterface = {
  name: 'Root',
  kind: 'root',
  containsDraggables: true,
  containerGroup: 'root',
  isDraggable: false,
  isDeletable: false,
  defaultData: {
    styles: {
      // minHeight: '50px'
      body: {
        width: defaultWidth,
        margin: '0 auto',
        backgroundColor: '#FFFFFF'
      },
      h1: {
        color: '#000000',
        fontSize: '34px',
        fontStyle: 'normal',
        fontWeight: 400,
        paddingControl: 'separate', // all, separate
        padding: '0px',
        paddingTop: '0px',
        paddingRight: '0px',
        paddingBottom: '10px',
        paddingLeft: '0px',
        margin: 0,
        fontFamily: Fonts[2].value
      },
      h2: {
        color: '#000000',
        fontSize: '28px',
        fontStyle: 'normal',
        fontWeight: 400,
        paddingControl: 'separate', // all, separate
        padding: '0px',
        paddingTop: '0px',
        paddingRight: '0px',
        paddingBottom: '10px',
        paddingLeft: '0px',
        margin: 0,
        fontFamily: Fonts[2].value
      },
      h3: {
        color: '#000000',
        fontSize: '22px',
        fontStyle: 'normal',
        fontWeight: 400,
        paddingControl: 'separate', // all, separate
        padding: '0px',
        paddingTop: '0px',
        paddingRight: '0px',
        paddingBottom: '10px',
        paddingLeft: '0px',
        margin: 0,
        fontFamily: Fonts[2].value
      },
      paragraph: {
        color: '#000000',
        fontSize: '16px',
        fontStyle: 'normal',
        fontWeight: 400,
        paddingControl: 'separate', // all, separate
        padding: '0px',
        paddingTop: '0px',
        paddingRight: '0px',
        paddingBottom: '10px',
        paddingLeft: '0px',
        margin: 0,
        fontFamily: Fonts[2].value
      },
      hyperlink: {
        color: '#4e6cff',
        textDecoration: 'none',
        fontFamily: Fonts[2].value,
        fontSize: '16px',
        fontWeight: 400,
        fontStyle: 'normal',
        textTransform: 'none'
      }
    }
  },
  menuSettings: {},

  RenderSettings: (props: BlockRenderSettingsProps) => {
    // console.log('render settings', props)

    return (
      <>
        <div className="rmdeditor-padding-h-l">
          <Form.Item
            label="Email width"
            labelAlign="left"
            className="rmdeditor-form-item-align-right"
            labelCol={{ span: 12 }}
            wrapperCol={{ span: 12 }}
          >
            <InputNumber
              size="small"
              style={{ width: '100%' }}
              step={1}
              // min={200}
              value={parseInt(props.block.data.styles.body.width)}
              onChange={(newValue) => {
                props.block.data.styles.body.width = newValue + 'px'
                props.updateTree(props.block.path, props.block)
              }}
              formatter={(value) => value + 'px'}
              parser={(value) => parseInt((value || defaultWidth).replace('px', ''))}
            />
          </Form.Item>

          <Form.Item
            label="Background color"
            labelAlign="left"
            className="rmdeditor-form-item-align-right"
            labelCol={{ span: 12 }}
            wrapperCol={{ span: 12 }}
          >
            <ColorPickerInput
              size="small"
              value={props.block.data.styles.body.backgroundColor}
              onChange={(newColor) => {
                props.block.data.styles.body.backgroundColor = newColor
                props.updateTree(props.block.path, props.block)
              }}
            />
          </Form.Item>
        </div>
      </>
    )
  },

  renderEditor: (props: BlockEditorRendererProps, content: ReactNode) => {
    const styles = {
      margin: '0 auto',
      width: props.block.data.styles.body.width
    }

    if (parseInt(props.block.data.styles.body.width || 0) > props.deviceWidth) {
      styles.width = props.deviceWidth + 'px'
    }
    return (
      <div
        style={{
          paddingTop: '56px',
          minHeight: '100vh',
          backgroundColor: props.block.data.styles.body.backgroundColor
        }}
      >
        <div style={styles}>{content}</div>
      </div>
    )
  }
}

export default RootBlockDefinition
