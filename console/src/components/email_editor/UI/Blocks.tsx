import React, { ReactNode } from 'react'
import { BlockDefinitionMap, BlockDefinitionInterface, BlockInterface } from '../Block'
import { Row, Col, Collapse, Tooltip } from 'antd'
import { truncate } from 'lodash'

export interface BlocksProps {
    blockDefinitions: BlockDefinitionMap
    savedBlocks: BlockDefinitionInterface[]
    renderBlockForMenu: (blockDefinition: BlockDefinitionInterface) => ReactNode
    renderSavedBlockForMenu: (block: BlockInterface, renderMenu: ReactNode) => ReactNode
}

export const Blocks = (props: BlocksProps) => {

    return <>

        <Collapse defaultActiveKey={['1', '3']} ghost>
            <Collapse.Panel header={<span className="cmeditor-ui-menu-title" style={{ padding: 0, margin: 0 }}>Content blocks</span>} key="1" style={{ paddingLeft: 0, paddingRight: 0, paddingTop: 12, marginBottom: 0 }}>
                <Row gutter={12}>
                    <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["image"])}</Col>
                    <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["button"])}</Col>
                </Row>
                <Row gutter={12}>
                    <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["heading"])}</Col>
                    <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["text"])}</Col>
                </Row>
                <Row gutter={12}>
                    <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["divider"])}</Col>
                    {/* <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["text"])}</Col> */}
                </Row>
            </Collapse.Panel>
            <Collapse.Panel header={<span className="cmeditor-ui-menu-title" style={{ padding: 0, margin: 0 }}>Saved blocks ({props.savedBlocks.length})</span>} key="2" style={{ paddingLeft: 0, paddingRight: 0, paddingTop: 12, marginBottom: 0 }}>
                {props.savedBlocks.map((b: any, i: number) => {
                    return <div key={i}>
                        {props.renderSavedBlockForMenu(JSON.parse(b.block), <Tooltip title={b.name}>
                            <div className="cmeditor-ui-saved-block">{truncate(b.name, { length: 20 })}</div>
                        </Tooltip>)}
                    </div>
                })}
            </Collapse.Panel>
            <Collapse.Panel header={<span className="cmeditor-ui-menu-title" style={{ padding: 0, margin: 0 }}>Layout blocks</span>} key="3" style={{ paddingLeft: 0, paddingRight: 0, paddingTop: 12, marginBottom: 0 }}>
                {props.renderBlockForMenu(props.blockDefinitions["oneColumn"])}
                {props.renderBlockForMenu(props.blockDefinitions["columns420"])}
                {props.renderBlockForMenu(props.blockDefinitions["columns816"])}
                {props.renderBlockForMenu(props.blockDefinitions["columns1212"])}
                {props.renderBlockForMenu(props.blockDefinitions["columns168"])}
                {props.renderBlockForMenu(props.blockDefinitions["columns204"])}
                {props.renderBlockForMenu(props.blockDefinitions["columns888"])}
                {props.renderBlockForMenu(props.blockDefinitions["columns6666"])}
            </Collapse.Panel>
        </Collapse>,

        {/* <div className="cmeditor-ui-menu-title">Content blocks</div>
        <div className="cmeditor-padding-h-l">
            <Row gutter={12}>
                <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["image"])}</Col>
                <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["button"])}</Col>
            </Row>
            <Row gutter={12}>
                <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["heading"])}</Col>
                <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["text"])}</Col>
            </Row>
            <Row gutter={12}>
                <Col span={12}>{props.renderBlockForMenu(props.blockDefinitions["divider"])}</Col>
            </Row>
        </div>

        <div className="cmeditor-ui-menu-title">Saved blocks </div>

        <div className="cmeditor-ui-menu-title">Layout blocks</div>
        <div className="cmeditor-padding-h-l">
            {props.renderBlockForMenu(props.blockDefinitions["oneColumn"])}
            {props.renderBlockForMenu(props.blockDefinitions["columns420"])}
            {props.renderBlockForMenu(props.blockDefinitions["columns816"])}
            {props.renderBlockForMenu(props.blockDefinitions["columns1212"])}
            {props.renderBlockForMenu(props.blockDefinitions["columns168"])}
            {props.renderBlockForMenu(props.blockDefinitions["columns204"])}
            {props.renderBlockForMenu(props.blockDefinitions["columns888"])}
            {props.renderBlockForMenu(props.blockDefinitions["columns6666"])}
        </div> */}

    </>
}