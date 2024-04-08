import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import CSS from 'utils/css'
import { useQuery } from '@tanstack/react-query'
import { BroadcastCampaign } from 'interfaces'
import { Table, Tooltip } from 'antd'
import ButtonUpsertCampaign from './button_upsert_broadcast'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPenToSquare } from '@fortawesome/free-regular-svg-icons'

const RouteBroadcasts = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  const { isLoading, data, refetch, isFetching } = useQuery<BroadcastCampaign[]>(
    ['templates', workspaceCtx.workspace.id],
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
        <div className={CSS.top + ' ' + CSS.margin_l_l}>
          <h1>Broadcast campaigns</h1>
        </div>
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
            title: 'Template',
            key: 'id',
            render: (x: BroadcastCampaign) => (
              <div>
                <Tooltip title={x.id}>{x.name}</Tooltip>
              </div>
            )
          },
          // {
          //   title: 'Last update',
          //   key: 'createdAt',
          //   render: (x: BroadcastCampaign) =>
          //     dayjs(x.db_created_at)
          //       .tz(accountCtx.account?.account.timezone as string)
          //       .format('lll')
          // },
          {
            title: (
              <>
                {!isLoading && data && data.length > 0 && (
                  <ButtonUpsertCampaign btnProps={{ type: 'primary' }} onSuccess={() => refetch()}>
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
    </Layout>
  )
}

export default RouteBroadcasts
