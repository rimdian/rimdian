import { ReactNode } from 'react'
import { BlockDefinitionMap, BlockDefinitionInterface, BlockInterface } from '../Block'
import { Collapse, Tooltip } from 'antd'
import { truncate } from 'lodash'

export interface BlocksProps {
  blockDefinitions: BlockDefinitionMap
  savedBlocks: BlockDefinitionInterface[]
  renderBlockForMenu: (blockDefinition: BlockDefinitionInterface) => ReactNode
  renderSavedBlockForMenu: (block: BlockInterface, renderMenu: ReactNode) => ReactNode
}

export const Blocks = (props: BlocksProps) => {
  return (
    <>
      <Collapse defaultActiveKey={['1', '3']} ghost>
        <Collapse.Panel
          header={
            <span className="rmdeditor-ui-menu-title" style={{ padding: 0, margin: 0 }}>
              Content blocks
            </span>
          }
          key="1"
          style={{ paddingLeft: 0, paddingRight: 0, paddingTop: 12, marginBottom: 0 }}
        >
          {props.renderBlockForMenu(props.blockDefinitions['image'])}
          {props.renderBlockForMenu(props.blockDefinitions['button'])}

          {props.renderBlockForMenu(props.blockDefinitions['heading'])}
          {props.renderBlockForMenu(props.blockDefinitions['text'])}

          {props.renderBlockForMenu(props.blockDefinitions['divider'])}
          {props.renderBlockForMenu(props.blockDefinitions['openTracking'])}
        </Collapse.Panel>
        <Collapse.Panel
          header={
            <span className="rmdeditor-ui-menu-title" style={{ padding: 0, margin: 0 }}>
              Saved blocks ({props.savedBlocks.length})
            </span>
          }
          key="2"
          style={{ paddingLeft: 0, paddingRight: 0, paddingTop: 12, marginBottom: 0 }}
        >
          {props.savedBlocks.map((b: any, i: number) => {
            return (
              <div key={i}>
                {props.renderSavedBlockForMenu(
                  JSON.parse(b.block),
                  <Tooltip title={b.name}>
                    <div className="rmdeditor-ui-saved-block">
                      {truncate(b.name, { length: 20 })}
                    </div>
                  </Tooltip>
                )}
              </div>
            )
          })}
        </Collapse.Panel>
        <Collapse.Panel
          header={
            <span className="rmdeditor-ui-menu-title" style={{ padding: 0, margin: 0 }}>
              Layout blocks
            </span>
          }
          key="3"
          style={{ paddingLeft: 0, paddingRight: 0, paddingTop: 12, marginBottom: 0 }}
        >
          {props.renderBlockForMenu(props.blockDefinitions['oneColumn'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns1212'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns888'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns6666'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns420'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns816'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns168'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns204'])}
        </Collapse.Panel>
      </Collapse>
    </>
  )
}
