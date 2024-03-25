import React, { ReactNode } from 'react'
import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import { Form, InputNumber } from 'antd'
import ColorPickerInput from '../Widgets/ColorPicker'
import {Fonts} from '../Widgets/ElementForms'
// import ElementForms from '../Widgets/ElementForms'

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
                backgroundColor: '#FFFFFF',
            },
            'h1': {
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
                fontFamily: Fonts[2].value,
            },
            'h2': {
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
                fontFamily: Fonts[2].value,
            },
            'h3': {
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
                fontFamily: Fonts[2].value,
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
                fontFamily: Fonts[2].value,
            },
        },
    },
    menuSettings: {},

    RenderSettings: (props: BlockRenderSettingsProps) => {

        // console.log('render settings', props)
        
        return <>
            <div className="cmeditor-padding-h-l">
                <Form.Item label="Email width" labelAlign="left" className="cmeditor-form-item-align-right" labelCol={{ span: 12 }} wrapperCol={{ span: 12 }}>
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
                        formatter={value => value + 'px'}
                        parser={value => parseInt((value || defaultWidth).replace('px', ''))}
                    />
                </Form.Item>

                <Form.Item label="Background color" labelAlign="left" className="cmeditor-form-item-align-right" labelCol={{ span: 12 }} wrapperCol={{ span: 12 }}>
                    <ColorPickerInput size="small" value={props.block.data.styles.body.backgroundColor} onChange={(newColor) => {
                        props.block.data.styles.body.backgroundColor = newColor
                        props.updateTree(props.block.path, props.block)
                    }} />
                </Form.Item>
            </div>

            {/* <Collapse className="cmeditor-padding-h-s" defaultActiveKey={['1']} ghost accordion>
                <Collapse.Panel header="Paragraph" key="1">
                    <ElementForms block={props.block} updateTree={props.updateTree} element="paragraph" />
                </Collapse.Panel>
                <Collapse.Panel header="Heading 1" key="2">
                    <ElementForms block={props.block} updateTree={props.updateTree} element="h1" />
                </Collapse.Panel>
                <Collapse.Panel header="Heading 2" key="3">
                    <ElementForms block={props.block} updateTree={props.updateTree} element="h2" />
                </Collapse.Panel>
            </Collapse> */}
        </>
    },

    renderEditor: (props: BlockEditorRendererProps, content: ReactNode) => {
        const styles = {
            margin: '0 auto',
            width: props.block.data.styles.body.width,
        }

        if (parseInt(props.block.data.styles.body.width || 0) > props.deviceWidth) {
            styles.width = props.deviceWidth + 'px'
        }
        return <div style={{
            paddingTop: '56px',
            minHeight: '100vh',
            backgroundColor: props.block.data.styles.body.backgroundColor,
        }}>
            <div style={styles}>{content}</div>
        </div>
    },

    // transformer: (block: BlockInterface) => {
    //     return <div>TODO transformer for {block.kind}</div>
    // },

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

export default RootBlockDefinition