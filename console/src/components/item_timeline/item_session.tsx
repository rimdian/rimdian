import { Badge, Button, Divider, Space, Tag, Tooltip } from 'antd'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { DataLog, Session, Device } from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'
import dayjs from 'dayjs'
import FormatDuration from 'utils/format_duration'
import FormatCurrency from 'utils/format_currency'
import FormatConversionRole from 'utils/format_conversion_role'
import FormatPercent from 'utils/format_percent'
import { PartialDevice } from 'components/common/partial_device_icon'
import TableTag from 'components/common/partial_table_tag'
import Property from 'components/common/partial_property'
import Block from 'components/common/block'
import { AppColumns, Preview } from './block_user_timeline'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCircleXmark } from '@fortawesome/free-regular-svg-icons'

const UserTimelineSession = (props: {
  workspaceCtx: CurrentWorkspaceCtxValue
  devices: Device[]
  sessions: Session[]
  isLoading: boolean
  line: DataLog
}) => {
  const [isOpen, setIsOpen] = useState(false)
  if (props.isLoading) {
    return Preview({ notFound: false, line: props.line })
  }

  const session = props.sessions.find((s) => s.id === props.line.item_id)
  if (!session) return Preview({ notFound: true, line: props.line })

  const device = props.devices.find((d) => d.id === session.device_id)
  const domain = props.workspaceCtx.workspace.domains.find((d) => d.id === session.domain_id)

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
        {device && <PartialDevice device={device} />}
        {domain && domain.type !== 'web' && (
          <>
            <Tooltip title={domain.name}>{domain.type}</Tooltip>
          </>
        )}
        {session.duration && FormatDuration(session.duration)}
        {session.bounced === 1 && (
          <Tooltip title="Bounced">
            <FontAwesomeIcon className={CSS.text_red} icon={faCircleXmark} />
          </Tooltip>
        )}
        <b>
          <Tooltip title="utm_source">{session.utm_source}</Tooltip>
          {' / '}
          <Tooltip title="utm_medium">{session.utm_medium}</Tooltip>
          {session.utm_campaign && (
            <Tooltip title="utm_campaign"> / {session.utm_campaign}</Tooltip>
          )}
          {session.utm_term && <Tooltip title="utm_term"> / {session.utm_term}</Tooltip>}
          {session.utm_content && <Tooltip title="utm_content"> / {session.utm_content}</Tooltip>}
        </b>
      </Space>
    </div>
  )

  if (!isOpen) return header

  const channel = props.workspaceCtx.workspace.channels.find((c) => c.id === session.channel_id)
  const channelName = channel ? channel.name : session.channel_id
  const channelGroup = props.workspaceCtx.workspace.channel_groups.find(
    (cg: any) => cg.id === session.channel_group_id
  )
  const channelGroupName = channelGroup ? channelGroup.name : session.channel_group_id

  return (
    <div>
      {header}
      <Block classNames={[CSS.margin_l_xl, CSS.margin_t_m]}>
        <div className={CSS.padding_a_m}>
          {(session.bounced === 1 || session.bounced === true) && (
            <Badge status="error" text="Bounced" className={CSS.margin_b_s} />
          )}
          <Property label="Channel">
            <>
              {channelName}
              {channelGroup && (
                <Tag color={channelGroup ? channelGroup.color : undefined}>{channelGroupName}</Tag>
              )}
            </>
          </Property>
          <table>
            <tbody>
              <tr>
                <td>
                  <Property label="utm_source">
                    <>{session.utm_source}</>
                  </Property>
                </td>
                <td>
                  <Property label="utm_medium">
                    <>{session.utm_medium}</>
                  </Property>
                </td>
              </tr>
              <tr>
                <td>
                  {session.utm_campaign && (
                    <Property label="utm_campaign">
                      <>{session.utm_campaign}</>
                    </Property>
                  )}
                </td>
                <td>
                  {session.utm_term && (
                    <Property label="utm_term">
                      <>{session.utm_term}</>
                    </Property>
                  )}
                  {session.utm_content && (
                    <Property label="utm_content">
                      <>{session.utm_content}</>
                    </Property>
                  )}
                </td>
              </tr>
              <tr>
                <td colSpan={2}>
                  {session.referrer && (
                    <Property label="Referrer">
                      <a href={session.referrer} rel="noreferrer" target="_blank">
                        {session.referrer}
                      </a>
                    </Property>
                  )}
                </td>
              </tr>
              <tr>
                <td colSpan={2}>
                  {session.landing_page && (
                    <Property label="Landing">
                      <a href={session.landing_page} rel="noreferrer" target="_blank">
                        {session.landing_page}
                      </a>
                    </Property>
                  )}
                </td>
              </tr>
              <tr>
                <td>
                  <Property label="Session duration">
                    <>{session.duration ? FormatDuration(session.duration) : 'n/a'}</>
                  </Property>
                </td>
                <td>
                  {/* {(session.bounced === 1 || session.bounced === true) &&
                    Property('Bounced', <>yes</>)} */}
                </td>
              </tr>
              <AppColumns kind="session" item={session} apps={props.workspaceCtx.workspace.apps} />
            </tbody>
          </table>
        </div>
        {session.conversion_type === 'order' && (
          <>
            <Divider style={{ marginTop: 0 }} orientation="left" plain>
              Attribution
            </Divider>
            <div className={CSS.padding_h_l + ' ' + CSS.padding_b_m}>
              <table>
                <tbody>
                  <tr>
                    <td>
                      <Property label="Order ext. ID">
                        <>{session.conversion_external_id}</>
                      </Property>
                    </td>
                    <td>
                      <Property label="Amount">
                        {FormatCurrency(
                          session.conversion_amount || 0,
                          props.workspaceCtx.workspace.currency
                        )}
                      </Property>
                    </td>
                  </tr>
                  <tr>
                    <td>
                      <Property label="From session to conversion">
                        {FormatDuration(
                          dayjs(session.conversion_at).diff(dayjs(session.created_at), 'second')
                        )}
                      </Property>
                    </td>
                    <td>
                      <Property label="Role">
                        {FormatConversionRole(session.role as number)}
                      </Property>
                    </td>
                  </tr>
                  <tr>
                    <td>
                      <Property label="Linear amount attributed">
                        {FormatCurrency(
                          session.linear_amount_attributed || 0,
                          props.workspaceCtx.workspace.currency
                        )}
                      </Property>
                    </td>
                    <td>
                      <Property label="Linear % attributed">
                        {FormatPercent(
                          session.linear_percentage_attributed
                            ? session.linear_percentage_attributed / 10000
                            : 0
                        )}
                      </Property>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </>
        )}
      </Block>
    </div>
  )
}

export default UserTimelineSession
