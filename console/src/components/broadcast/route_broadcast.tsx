import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import CSS from 'utils/css'
import { useQuery } from '@tanstack/react-query'
import { BroadcastCampaign } from 'interfaces'
import { Table, Tag, Tooltip } from 'antd'
import ButtonUpsertCampaign from './button_upsert_broadcast'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPenToSquare } from '@fortawesome/free-regular-svg-icons'
import dayjs from 'dayjs'
import { useAccount } from 'components/login/context_account'

const RouteBroadcasts = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const accountCtx = useAccount()

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

  return (
    <Layout
      currentOrganization={workspaceCtx.organization}
      currentWorkspace={workspaceCtx.workspace}
    >
      <div className={CSS.container}>
        <div className={CSS.top}>
          <h1>Broadcast campaigns</h1>
        </div>

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
                      <Tag key={list.id} color={subscriptionList?.color}>
                        {subscriptionList?.name}
                      </Tag>
                    )
                  })}
                </div>
              )
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
                      <div className={CSS.font_size_xs}>in {x.timezone}</div>
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
                <>
                  {!isLoading && data && data.length > 0 && (
                    <ButtonUpsertCampaign
                      btnProps={{ type: 'primary' }}
                      onSuccess={() => refetch()}
                    >
                      New campaign
                    </ButtonUpsertCampaign>
                  )}
                </>
              ),
              key: 'actions',
              className: 'actions',
              width: 170,
              render: (row: BroadcastCampaign) => (
                <div className={CSS.text_right}>
                  <ButtonUpsertCampaign
                    btnProps={{ size: 'small', type: 'text' }}
                    onSuccess={() => refetch()}
                    campaign={row}
                  >
                    <FontAwesomeIcon icon={faPenToSquare} />
                  </ButtonUpsertCampaign>
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
