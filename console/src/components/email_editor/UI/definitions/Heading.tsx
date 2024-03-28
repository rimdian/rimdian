import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import { Form, Radio } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faAlignLeft, faAlignCenter, faAlignRight } from '@fortawesome/free-solid-svg-icons'
import ColorPickerInput from '../Widgets/ColorPicker'
import MyEditor, { EditorDataToReact } from '../Widgets/MyEditor'
import { cloneDeep } from 'lodash'
import ElementForms from '../Widgets/ElementForms'

const HeadingBlockDefinition: BlockDefinitionInterface = {
  name: 'Heading',
  kind: 'heading',
  containsDraggables: false,
  isDraggable: true,
  draggableIntoGroup: 'column',
  isDeletable: true,
  defaultData: {
    type: 'h1',
    align: 'left',
    width: '100%',
    editorData: [
      {
        type: 'h1',
        children: [{ text: 'Heading' }]
      }
    ]
  },
  menuSettings: {},

  RenderSettings: (props: BlockRenderSettingsProps) => {
    const root = props.tree

    return (
      <>
        <div className="rmdeditor-padding-h-l">
          <Form.Item
            label="Type"
            labelAlign="left"
            className="rmdeditor-form-item-align-right"
            labelCol={{ span: 10 }}
            wrapperCol={{ span: 14 }}
          >
            <Radio.Group
              style={{ width: '100%' }}
              onChange={(e) => {
                props.block.data.type = e.target.value
                props.block.data.editorData[0].type = props.block.data.type
                props.updateTree(props.block.path, props.block)
              }}
              value={props.block.data.type}
              optionType="button"
              size="small"
            >
              <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="h1">
                H1
              </Radio.Button>
              <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="h2">
                H2
              </Radio.Button>
              <Radio.Button style={{ width: '33.33%', textAlign: 'center' }} value="h3">
                H3
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
        </div>

        <div className="rmdeditor-ui-menu-title">{props.block.data.type} global settings</div>

        <div className="rmdeditor-padding-h-l">
          {props.block.data.type === 'h1' && (
            <ElementForms block={root} updateTree={props.updateTree} element="h1" />
          )}
          {props.block.data.type === 'h2' && (
            <ElementForms block={root} updateTree={props.updateTree} element="h2" />
          )}
          {props.block.data.type === 'h3' && (
            <ElementForms block={root} updateTree={props.updateTree} element="h3" />
          )}
        </div>
      </>
    )
  },

  renderEditor: (props: BlockEditorRendererProps) => {
    const root = props.tree
    // const editorRef = useRef(null)

    // const handleSave = async () => {
    //     const savedData = await editorRef.current.save();
    // }

    const wrapperStyles: any = {
      position: 'relative'
    }

    wrapperStyles.textAlign = props.block.data.align

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

    const elementStyles = cloneDeep(root.data.styles)

    if (elementStyles.paragraph.paddingControl === 'separate') {
      elementStyles.paragraph.padding = 0
    }
    if (elementStyles['h1'].paddingControl === 'separate') {
      elementStyles['h1'].padding = 0
    }
    if (elementStyles['h2'].paddingControl === 'separate') {
      elementStyles['h2'].padding = 0
    }
    if (elementStyles['h3'].paddingControl === 'separate') {
      elementStyles['h3'].padding = 0
    }

    // console.log('VALUE', props.block.data.editorData)

    const isFocused = props.selectedBlockId === props.block.id

    return (
      <div style={wrapperStyles}>
        {!isFocused && EditorDataToReact(props.block.data.editorData, elementStyles)}
        {isFocused && (
          <MyEditor
            // type={props.block.data.type}
            toolbarButtons={['bold', 'italic', 'underlined']}
            styles={elementStyles}
            onChange={(value) => {
              const newBlock = cloneDeep(props.block)
              newBlock.data.editorData = value
              props.updateTree(newBlock.path, newBlock)
            }}
            value={props.block.data.editorData}
            isFocused={props.selectedBlockId === props.block.id}
          />
        )}
      </div>
    )
  },

  renderMenu: (_blockDefinition: BlockDefinitionInterface) => {
    return (
      <div className="rmdeditor-ui-block square">
        <div className="rmdeditor-ui-block-icon">
          <div
            style={{
              fontFamily: '"Times New Roman", Times, serif',
              fontSize: '28px',
              lineHeight: '28px',
              margin: '5px 0'
            }}
          >
            T
          </div>
        </div>
        Heading
      </div>
    )
  }
}

export default HeadingBlockDefinition
