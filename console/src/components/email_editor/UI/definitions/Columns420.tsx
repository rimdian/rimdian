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

const Columns420BlockDefinition: BlockDefinitionInterface = cloneDeep(SectionBlockDefinition)

Columns420BlockDefinition.name = 'Section 1-5'
Columns420BlockDefinition.kind = 'columns420'
Columns420BlockDefinition.columns = [4, 20]
Columns420BlockDefinition.children = [Column1, Column2]
Columns420BlockDefinition.renderMenu = () => (
  <div className="rmdeditor-ui-block">
    <Row gutter={12}>
      <Col span={4}>
        <div className="rmdeditor-ui-block-col"></div>
      </Col>
      <Col span={20}>
        <div className="rmdeditor-ui-block-col"></div>
      </Col>
    </Row>
  </div>
)

export default Columns420BlockDefinition
