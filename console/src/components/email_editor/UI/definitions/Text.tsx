import React from 'react'
import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import { Form, Radio, Divider } from 'antd'
// import PaddingInputs from '../Widgets/PaddingInputs'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faAlignLeft, faAlignCenter, faAlignRight } from '@fortawesome/free-solid-svg-icons'
import ColorPickerInput from '../Widgets/ColorPicker'
import MyEditor, { EditorDataToReact } from '../Widgets/MyEditor'
import { cloneDeep } from 'lodash'
import ElementForms from '../Widgets/ElementForms'
// import { MenuOutlined } from '@ant-design/icons';

// import ReactQuill, { Quill } from 'react-quill';
// import 'react-quill/dist/quill.snow.css';

const TextBlockDefinition: BlockDefinitionInterface = {
  name: 'Text',
  kind: 'text',
  containsDraggables: false,
  isDraggable: true,
  draggableIntoGroup: 'column',
  isDeletable: true,
  defaultData: {
    align: 'left',
    // paddingControl: 'all', // all, separate
    // padding: '20px',
    width: '100%',
    // editorData: ' azeaze a ez',
    editorData: [
      // {
      //     type: 'h1',
      //     children: [{ text: 'Heading 1' }],
      // },
      // {
      //     type: 'h2',
      //     children: [{ text: 'Heading 2' }],
      // },
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

          <Divider className="margin-v-t" orientation="left" plain>
            Paragraphs global settings
          </Divider>
          {/* <div className="rmdeditor-margin-b-l rmdeditor-margin-t-l">Paragraphs Global Settings:</div> */}

          <ElementForms block={root} updateTree={props.updateTree} element="paragraph" />

          {/* <Form.Item label="Padding control" labelAlign="left" className="rmdeditor-form-item-align-right" labelCol={{ span: 10 }} wrapperCol={{ span: 14 }}>
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
                        <Radio.Button style={{ width: '50%', textAlign: 'center' }} value="all">All</Radio.Button>
                        <Radio.Button style={{ width: '50%', textAlign: 'center' }} value="separate">Separate</Radio.Button>
                    </Radio.Group>
                </Form.Item>

                {props.block.data.paddingControl === 'all' && <>
                    <Form.Item label="Paddings" labelAlign="left" className="rmdeditor-form-item-align-right" labelCol={{ span: 10 }} wrapperCol={{ span: 14 }}>
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
                            parser={(value: string) => {
                                // if (['▭'].indexOf(value)) {
                                //     value = value.substring(1)
                                // }
                                return parseInt(value.replace('px', ''))
                            }}
                            // formatter={value => '▭  ' + value + 'px'}
                            formatter={value => value + 'px'}
                        />
                    </Form.Item>
                </>}

                {props.block.data.paddingControl === 'separate' && <>
                    <Form.Item label="Paddings" labelAlign="left" className="rmdeditor-form-item-align-right" labelCol={{ span: 10 }} wrapperCol={{ span: 14 }}>
                        <PaddingInputs
                            styles={props.block.data}
                            onChange={(updatedStyles: any) => {
                                props.block.data = updatedStyles
                                props.updateTree(props.block.path, props.block)
                            }}
                        />
                    </Form.Item>
                </>} */}
        </div>

        {/* 
            <Collapse className="rmdeditor-padding-h-s" defaultActiveKey={['1']} ghost accordion>
                <Collapse.Panel header="Paragraph" key="1">
                    <ElementForms block={root} updateTree={props.updateTree} element="paragraph" />
                </Collapse.Panel>
                <Collapse.Panel header="Heading 1" key="2">
                    <ElementForms block={root} updateTree={props.updateTree} element="h1" />
                </Collapse.Panel>
                <Collapse.Panel header="Heading 2" key="3">
                    <ElementForms block={root} updateTree={props.updateTree} element="h2" />
                </Collapse.Panel>
            </Collapse> */}
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

    // console.log('selected: ' + props.selectedBlockId, props.selectedBlockId === props.block.id)
    // console.log('props.block.data.editorData', props.block.data.editorData)

    const isFocused = props.selectedBlockId === props.block.id

    return (
      <div style={wrapperStyles}>
        {!isFocused && EditorDataToReact(props.block.data.editorData, elementStyles)}
        {isFocused && (
          <MyEditor
            styles={elementStyles}
            toolbarButtons={['bold', 'italic', 'underlined', 'fonts']}
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
          {/* <MenuOutlined /> */}
          {/* <FontAwesomeIcon icon={faAlignJustify} /> */}
          {/* <FontAwesomeIcon icon={faBars} /> */}
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
