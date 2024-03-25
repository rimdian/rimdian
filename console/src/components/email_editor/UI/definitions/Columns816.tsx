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

const Columns816BlockDefinition: BlockDefinitionInterface = cloneDeep(SectionBlockDefinition)

Columns816BlockDefinition.name = 'Section 1-2'
Columns816BlockDefinition.kind = 'columns816'
Columns816BlockDefinition.columns = [8, 16]
Columns816BlockDefinition.children = [Column1, Column2]
Columns816BlockDefinition.renderMenu = () => (
  <div className="rmdeditor-ui-block">
    <Row gutter={12}>
      <Col span={8}>
        <div className="rmdeditor-ui-block-col"></div>
      </Col>
      <Col span={16}>
        <div className="rmdeditor-ui-block-col"></div>
      </Col>
    </Row>
  </div>
)

export default Columns816BlockDefinition
