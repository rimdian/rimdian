import { Row, Col, Button, Popover, Spin, Table, Tag } from 'antd'
import { Channel, ChannelGroup, Origin, VoucherCode } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import AlphaSort from 'utils/alpha_sort'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPlus, faTag } from '@fortawesome/free-solid-svg-icons'
import DeleteChannelButton from './button_delete_channel'
import UpsertChannelButton from './button_upsert_channel'
import { faPenToSquare } from '@fortawesome/free-regular-svg-icons'
import DeleteChannelGroupButton from './button_delete_channel_group'
import UpsertChannelGroupButton from './button_upsert_channel_group'
import CSS from 'utils/css'
import { blockCss } from 'components/common/block'

export const reservedGroups = ['not-mapped', 'direct']

const TabTrafficMapping = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  return (
    <>
      <Row gutter={24}>
        <Col span={5}>
          <Table
            dataSource={workspaceCtx.workspace.channel_groups.sort(AlphaSort('created_at'))}
            rowKey="id"
            className={blockCss.self}
            pagination={false}
            size="middle"
            columns={[
              {
                key: 'name',
                title: (
                  <div>
                    Groups
                    <span className={CSS.pull_right}>
                      <UpsertChannelGroupButton
                        workspaceId={workspaceCtx.workspace.id}
                        channelGroups={workspaceCtx.workspace.channel_groups}
                        btnSize="small"
                        btnType="default"
                        btnContent={
                          <>
                            <FontAwesomeIcon icon={faPlus} />
                            &nbsp; New group
                          </>
                        }
                        apiPOST={workspaceCtx.apiPOST}
                        onComplete={() => {
                          workspaceCtx.refreshWorkspace()
                        }}
                      />
                    </span>
                  </div>
                ),
                render: (item: ChannelGroup) => {
                  return (
                    <>
                      <Tag color={item.color}>{item.name}</Tag>
                      <span className={CSS.pull_right}>
                        {!reservedGroups.includes(item.id) && (
                          <Button.Group>
                            <DeleteChannelGroupButton
                              channelGroupId={item.id}
                              workspaceId={workspaceCtx.workspace.id}
                              apiPOST={workspaceCtx.apiPOST}
                              onComplete={() => {
                                workspaceCtx.refreshWorkspace()
                              }}
                              btnSize="small"
                              btnType="text"
                            />

                            <UpsertChannelGroupButton
                              workspaceId={workspaceCtx.workspace.id}
                              channelGroup={item}
                              channelGroups={workspaceCtx.workspace.channel_groups}
                              btnSize="small"
                              btnType="text"
                              btnContent={<FontAwesomeIcon icon={faPenToSquare} />}
                              apiPOST={workspaceCtx.apiPOST}
                              onComplete={() => {
                                workspaceCtx.refreshWorkspace()
                              }}
                            />
                          </Button.Group>
                        )}
                      </span>
                    </>
                  )
                }
              }
              // {
              //     key: 'actions',
              //     className: 'actions',
              //     title: <div className={GlobalCSS.text_right}>
              //         <UpsertChannelGroupButton
              //             workspaceId={workspaceCtx.workspace.id}
              //             channelGroups={workspaceCtx.workspace.channelGroups}
              //             btnSize="small"
              //             btnType='ghost'
              //             btnContent={<><FontAwesomeIcon icon={faPlus} />&nbsp; New group</>}
              //             apiPOST={workspaceCtx.apiPOST}
              //             onComplete={() => {
              //                 workspaceCtx.refreshWorkspace()
              //             }}
              //         />
              //     </div>,
              //     render: (item: ChannelGroup) => <div className={GlobalCSS.text_right}>
              //         {!reservedGroups.includes(item.id) && <Button.Group>
              //             <DeleteChannelGroupButton
              //                 channelGroupId={item.id}
              //                 workspaceId={workspaceCtx.workspace.id}
              //                 apiPOST={workspaceCtx.apiPOST}
              //                 onComplete={() => {
              //                     workspaceCtx.refreshWorkspace()
              //                 }}
              //                 btnSize="small"
              //             />

              //             <UpsertChannelGroupButton
              //                 workspaceId={workspaceCtx.workspace.id}
              //                 channelGroup={item}
              //                 channelGroups={workspaceCtx.workspace.channelGroups}
              //                 btnSize="small"
              //                 btnType='ghost'
              //                 btnContent={<FontAwesomeIcon icon={faPenToSquare} />}
              //                 apiPOST={workspaceCtx.apiPOST}
              //                 onComplete={() => {
              //                     workspaceCtx.refreshWorkspace()
              //                 }}
              //             />
              //         </Button.Group>}
              //     </div>
              // }
            ]}
          />
        </Col>
        <Col span={19}>
          <Table
            className={blockCss.self}
            pagination={false}
            dataSource={workspaceCtx.workspace.channels}
            size="middle"
            rowKey="id"
            columns={[
              {
                key: 'name',
                title: 'Channel',
                sorter: AlphaSort('name'),
                // sortDirections: ['ascend'],
                render: (row: Channel) => row.name
              },
              {
                key: 'group',
                title: 'Group',
                sorter: AlphaSort('groupId'),
                // sortDirections: ['ascend'],
                render: (row: Channel) => {
                  const group = workspaceCtx.workspace.channel_groups.find(
                    (x) => x.id === row.group_id
                  )
                  if (!group) return <Tag>{row.group_id}</Tag>
                  return <Tag color={group.color}>{group.name}</Tag>
                }
              },
              {
                key: 'origins',
                title: 'Origins',
                render: (row: Channel) => {
                  return (
                    <>
                      {row.origins.map((x: Origin) => (
                        <div key={row.id + x.id}>{x.id}</div>
                      ))}
                    </>
                  )
                }
              },
              {
                key: 'voucherCodes',
                title: 'Voucher codes',
                render: (row: Channel) => (
                  <>
                    {row.voucher_codes.map((x: VoucherCode) => (
                      <div key={x.code}>
                        <VoucherCodePopover voucherCode={x} />
                      </div>
                    ))}
                  </>
                )
              },
              {
                key: 'actions',
                title: (
                  <div className={CSS.text_right}>
                    {workspaceCtx.isRefreshingWorkspace && (
                      <Spin className={CSS.margin_r_s} size="small" />
                    )}
                    <UpsertChannelButton
                      workspaceId={workspaceCtx.workspace.id}
                      channels={workspaceCtx.workspace.channels}
                      channelGroups={workspaceCtx.workspace.channel_groups}
                      btnSize="small"
                      btnType="primary"
                      btnContent={
                        <>
                          <FontAwesomeIcon icon={faPlus} />
                          &nbsp; New channel
                        </>
                      }
                      apiPOST={workspaceCtx.apiPOST}
                      onComplete={() => {
                        workspaceCtx.refreshWorkspace()
                      }}
                    />
                  </div>
                ),
                render: (row: Channel) => (
                  <div className={CSS.text_right}>
                    <Button.Group>
                      <DeleteChannelButton
                        channelId={row.id}
                        workspaceId={workspaceCtx.workspace.id}
                        apiPOST={workspaceCtx.apiPOST}
                        onComplete={() => {
                          workspaceCtx.refreshWorkspace()
                        }}
                        btnSize="small"
                        btnType="text"
                      />

                      <UpsertChannelButton
                        workspaceId={workspaceCtx.workspace.id}
                        channel={row}
                        channels={workspaceCtx.workspace.channels}
                        channelGroups={workspaceCtx.workspace.channel_groups}
                        btnSize="small"
                        btnType="text"
                        btnContent={<FontAwesomeIcon icon={faPenToSquare} />}
                        apiPOST={workspaceCtx.apiPOST}
                        onComplete={() => {
                          workspaceCtx.refreshWorkspace()
                        }}
                      />
                    </Button.Group>
                  </div>
                )
              }
            ]}
          />
        </Col>
      </Row>
    </>
  )
}

export default TabTrafficMapping

type VoucherCodePopoverProps = {
  voucherCode: VoucherCode
}
export const VoucherCodePopover = (props: VoucherCodePopoverProps) => {
  const content = (
    <div>
      <p>
        <Tag>
          <FontAwesomeIcon icon={faTag} />
          &nbsp; {props.voucherCode.code}
        </Tag>
      </p>
      {props.voucherCode.description && <p>{props.voucherCode.description}</p>}

      <p>
        <b>&rarr; Attribute to:</b> {props.voucherCode.origin_id}
        {props.voucherCode.set_utm_campaign && (
          <div>
            <b>Replace utm_campaign with:</b> {props.voucherCode.set_utm_campaign}
          </div>
        )}
        {props.voucherCode.set_utm_content && (
          <div>
            <b>Replace utm_content with:</b> {props.voucherCode.set_utm_content}
          </div>
        )}
      </p>
    </div>
  )
  return (
    <Popover content={content}>
      <Tag>
        <FontAwesomeIcon icon={faTag} />
        &nbsp; {props.voucherCode.code}
      </Tag>
      {/* {props.voucherCode.set_utm_campaign && <><FontAwesomeIcon icon={faBullhorn} />&nbsp;</>}
        {props.voucherCode.set_utm_content && <><FontAwesomeIcon icon={faFileImage} />&nbsp;</>}

        &rarr; {props.voucherCode.originId} */}
    </Popover>
  )
}
