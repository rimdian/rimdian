import { BlockDefinitionInterface } from '../../Block'
import SectionBlockDefinition from './Section'
import Column from './Column'
import cloneDeep from 'lodash/cloneDeep'
import React from 'react'

const OneColumnBlockDefinition: BlockDefinitionInterface = cloneDeep(SectionBlockDefinition)

// OneColumnBlockDefinition.name = 'Section'
OneColumnBlockDefinition.kind = 'oneColumn'
OneColumnBlockDefinition.children = [cloneDeep(Column)]

OneColumnBlockDefinition.renderMenu = () => (
  <div className="rmdeditor-ui-block">
    <div className="rmdeditor-ui-block-col"></div>
  </div>
)

export default OneColumnBlockDefinition
