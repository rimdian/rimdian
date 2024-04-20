import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import CSS from 'utils/css'
import { useQuery } from '@tanstack/react-query'
import { BroadcastCampaign } from 'interfaces'
import { Alert, Button, Popconfirm, Space, Table, Tag, Tooltip, message } from 'antd'
import ButtonUpsertCampaign from './button_upsert_broadcast'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPenToSquare } from '@fortawesome/free-regular-svg-icons'
import dayjs from 'dayjs'
import { useAccount } from 'components/login/context_account'
import { faPause, faPlay, faRefresh } from '@fortawesome/free-solid-svg-icons'
import { useState } from 'react'
import ButtonPreviewMessageTemplate from 'components/assets/message_template/button_preview_template'
import numbro from 'numbro'
import { EmailProviderSettings } from 'components/workspace/block_messaging_settings'

const RouteBroadcasts = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const accountCtx = useAccount()
  const [isLoadingAction, setIsLoadingAction] = useState(false)

  const { isLoading, data, refetch, isFetching } = useQuery<BroadcastCampaign[]>(
    ['broadcast_campaigns', workspaceCtx.workspace.id],
    (): Promise<BroadcastCampaign[]> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET('/broadcastCampaign.list?workspace_id=' + workspaceCtx.workspace.id)
          .then((data: any) => {
            // console.log(data)
            resolve(data as BroadcastCampaign[])
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  const pauseCampaign = (id: string) => {
    const campaign = data?.find((c) => c.id === id)

    if (!campaign) return
    if (campaign.status !== 'scheduled' || !campaign.scheduled_at) return
    if (isLoadingAction) return
    setIsLoadingAction(true)

    campaign.scheduled_at = undefined
    campaign.status = 'draft'

    workspaceCtx
      .apiPOST('/broadcastCampaign.upsert', {
        workspace_id: workspaceCtx.workspace.id,
        ...campaign
      })
      .then(() => {
        refetch().then(() => {
          message.success('The campaign has been paused!')
          setIsLoadingAction(false)
        })
      })
      .finally(() => {
        setIsLoadingAction(false)
      })
  }

  const launchCampaign = (id: string) => {
    if (isLoadingAction) return
    setIsLoadingAction(true)

    workspaceCtx
      .apiPOST('/broadcastCampaign.launch', {
        workspace_id: workspaceCtx.workspace.id,
        id: id
      })
      .then(() => {
        refetch().then(() => {
          message.success('The campaign has been launched!')
          setIsLoadingAction(false)
        })
      })
      .finally(() => {
        setIsLoadingAction(false)
      })
  }

  return (
    <Layout
      currentOrganization={workspaceCtx.organization}
      currentWorkspace={workspaceCtx.workspace}
    >
      <div className={CSS.container}>
        <div className={CSS.top}>
          <h1>Broadcast campaigns</h1>
          <div className={CSS.topSeparator}></div>

          {!isLoading && data && data.length > 0 && (
            <ButtonUpsertCampaign btnProps={{ type: 'primary' }} onSuccess={() => refetch()}>
              New campaign
            </ButtonUpsertCampaign>
          )}
        </div>

        {!workspaceCtx.workspace.messaging_settings.marketing_email_provider && (
          <div className={CSS.margin_b_l}>
            <Alert
              message={
                <>
                  No marketing email provider configured.{' '}
                  <EmailProviderSettings
                    btnProps={{ size: 'small', type: 'primary' }}
                    kind="marketing_email_provider"
                    workspaceCtx={workspaceCtx}
                  >
                    <>Setup now</>
                  </EmailProviderSettings>
                </>
              }
              type="warning"
            />
          </div>
        )}

        <Table
          pagination={false}
          dataSource={data}
          loading={isLoading || isFetching}
          rowKey="id"
          locale={{
            emptyText: (
              <>
                <p>No campaign found</p>
                <ButtonUpsertCampaign btnProps={{ type: 'primary' }} onSuccess={() => refetch()}>
                  New campaign
                </ButtonUpsertCampaign>
              </>
            )
          }}
          columns={[
            {
              title: 'Campaign',
              key: 'id',
              render: (x: BroadcastCampaign) => (
                <div>
                  <Tooltip title={x.id}>{x.name}</Tooltip>
                </div>
              )
            },
            {
              title: 'Channel',
              key: 'channel',
              render: (x: BroadcastCampaign) => <div>{x.channel}</div>
            },
            {
              title: 'utm_source / medium / campaign',
              key: 'channel',
              render: (x: BroadcastCampaign) => (
                <div>
                  {x.utm_source} / {x.utm_medium} / {x.id}
                </div>
              )
            },
            {
              title: 'To',
              key: 'to',
              render: (x: BroadcastCampaign) => (
                <div>
                  {x.subscription_lists.map((list) => {
                    const subscriptionList = workspaceCtx.subscriptionLists.find(
                      (l) => l.id === list.id
                    )
                    return (
                      <p key={list.id}>
                        <Tag color={subscriptionList?.color}>{subscriptionList?.name}</Tag>(
                        {subscriptionList &&
                          numbro(subscriptionList.active_users).format({
                            totalLength: 3,
                            trimMantissa: true
                          })}
                        )
                      </p>
                    )
                  })}
                </div>
              )
            },
            {
              title: 'A/B templates',
              key: 'templates',
              render: (x: BroadcastCampaign) => {
                const templates = x.message_templates.map((msgTemplate) => {
                  return {
                    id: msgTemplate.id,
                    percentage: msgTemplate.percentage
                  }
                })

                return (
                  <div>
                    {x.message_templates.map((msgTemplate) => {
                      return (
                        <p key={msgTemplate.id}>
                          <Space>
                            <span className={CSS.font_size_xs + ' ' + CSS.font_weight_bold}>
                              {msgTemplate.percentage}%
                            </span>
                            <Tooltip title="Preview template">
                              <ButtonPreviewMessageTemplate
                                templates={templates}
                                selectedID={msgTemplate.id}
                              >
                                <Button type="link" size="small">
                                  {msgTemplate.id}
                                </Button>
                              </ButtonPreviewMessageTemplate>
                            </Tooltip>
                          </Space>
                        </p>
                      )
                    })}
                  </div>
                )
              }
            },
            {
              title: 'Status',
              key: 'status',
              render: (x: BroadcastCampaign) => {
                switch (x.status) {
                  case 'draft':
                    return <Tag color="blue">Draft</Tag>
                  case 'scheduled':
                    return <Tag color="purple">Scheduled</Tag>
                  case 'launched':
                    return <Tag color="gold">Launched</Tag>
                  case 'sent':
                    return <Tag color="green">Sent</Tag>
                  case 'failed':
                    return <Tag color="volcano">Failed</Tag>
                  default:
                    return <Tag color="default">{x.status}</Tag>
                }
              }
            },
            {
              title: 'Scheduled / launched at',
              key: 'createdAt',
              render: (x: BroadcastCampaign) => {
                if (x.scheduled_at) {
                  return (
                    <>
                      {dayjs(x.scheduled_at).format('lll')}
                      <div className={CSS.font_size_xs}>
                        in {x.timezone} -{' '}
                        {dayjs(x.scheduled_at)
                          .tz(accountCtx.account?.account.timezone as string)
                          .fromNow()}
                      </div>
                    </>
                  )
                }
                if (x.launched_at) {
                  return (
                    <>
                      {dayjs(x.launched_at)
                        .tz(accountCtx.account?.account.timezone as string)
                        .format('lll')}
                      <div className={CSS.font_size_xs}>
                        in {accountCtx.account?.account.timezone}
                      </div>
                    </>
                  )
                }
                return ''
              }
            },
            {
              title: (
                <div className={CSS.text_right}>
                  <Space>
                    <Tooltip title="Refresh">
                      <Button type="text" size="small" onClick={() => refetch()}>
                        <FontAwesomeIcon icon={faRefresh} />
                      </Button>
                    </Tooltip>
                  </Space>
                </div>
              ),
              key: 'actions',
              className: CSS.text_right,
              width: 50,
              render: (row: BroadcastCampaign) => (
                <div className={CSS.text_right}>
                  <Space>
                    {(row.status === 'draft' || row.status === 'scheduled') && (
                      <ButtonUpsertCampaign
                        btnProps={{ size: 'small', type: 'text' }}
                        onSuccess={() => refetch()}
                        campaign={row}
                      >
                        <Tooltip title="Edit campaign" placement="bottomRight">
                          <FontAwesomeIcon icon={faPenToSquare} />
                        </Tooltip>
                      </ButtonUpsertCampaign>
                    )}

                    {row.status === 'scheduled' && (
                      <Tooltip title="Pause campaign" placement="bottomRight">
                        <Popconfirm
                          title="Do your really want to pause the scheduled campaign?"
                          onConfirm={pauseCampaign.bind(null, row.id)}
                          okText="Pause campaign"
                          placement="topRight"
                          okButtonProps={{ loading: isLoadingAction }}
                        >
                          <Button type="text" size="small">
                            <FontAwesomeIcon icon={faPause} />
                          </Button>
                        </Popconfirm>
                      </Tooltip>
                    )}
                    {row.status === 'draft' && (
                      <Tooltip title="Launch campaign" placement="bottomRight">
                        <Popconfirm
                          title="Do your really want to launch the campaign?"
                          onConfirm={launchCampaign.bind(null, row.id)}
                          okText="Launch campaign"
                          placement="topRight"
                          okButtonProps={{ loading: isLoadingAction }}
                        >
                          <Button type="text" size="small">
                            <FontAwesomeIcon icon={faPlay} />
                          </Button>
                        </Popconfirm>
                      </Tooltip>
                    )}
                  </Space>
                </div>
              )
            }
          ]}
        />
      </div>
    </Layout>
  )
}

export default RouteBroadcasts
