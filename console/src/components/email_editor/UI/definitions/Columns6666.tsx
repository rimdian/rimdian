import { BlockDefinitionInterface } from '../../Block'
import SectionBlockDefinition from './Section'
import Column from './Column'
import cloneDeep from 'lodash/cloneDeep'
// import { BlockEditorRendererProps, RenderChildren } from '../BlockEditorRenderer'
import { Row, Col } from 'antd'
import React from 'react'

const Column1 = cloneDeep(Column)
const Column2 = cloneDeep(Column)
const Column3 = cloneDeep(Column)
const Column4 = cloneDeep(Column)

// Column1.defaultData.paddingControl = 'separate'
// Column1.defaultData.styles.paddingRight = '15px'
// Column2.defaultData.paddingControl = 'separate'
// Column2.defaultData.styles.paddingLeft = '15px'
// Column2.defaultData.styles.paddingRight = '15px'
// Column3.defaultData.paddingControl = 'separate'
// Column3.defaultData.styles.paddingLeft = '15px'
// Column4.defaultData.paddingControl = 'separate'
// Column4.defaultData.styles.paddingLeft = '15px'

const Columns6666BlockDefinition: BlockDefinitionInterface = cloneDeep(SectionBlockDefinition)

Columns6666BlockDefinition.name = 'Section 1-1-1-1'
Columns6666BlockDefinition.kind = 'columns6666'
Columns6666BlockDefinition.columns = [6, 6, 6, 6]
Columns6666BlockDefinition.children = [Column1, Column2, Column3, Column4]
Columns6666BlockDefinition.renderMenu = () => <div className="cmeditor-ui-block">
    <Row gutter={12}>
        <Col span={6}><div className="cmeditor-ui-block-col"></div></Col>
        <Col span={6}><div className="cmeditor-ui-block-col"></div></Col>
        <Col span={6}><div className="cmeditor-ui-block-col"></div></Col>
        <Col span={6}><div className="cmeditor-ui-block-col"></div></Col>
    </Row>
</div>

export default Columns6666BlockDefinition