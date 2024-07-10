import { Tag, Table, TablePaginationConfig, Tooltip, Button, message, Alert } from 'antd'
import { TaskExec } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { FilterValue } from 'antd/lib/table/interface'
import { forEach } from 'lodash'
import { paramsToQueryString } from 'utils/searchParams'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useAccount } from 'components/login/context_account'
import CSS from 'utils/css'
import { faChevronLeft, faChevronRight, faRefresh } from '@fortawesome/free-solid-svg-icons'
import ButtonTaskAbout from './button_about'
import dayjs from 'dayjs'
import ButtonAbortTask from './button_abort'

interface TaskExecList {
  task_execs: TaskExec[]
  next_token?: string // next page: older rows = created before
  previous_token?: string // previous page: newer rows = created after
}

interface TaskExecListParams {
  limit: string
  next_token?: string
  previous_token?: string
  // filters:
  task_id?: string
  status?: string
}

const pageSize = 25
const RouteTasks = () => {
  const accountCtx = useAccount()
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [searchParams, setSearchParams] = useSearchParams()

  const params: TaskExecListParams = {
    limit: searchParams.get('limit') || '' + pageSize,
    next_token: searchParams.get('next_token') || undefined,
    previous_token: searchParams.get('previous_token') || undefined,
    // filters:
    task_id: searchParams.get('task_id') || undefined,
    status: searchParams.get('status') || undefined
  }
  // console.log(params)

  // task execs
  const { isLoading, data, refetch, isFetching } = useQuery<TaskExecList>(
    ['tasks', workspaceCtx.workspace.id, params],
    (): Promise<TaskExecList> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET(
            '/taskExec.list?' +
              paramsToQueryString(params) +
              '&workspace_id=' +
              workspaceCtx.workspace.id
          )
          .then((data: any) => {
            // console.log(data)
            resolve(data as TaskExecList)
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  const nextPage = (next_token?: string) => {
    if (!next_token) return
    const newParams: any = {}
    forEach(params, (val, key) => {
      if (val !== undefined) {
        newParams[key] = val
      }
    })
    delete newParams.previous_token
    newParams.next_token = next_token
    setSearchParams(newParams)
  }

  const previousPage = (previous_token?: string) => {
    if (!previous_token) return
    const newParams: any = {}
    forEach(params, (val, key) => {
      if (val !== undefined) {
        newParams[key] = val
      }
    })
    delete newParams.next_token
    newParams.previous_token = previous_token
    setSearchParams(newParams)
  }

  const onTableChange = (
    pagination: TablePaginationConfig,
    filters: Record<string, FilterValue | null>
  ) => {
    // console.log('pagination', pagination);
    // console.log('filters', filters);
    // console.log('sorter', sorter);

    const newParams: any = {}
    forEach(params, (val, key) => {
      if (val !== undefined) {
        newParams[key] = val
      }
    })
    newParams.page = pagination.current || 1

    if (filters.status && filters.status.length) {
      newParams.status = filters.status[0]
    } else {
      delete newParams.status
    }

    if (filters.task_id && filters.task_id.length) {
      newParams.task_id = filters.task_id[0]
    } else {
      delete newParams.task_id
    }

    setSearchParams(newParams)
  }

  const alertTypeFromStatus = (status: number) => {
    if (status === -2) {
      return 'error'
    }
    if (status === -1) {
      return 'warning'
    }
    if (status === 0) {
      return 'info'
    }
    if (status === 1) {
      return 'success'
    }
    return 'info'
  }

  return (
    <Layout currentOrganization={workspaceCtx.organization} currentWorkspaceCtx={workspaceCtx}>
      <div className={CSS.top}>
        <h1>Task execs</h1>
      </div>

      <Table
        pagination={false}
        dataSource={data?.task_execs}
        loading={isLoading}
        onChange={onTableChange}
        rowKey="id"
        columns={[
          {
            title: 'Task',
            key: 'task_id',
            filteredValue: params.task_id ? [params.task_id] : null,
            filterMultiple: false,
            filters: [
              {
                text: 'Generate demo',
                value: 'system_generate_demo'
              }
            ],
            onFilter: (value, record) => record.task_id === value,
            render: (x: TaskExec) => (
              <div>
                {x.task_id === 'webhook' && <Tag color="blue">Webhook</Tag>}
                <Tooltip title={x.id}>{x.name}</Tooltip>
                {x.message && (
                  <Alert
                    message={<small style={{ wordBreak: 'break-all' }}>{x.message}</small>}
                    type={alertTypeFromStatus(x.status)}
                    className={CSS.margin_v_m}
                  />
                )}
              </div>
            )
          },
          {
            title: 'Status',
            key: 'status',
            filterMultiple: false,
            filteredValue: params.status ? [params.status] : null,
            filters: [
              {
                text: 'Aborted',
                value: '-2'
              },
              {
                text: 'Retrying error...',
                value: '-1'
              },
              {
                text: 'Processing...',
                value: '0'
              },
              {
                text: 'Success',
                value: '1'
              }
            ],
            onFilter: (value, record) => '' + record.status === value,
            render: (x: TaskExec) => {
              let tag = <></>
              if (x.status) {
                if (x.status === -2) {
                  tag = <Tag color="red">Aborted</Tag>
                }
                if (x.status === -1) {
                  tag = <Tag color="orange">Retrying error...</Tag>
                }
                if (x.status === 0) {
                  tag = <Tag color="blue">Processing...</Tag>
                }
                if (x.status === 1) {
                  tag = <Tag color="green">Done</Tag>
                }
              }
              return tag
            }
          },
          {
            title: 'Created',
            key: 'createdAt',
            render: (x: TaskExec) =>
              dayjs(x.db_created_at)
                .tz(accountCtx.account?.account.timezone as string)
                .format('lll')
          },
          {
            title: 'Updated',
            key: 'updatedAt',
            render: (x: TaskExec) => (
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
            title: (
              <div className={CSS.text_right}>
                <Button.Group>
                  <Button
                    size="small"
                    disabled={!data?.previous_token}
                    className={CSS.pull_right}
                    onClick={previousPage.bind(null, data?.previous_token)}
                  >
                    <FontAwesomeIcon icon={faChevronLeft} />
                  </Button>
                  <Tooltip title="Refresh">
                    <Button size="small" onClick={() => refetch()} disabled={isFetching}>
                      <FontAwesomeIcon spin={isFetching} icon={faRefresh} />
                    </Button>
                  </Tooltip>
                  <Button
                    disabled={!data?.next_token}
                    size="small"
                    className={CSS.pull_right}
                    onClick={nextPage.bind(null, data?.next_token)}
                  >
                    <FontAwesomeIcon icon={faChevronRight} />
                  </Button>
                </Button.Group>
              </div>
            ),
            key: 'actions',
            className: 'actions',
            width: 170,
            render: (row: TaskExec) => (
              <div className={CSS.text_right}>
                <ButtonAbortTask
                  onAbort={() => {
                    refetch().then(() => {
                      message.success('Task successfully aborted')
                    })
                  }}
                  taskExec={row}
                  workspaceId={workspaceCtx.workspace.id}
                  apiPOST={workspaceCtx.apiPOST}
                />
                <ButtonTaskAbout
                  workspaceId={workspaceCtx.workspace.id}
                  apiGET={workspaceCtx.apiGET}
                  taskExec={row}
                  accountTimezone={accountCtx.account?.account.timezone as string}
                />
              </div>
            )
          }
        ]}
      />
    </Layout>
  )
}

export default RouteTasks
