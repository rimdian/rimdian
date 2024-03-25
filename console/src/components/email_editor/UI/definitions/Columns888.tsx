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

// Column1.defaultData.paddingControl = 'separate'
// Column1.defaultData.styles.paddingRight = '15px'
// Column2.defaultData.paddingControl = 'separate'
// Column2.defaultData.styles.paddingLeft = '15px'
// Column2.defaultData.styles.paddingRight = '15px'
// Column3.defaultData.paddingControl = 'separate'
// Column3.defaultData.styles.paddingLeft = '15px'

const Columns888BlockDefinition: BlockDefinitionInterface = cloneDeep(SectionBlockDefinition)

Columns888BlockDefinition.name = 'Section 1-1-1'
Columns888BlockDefinition.kind = 'columns888'
Columns888BlockDefinition.columns = [8, 8, 8]
Columns888BlockDefinition.children = [Column1, Column2, Column3]
Columns888BlockDefinition.renderMenu = () => (
  <div className="rmdeditor-ui-block">
    <Row gutter={12}>
      <Col span={8}>
        <div className="rmdeditor-ui-block-col"></div>
      </Col>
      <Col span={8}>
        <div className="rmdeditor-ui-block-col"></div>
      </Col>
      <Col span={8}>
        <div className="rmdeditor-ui-block-col"></div>
      </Col>
    </Row>
  </div>
)

export default Columns888BlockDefinition
