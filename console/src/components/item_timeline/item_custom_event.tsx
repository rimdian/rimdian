import { Button, Space, Tooltip } from 'antd'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { DataLog, CustomEvent } from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'
import TableTag from 'components/common/partial_table_tag'
import Block from 'components/common/block'
import { AppColumns, Preview } from './block_user_timeline'
import Property from 'components/common/partial_property'

const UserTimelineCustomEvent = (props: {
  workspaceCtx: CurrentWorkspaceCtxValue
  line: DataLog
  customEvents: CustomEvent[]
  isLoading: boolean
}) => {
  // console.log('UserTimelineCustomEvent', props)
  const [isOpen, setIsOpen] = useState(false)

  if (props.isLoading) return Preview({ notFound: false, line: props.line })

  const customEvent = props.customEvents.find((p) => p.id === props.line.item_id)
  if (!customEvent) return Preview({ notFound: true, line: props.line })

  const header = (
    <div onClick={() => setIsOpen(!isOpen)} style={{ cursor: 'pointer' }}>
      <span className={CSS.pull_right}>
        {isOpen ? (
          <Button size="small" type="link" onClick={() => setIsOpen(false)}>
            hide -
          </Button>
        ) : (
          <Button size="small" type="link" onClick={() => setIsOpen(true)}>
            details +
          </Button>
        )}
      </span>

      <Space size="large">
        <Tooltip
          title={
            <>
              <p>Data log ID: {props.line.id}</p>
              {props.line.merged_from_user_external_id && (
                <p>Merged from user: {props.line.merged_from_user_external_id}</p>
              )}
              Ext. ID: {props.line.item_external_id}
            </>
          }
        >
          <span>
            <TableTag table={props.line.kind} />
          </span>
        </Tooltip>
        <b>{customEvent.label}</b>
        {customEvent.string_value && customEvent.string_value}
        {customEvent.number_value && customEvent.number_value}
        {customEvent.boolean_value !== undefined && JSON.stringify(customEvent.boolean_value)}
      </Space>
    </div>
  )

  if (!isOpen) return header

  const properties: any = []

  if (customEvent.string_value)
    properties.push(<Property label="String value">{customEvent.string_value}</Property>)
  if (customEvent.number_value !== undefined)
    properties.push(<Property label="Number value">{customEvent.number_value}</Property>)
  if (customEvent.boolean_value !== undefined)
    properties.push(
      <Property label="Boolean value">{JSON.stringify(customEvent.boolean_value)}</Property>
    )

  if (customEvent.non_interactive)
    properties.push(<Property label="Non interactive">{customEvent.non_interactive}</Property>)
  return (
    <div>
      {header}
      <Block classNames={[CSS.margin_l_xl, CSS.margin_t_m, CSS.padding_a_m]}>
        <table>
          <tbody>
            <tr>
              <td colSpan={2}>
                <Property label="Label">
                  <>{customEvent.label}</>
                </Property>
              </td>
            </tr>
            {/* split properties in two columns */}
            {properties.map((property: any, index: number) => {
              if (index % 2 === 0) {
                return (
                  <tr key={index}>
                    <td>{property}</td>
                    <td>{properties[index + 1]}</td>
                  </tr>
                )
              }
              return null
            })}
            <AppColumns
              kind="pageview"
              item={customEvent}
              apps={props.workspaceCtx.workspace.apps}
            />
          </tbody>
        </table>
      </Block>
    </div>
  )
}

export default UserTimelineCustomEvent
