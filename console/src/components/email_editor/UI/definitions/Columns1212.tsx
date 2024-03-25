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

const Columns1212BlockDefinition: BlockDefinitionInterface = cloneDeep(SectionBlockDefinition)

Columns1212BlockDefinition.name = 'Section 1-1'
Columns1212BlockDefinition.kind = 'columns1212'
Columns1212BlockDefinition.columns = [12,12]
Columns1212BlockDefinition.children = [Column1, Column2]
Columns1212BlockDefinition.renderMenu = () => <div className="cmeditor-ui-block">
    <Row gutter={12}>
        <Col span={12}><div className="cmeditor-ui-block-col"></div></Col>
        <Col span={12}><div className="cmeditor-ui-block-col"></div></Col>
    </Row>
</div>

export default Columns1212BlockDefinition