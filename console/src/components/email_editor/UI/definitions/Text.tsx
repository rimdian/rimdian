import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import { Form, Radio } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faAlignLeft, faAlignCenter, faAlignRight } from '@fortawesome/free-solid-svg-icons'
import ColorPickerInput from '../Widgets/ColorPicker'
import MyEditor, { EditorDataToReact } from '../Widgets/MyEditor'
import { cloneDeep } from 'lodash'
import ElementForms from '../Widgets/ElementForms'
import FontStyleInputs from '../Widgets/FontStyleInputs'
import RootBlockDefinition from './Root'

const TextBlockDefinition: BlockDefinitionInterface = {
  name: 'Text',
  kind: 'text',
  containsDraggables: false,
  isDraggable: true,
  draggableIntoGroup: 'column',
  isDeletable: true,
  defaultData: {
    align: 'left',
    width: '100%',
    hyperlinkStyles: RootBlockDefinition.defaultData.styles.hyperlink,
    editorData: [
      {
        type: 'paragraph',
        children: [{ text: 'A line of text in a paragraph.' }]
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

          <Form.Item
            label="Hyperlink style"
            labelAlign="left"
            className="rmdeditor-form-item-align-right"
            labelCol={{ span: 10 }}
            wrapperCol={{ span: 14 }}
          >
            <FontStyleInputs
              styles={props.block.data.hyperlinkStyles}
              onChange={(updatedStyles: any) => {
                props.block.data.hyperlinkStyles = updatedStyles
                props.updateTree(props.block.path, props.block)
              }}
            />
          </Form.Item>
        </div>

        <div className="rmdeditor-ui-menu-title">Paragraphs global settings</div>

        <div className="rmdeditor-padding-h-l">
          <ElementForms block={root} updateTree={props.updateTree} element="paragraph" />
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

    elementStyles.hyperlink = props.block.data.hyperlinkStyles

    // console.log('selected: ' + props.selectedBlockId, props.selectedBlockId === props.block.id)
    // console.log('props.block.data.editorData', props.block.data.editorData)

    const isFocused = props.selectedBlockId === props.block.id

    return (
      <div style={wrapperStyles}>
        {!isFocused && EditorDataToReact(props.block.data.editorData, elementStyles)}
        {isFocused && (
          <MyEditor
            styles={elementStyles}
            toolbarButtons={['bold', 'italic', 'underlined', 'hyperlink', 'fonts']}
            onChange={(value) => {
              const newBlock = cloneDeep(props.block)
              newBlock.data.editorData = value
              props.updateTree(newBlock.path, newBlock)
            }}
            value={props.block.data.editorData}
            isFocused={true}
          />
        )}
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
          <div
            style={{ backgroundColor: '#1890ff', height: '2px', margin: '8px 14px 4px 14px' }}
          ></div>
          <div
            style={{ backgroundColor: '#1890ff', height: '2px', margin: '0px 14px 4px 14px' }}
          ></div>
          <div
            style={{ backgroundColor: '#1890ff', height: '2px', margin: '0px 14px 4px 14px' }}
          ></div>
          <div
            style={{ backgroundColor: '#1890ff', height: '2px', margin: '0px 14px 8px 14px' }}
          ></div>
        </div>
        Text
      </div>
    )
  }
}

export default TextBlockDefinition
