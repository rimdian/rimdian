import { BlockDefinitionInterface } from '../../Block'
import SectionBlockDefinition from './Section'
import Column from './Column'
import cloneDeep from 'lodash/cloneDeep'
// import { BlockEditorRendererProps, RenderChildren } from '../BlockEditorRenderer'
import { Row, Col } from 'antd'
import React from 'react'

const Column1 = cloneDeep(Column)
const Column2 = cloneDeep(Column)

Column1.defaultData.paddingControl = 'separate'
Column1.defaultData.styles.paddingRight = '15px'
Column2.defaultData.paddingControl = 'separate'
Column2.defaultData.styles.paddingLeft = '15px'

const Columns168BlockDefinition: BlockDefinitionInterface = cloneDeep(SectionBlockDefinition)

Columns168BlockDefinition.name = 'Section 2-1'
Columns168BlockDefinition.kind = 'columns168'
Columns168BlockDefinition.columns = [16, 8]
Columns168BlockDefinition.children = [Column1, Column2]
Columns168BlockDefinition.renderMenu = () => <div className="cmeditor-ui-block">
    <Row gutter={12}>
        <Col span={16}><div className="cmeditor-ui-block-col"></div></Col>
        <Col span={8}><div className="cmeditor-ui-block-col"></div></Col>
    </Row>
</div>

export default Columns168BlockDefinition