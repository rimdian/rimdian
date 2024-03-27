import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import CSS from 'utils/css'
import { Table, Tooltip } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { MessageTemplate } from 'components/assets/message_template/interfaces'
import dayjs from 'dayjs'
import { useAccount } from 'components/login/context_account'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import ButtonUpsertTemplate from './button_upsert_email'
import { faPenToSquare } from '@fortawesome/free-regular-svg-icons'

const ListTemplates = () => {
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
    <Table
      pagination={false}
      dataSource={data}
      loading={isLoading || isFetching}
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
          title: 'Template',
          key: 'id',
          render: (x: MessageTemplate) => (
            <div>
              <Tooltip title={x.id}>{x.name}</Tooltip>
            </div>
          )
        },
        {
          title: 'Version',
          key: 'version',
          render: (x: MessageTemplate) => {
            return <div>{x.version}</div>
          }
        },
        {
          title: 'Channel',
          key: 'channel',
          render: (x: MessageTemplate) => {
            return <div>{x.channel}</div>
          }
        },
        {
          title: 'About',
          key: 'about',
          render: (x: MessageTemplate) => {
            if (x.channel === 'email') {
              return (
                <div>
                  <b>subject:</b> {x.email.subject}
                </div>
              )
            }
            return ''
          }
        },
        {
          title: 'Last update',
          key: 'createdAt',
          render: (x: MessageTemplate) =>
            dayjs(x.db_created_at)
              .tz(accountCtx.account?.account.timezone as string)
              .format('lll')
        },
        {
          title: (
            <>
              {!isLoading && data && data.length > 0 && (
                <ButtonUpsertTemplate btnProps={{ type: 'primary' }} onSuccess={() => refetch()}>
                  New template
                </ButtonUpsertTemplate>
              )}
            </>
          ),
          key: 'actions',
          className: 'actions',
          width: 170,
          render: (row: MessageTemplate) => (
            <div className={CSS.text_right}>
              <ButtonUpsertTemplate
                btnProps={{ size: 'small', type: 'text' }}
                onSuccess={() => refetch()}
                template={row}
              >
                <FontAwesomeIcon icon={faPenToSquare} />
              </ButtonUpsertTemplate>
            </div>
          )
        }
      ]}
    />
  )
}

export default ListTemplates
