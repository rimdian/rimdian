import { Button, Space, Tooltip } from 'antd'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { Device, DataLog } from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'
import TableTag from 'components/common/partial_table_tag'
import Block from 'components/common/block'
import { Preview } from './block_user_timeline'
import {
  PartialDeviceBrowserIcon,
  PartialDeviceOSIcon,
  PartialDeviceTypeIcon
} from 'components/common/partial_device_icon'
import Property from 'components/common/partial_property'

const UserTimelineDevice = (props: {
  workspaceCtx: CurrentWorkspaceCtxValue
  devices: Device[]
  line: DataLog
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const device = props.devices.find((d) => d.id === props.line.item_id)
  //   console.log('data', data)
  if (!device) return Preview({ notFound: true, line: props.line })

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
        <Space className={CSS.opacity_50}>
          {PartialDeviceTypeIcon(device.device_type)}
          {device.os && PartialDeviceOSIcon(device.os)}
          {device.browser && PartialDeviceBrowserIcon(device.browser)}
        </Space>
        created
      </Space>
    </div>
  )

  if (!isOpen) return header

  const properties: any = []
  if (device.os) properties.push(<Property label="OS">{device.os}</Property>)
  if (device.browser)
    properties.push(
      <Property label="Browser">
        {device.browser} {device.browser_version_major}
      </Property>
    )
  if (device.resolution)
    properties.push(<Property label="Resolution">{device.resolution}</Property>)
  if (device.ad_blocker) properties.push(<Property label="Ad block">yes</Property>)
  if (device.in_webview) properties.push(<Property label="Mobile Webview">yes</Property>)
  if (device.language) properties.push(<Property label="Language">{device.language}</Property>)

  return (
    <div>
      {header}
      <Block classNames={[CSS.margin_l_xl, CSS.margin_t_m, CSS.padding_a_m]}>
        <table>
          <tbody>
            {/* split the properties in two columns */}
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
          </tbody>
        </table>
      </Block>
    </div>
  )
}

export default UserTimelineDevice
