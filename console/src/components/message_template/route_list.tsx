import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import CSS from 'utils/css'
import { Button, Table, Tooltip } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { MessageTemplate } from 'components/message_template/interfaces'
import dayjs from 'dayjs'
import { useAccount } from 'components/login/context_account'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faRefresh } from '@fortawesome/free-solid-svg-icons'
import ButtonUpsertTemplate from './button_upsert_email'

const RouteTemplates = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const accountCtx = useAccount()

  const { isLoading, data, refetch, isFetching } = useQuery<MessageTemplate[]>(
    ['templates', workspaceCtx.workspace.id],
    (): Promise<MessageTemplate[]> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET('/messageTemplate.list?workspace_id=' + workspaceCtx.workspace.id)
          .then((data: any) => {
            // console.log(data)
            resolve(data as MessageTemplate[])
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
          <h1>Templates</h1>
          <div className={CSS.topSeparator}></div>
          <div></div>
        </div>

        <Table
          pagination={false}
          dataSource={data}
          loading={isLoading}
          rowKey="id"
          locale={{
            emptyText: (
              <>
                <p>No templates found</p>
                <ButtonUpsertTemplate btnProps={{ type: 'primary' }} onSuccess={() => refetch()}>
                  New template
                </ButtonUpsertTemplate>
              </>
            )
          }}
          columns={[
            {
              title: 'Templae',
              key: 'id',
              render: (x: MessageTemplate) => (
                <div>
                  <Tooltip title={x.id}>{x.name}</Tooltip>
                </div>
              )
            },
            {
              title: 'Channel',
              key: 'channel',
              render: (x: MessageTemplate) => {
                return <div>{x.channel}</div>
              }
            },
            {
              title: 'Created',
              key: 'createdAt',
              render: (x: MessageTemplate) =>
                dayjs(x.db_created_at)
                  .tz(accountCtx.account?.account.timezone as string)
                  .format('lll')
            },
            {
              title: 'Updated',
              key: 'updatedAt',
              render: (x: MessageTemplate) => (
                <Tooltip
                  title={
                    <span>
                      {dayjs(x.db_updated_at)
                        .tz(accountCtx.account?.account.timezone as string)
                        .format('lll')}{' '}
                      in {accountCtx.account?.account.timezone}
                    </span>
                  }
                >
                  {dayjs(x.db_updated_at).fromNow()}
                </Tooltip>
              )
            },
            {
              title: '',
              key: 'actions',
              className: 'actions',
              width: 170,
              render: (row: MessageTemplate) => <div className={CSS.text_right}></div>
            }
          ]}
        />
      </div>
    </Layout>
  )
}

export default RouteTemplates
