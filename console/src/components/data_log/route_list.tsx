import {
  Table,
  TablePaginationConfig,
  Tooltip,
  Button,
  Tag,
  Space,
  Alert,
  Input,
  InputRef,
  Badge
} from 'antd'
import { DataHookState, DataLog } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import { QueryFunctionContext, useInfiniteQuery, useQueryClient } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { FilterValue } from 'antd/lib/table/interface'
import { forEach, map, truncate } from 'lodash'
import { paramsToQueryString } from 'utils/searchParams'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import dayjs from 'dayjs'
import { useAccount } from 'components/login/context_account'
import CSS from 'utils/css'
import { faArrowDown, faArrowUp, faRefresh } from '@fortawesome/free-solid-svg-icons'
import ButtonDataLogState from './button_about'
import ReprocessDataLogButton from './button_reprocess'
import GraphActivity from './graph_activity'
import { useMemo, useRef } from 'react'
import { Fullscreenable } from 'components/common/fullscreenable'
import TableTag from 'components/common/partial_table_tag'

interface DataLogList {
  data_logs: DataLog[]
  next_token?: string // next page: older rows = created before
  // previous_token?: string // previous page: newer rows = created after
}

interface DataLogListParams {
  limit: string
  next_token?: string
  // previous_token?: string
  // filters:
  id?: string
  origin?: string
  origin_id?: string
  event_at_since?: string
  event_at_until?: string
  checkpoint?: string
  has_error?: string
  kind?: string
  user_id?: string
  item_id?: string
}

const pageSize = 12
const RouteDataLogs = () => {
  const accountCtx = useAccount()
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [searchParams, setSearchParams] = useSearchParams()

  const searchInput = useRef<InputRef>(null)

  const params: DataLogListParams = {
    limit: searchParams.get('limit') || '' + pageSize,
    next_token: searchParams.get('next_token') || undefined,
    // previous_token: searchParams.get('previous_token') || undefined,
    // filters:
    id: searchParams.get('id') || undefined,
    origin: searchParams.get('origin') || undefined,
    kind: searchParams.get('kind') || undefined,
    event_at_since: searchParams.get('received_at_since') || undefined,
    event_at_until: searchParams.get('received_at_until') || undefined,
    has_error: searchParams.get('has_error') || undefined,
    checkpoint: searchParams.get('checkpoint') || undefined
  }
  // console.log(params)
  const queryKey = ['data_log', workspaceCtx.workspace.id, params]
  const queryClient = useQueryClient()

  const { data, error, fetchNextPage, hasNextPage, isFetching, isFetchingNextPage, refetch } =
    useInfiniteQuery(
      queryKey,
      (ctx: QueryFunctionContext): Promise<DataLogList> => {
        // console.log('fetching', ctx)
        let path =
          '/dataLog.list?' +
          paramsToQueryString(params) +
          '&workspace_id=' +
          workspaceCtx.workspace.id
        if (ctx.pageParam) {
          path += '&next_token=' + ctx.pageParam
        }
        return new Promise((resolve, reject) => {
          workspaceCtx
            .apiGET(path)
            .then((data: any) => {
              resolve(data as DataLogList)
            })
            .catch((e) => {
              reject(e)
            })
        })
      },
      {
        getNextPageParam: (lastPage: DataLogList) => {
          return lastPage?.next_token || undefined
        }
      }
    )

  const onTableChange = (
    _pagination: TablePaginationConfig,
    filters: Record<string, FilterValue | null>
    // _sorter: SorterResult<DataLog> | SorterResult<DataLog>[]
  ) => {
    // console.log('pagination', pagination);
    console.log('filters', filters)
    // console.log('sorter', sorter);

    const newParams: any = {}
    forEach(params, (val, key) => {
      // remove empty values and pagination tokens
      if (val !== undefined && key !== 'next_token' && key !== 'previous_token') {
        newParams[key] = val
      }
    })

    if (filters.id && filters.id.length) {
      newParams.id = filters.id[0]
    } else {
      delete newParams.id
    }

    if (filters.checkpoint && filters.checkpoint.length) {
      newParams.checkpoint = filters.checkpoint[0]
    } else {
      delete newParams.checkpoint
    }

    if (filters.has_error && filters.has_error.length) {
      newParams.has_error = filters.has_error[0]
    } else {
      delete newParams.has_error
    }

    if (filters.origin && filters.origin.length) {
      newParams.origin = filters.origin[0]
    } else {
      delete newParams.origin
    }

    if (filters.kind && filters.kind.length) {
      newParams.kind = filters.kind[0]
    } else {
      delete newParams.kind
    }

    setSearchParams(newParams)
  }

  const lines = useMemo(() => {
    if (!data || !data.pages.length) return []
    const lines: DataLog[] = []
    data.pages.forEach((page) => {
      lines.push(...page.data_logs)
    })
    return lines
  }, [data])

  const backToHead = () => {
    queryClient.setQueryData(queryKey, () => ({
      pages: [],
      pageParams: []
    }))

    // console.log(lines)
    refetch({ refetchPage: (_page, index) => index === 0 })
  }

  return (
    <Layout currentOrganization={workspaceCtx.organization} currentWorkspaceCtx={workspaceCtx}>
      <div className={CSS.top}>
        <h1>Data logs</h1>
      </div>

      <div className={CSS.margin_v_m}>
        <GraphActivity />
      </div>

      <Fullscreenable>
        {error ? (
          <Alert type="error" className={CSS.margin_b_l} message={<>{error}</>} closable showIcon />
        ) : null}
        <Table
          pagination={false}
          dataSource={lines}
          loading={isFetching}
          onChange={onTableChange}
          rowKey="id"
          summary={() => {
            return (
              <Table.Summary.Row>
                <Table.Summary.Cell index={0} colSpan={8}>
                  <div style={{ textAlign: 'center' }}>
                    <Space>
                      <Button loading={isFetching} size="small" onClick={() => backToHead()}>
                        <Space>
                          <FontAwesomeIcon icon={faArrowUp} />
                          Back to head
                        </Space>
                      </Button>
                      <Button
                        loading={isFetching || isFetchingNextPage}
                        size="small"
                        onClick={() => fetchNextPage()}
                      >
                        <Space>
                          <FontAwesomeIcon icon={faRefresh} />
                          Refresh lines
                        </Space>
                      </Button>
                      {hasNextPage && (
                        <Button
                          loading={isFetchingNextPage}
                          // type={hasNextPage ? 'primary' : 'default'}
                          size="small"
                          onClick={() => fetchNextPage()}
                        >
                          <Space>
                            <FontAwesomeIcon icon={faArrowDown} />
                            Load more
                          </Space>
                        </Button>
                      )}
                    </Space>
                  </div>
                </Table.Summary.Cell>
              </Table.Summary.Row>
            )
          }}
          columns={[
            {
              title: 'ID',
              key: 'id',
              width: 200,
              filterDropdown: ({ setSelectedKeys, selectedKeys, confirm, clearFilters }) => (
                <div style={{ padding: 8, width: 300 }} onKeyDown={(e) => e.stopPropagation()}>
                  <Input
                    ref={searchInput}
                    placeholder="Data log ID"
                    value={selectedKeys[0]}
                    onChange={(e) => setSelectedKeys(e.target.value ? [e.target.value] : [])}
                    onPressEnter={() => confirm()}
                    style={{ marginBottom: 8, display: 'block' }}
                  />
                  <div>
                    <Button
                      type="link"
                      onClick={() => {
                        if (clearFilters) clearFilters()
                        confirm()
                      }}
                      size="small"
                    >
                      Reset
                    </Button>
                    <Button
                      className={CSS.pull_right}
                      type="primary"
                      onClick={() => confirm()}
                      // icon={<SearchOutlined />}
                      size="small"
                      style={{ width: 90 }}
                    >
                      Search
                    </Button>
                  </div>
                </div>
              ),
              filtered: !!params.id,
              filteredValue: params.id ? [params.id] : null,
              onFilterDropdownOpenChange: (visible) => {
                if (visible) {
                  setTimeout(() => searchInput.current?.select(), 100)
                }
              },
              render: (x) => (
                <Tooltip title={x.id}>
                  <span className={CSS.font_size_xs}>{truncate(x.id, { length: 15 })}</span>
                </Tooltip>
              )
            },
            {
              title: 'Origin',
              key: 'origin',
              filteredValue: params.origin ? [params.origin] : null,
              filterMultiple: false,
              filters: [
                { text: 'Client', value: 0 },
                { text: 'Token', value: 1 },
                { text: 'Data log child', value: 2 },
                { text: 'Workflow', value: 3 },
                { text: 'Task', value: 4 }
              ],
              onFilter: (value, record) => record.origin === value,
              render: (x: DataLog) => {
                // DataLogOriginClient           DataLogOriginType = iota
                // DataLogOriginToken                              // 1 = API token
                // DataLogOriginInternalDataLog                    // 2 = Internal data_log item (ie: user_alias)
                // DataLogOriginInternalWorkflow                   // 3 = Internal workflow
                // DataLogOriginInternalTask                       // 4 = Internal task
                let origin = '' + x.origin
                switch (x.origin) {
                  case 0:
                    origin = 'Client'
                    break

                  case 1:
                    origin = 'Token'
                    break

                  case 2:
                    origin = 'Data log child'
                    break

                  case 3:
                    origin = 'Workflow'
                    break

                  case 4:
                    origin = 'Task'
                    break

                  default:
                }
                return <Tooltip title={x.origin_id}>{origin}</Tooltip>
              }
            },
            {
              title: 'Received at',
              key: 'received_at',
              render: (x: DataLog) => (
                <span className={CSS.font_size_xs}>
                  {dayjs(x.context.received_at)
                    .tz(accountCtx.account?.account.timezone as string)
                    .format('lll')}
                </span>
              )
            },
            {
              title: 'Action',
              key: 'action',
              render: (x: DataLog) => {
                return x.action
              }
            },
            {
              title: 'Kind',
              key: 'kind',
              filterDropdown: ({ setSelectedKeys, selectedKeys, confirm, clearFilters }) => (
                <div style={{ padding: 8, width: 300 }} onKeyDown={(e) => e.stopPropagation()}>
                  <Input
                    ref={searchInput}
                    placeholder="Kind"
                    value={selectedKeys[0]}
                    onChange={(e) => setSelectedKeys(e.target.value ? [e.target.value] : [])}
                    onPressEnter={() => confirm()}
                    style={{ marginBottom: 8, display: 'block' }}
                  />
                  <div>
                    <Button
                      type="link"
                      onClick={() => {
                        if (clearFilters) clearFilters()
                        confirm()
                      }}
                      size="small"
                    >
                      Reset
                    </Button>
                    <Button
                      className={CSS.pull_right}
                      type="primary"
                      onClick={() => confirm()}
                      // icon={<SearchOutlined />}
                      size="small"
                      style={{ width: 90 }}
                    >
                      Search
                    </Button>
                  </div>
                </div>
              ),
              filtered: !!params.kind,
              filteredValue: params.kind ? [params.kind] : null,
              onFilterDropdownOpenChange: (visible) => {
                if (visible) {
                  setTimeout(() => searchInput.current?.select(), 100)
                }
              },
              render: (x: DataLog) => {
                return (
                  <Tooltip
                    title={
                      <>
                        <p>Ext. ID: {x.item_external_id}</p>
                        ID: {x.item_id}
                      </>
                    }
                  >
                    <TableTag table={x.kind} />
                  </Tooltip>
                )
              }
            },
            {
              title: 'Has error',
              key: 'has_error',
              filterMultiple: false,
              filteredValue: params.has_error ? [params.has_error] : null,
              filters: [
                {
                  text: 'Aborted',
                  value: '2'
                },
                {
                  text: 'Retrying error...',
                  value: '1'
                },
                {
                  text: 'None',
                  value: '0'
                }
              ],
              onFilter: (value, record) => '' + record.has_error === value,
              render: (x: DataLog) => {
                let tag = <></>
                // 0: none, 1: retryable, 2: not retryable
                switch (x.has_error) {
                  case 2:
                    tag = <Tag color="red">Aborted</Tag>
                    break
                  case 1:
                    tag = <Tag color="orange">Retrying error...</Tag>
                    break
                  case 0:
                    tag = <Tag color="green">None</Tag>
                    break
                  default:
                    tag = <Tag>{x.has_error}</Tag>
                }

                return (
                  <>
                    {tag}
                    {x.errors && Object.keys(x.errors).length > 0 && (
                      <>
                        {map(x.errors, (value: string, key: number) => (
                          <div
                            style={{ wordBreak: 'break-all' }}
                            className={CSS.text_red + ' ' + CSS.font_size_xs}
                            key={key}
                          >
                            <p>
                              <b>{key}</b>: {value}
                            </p>
                          </div>
                        ))}
                      </>
                    )}
                  </>
                )
              }
            },
            {
              title: 'Checkpoint',
              key: 'checkpoint',
              filterMultiple: false,
              filteredValue: params.checkpoint ? [params.checkpoint] : null,
              filters: [
                {
                  text: 'Pending',
                  value: '0'
                },
                {
                  text: 'Hook on_validation executed',
                  value: '10'
                },
                {
                  text: 'Data extracted and persisted in DB',
                  value: '20'
                },
                {
                  text: 'Item upserted',
                  value: '30'
                },
                {
                  text: 'Conversions attributed',
                  value: '40'
                },
                {
                  text: 'Segment processed',
                  value: '50'
                },
                {
                  text: 'Workflow triggered',
                  value: '60'
                },
                {
                  text: 'Hooks on_success executed',
                  value: '70'
                },
                {
                  text: 'Done',
                  value: '100'
                }
              ],
              onFilter: (value, record) => '' + record.checkpoint === value,
              render: (x: DataLog) => {
                // DataLogStatusErrorNotRetryable            DataLogStatusType = -2  // done with non retryable error
                // DataLogStatusErrorRetryable               DataLogStatusType = -1  // done with retryable error
                // DataLogStatusPending                      DataLogStatusType = 0   // waiting to be processed
                // DataLogStatusHookOnValidationExecuted DataLogStatusType = 10  // hook on_validation executed
                // DataLogStatusPersisted                    DataLogStatusType = 20  // data_log data extracted and persisted in DB
                // DataLogStatusItemUpserted                 DataLogStatusType = 30  // item upserted
                // DataLogStatusConversionsAttributed        DataLogStatusType = 40  // conversions attributed
                // DataLogStatusSegmentsRecomputed           DataLogStatusType = 50  // segment processed
                // DataLogStatusWorkflowsTriggered           DataLogStatusType = 60  // workflow triggered
                // DataLogStatusHooksFinalizeExecuted        DataLogStatusType = 70  // hooks finalize executed
                // DataLogStatusDeleteAfterMerge             DataLogStatusType = 80  // should delete after user_alias merge
                // DataLogStatusDone                         DataLogStatusType = 100 // all done

                switch (x.checkpoint) {
                  case 0:
                    return <Tag color="blue">Pending...</Tag>

                  case 10:
                    return <Tag color="blue">Hook on_validation executed</Tag>

                  case 20:
                    return <Tag color="blue">Data extracted and persisted in DB</Tag>

                  case 30:
                    return <Tag color="blue">Item upserted</Tag>

                  case 40:
                    return <Tag color="blue">Conversions attributed</Tag>

                  case 50:
                    return <Tag color="blue">Segment processed</Tag>

                  case 60:
                    return <Tag color="blue">Workflow triggered</Tag>

                  case 70:
                    return <Tag color="blue">Hooks finalize executed</Tag>

                  case 100:
                    return <Tag color="green">Done</Tag>

                  default:
                    return <>checkpoint: {x.checkpoint}</>
                }
              }
            },
            {
              title: 'Hooks',
              key: 'hooks',
              render: (x: DataLog) => {
                if (x.hooks && Object.keys(x.hooks).length > 0) {
                  return map(x.hooks, (state: DataHookState, hookID: number) => {
                    return (
                      <div
                        style={{ wordBreak: 'break-all' }}
                        className={CSS.font_size_xs}
                        key={hookID}
                      >
                        <Tooltip title={state.msg || ''}>
                          {state.done && (
                            <Badge status={state.err ? 'error' : 'success'} text={hookID} />
                          )}
                          {!state.done && <Badge status="processing" text={hookID} />}
                        </Tooltip>
                      </div>
                    )
                  })
                }
                return '-'
              }
            },
            // {
            //   title: 'Took',
            //   key: 'took',
            //   width: 160,
            //   render: (x: DataLog) => {
            //     if (x.checkpoint === 100) {
            //       return (
            //         <>
            //           <Tooltip
            //             title={
            //               <span>
            //                 Received at:{' '}
            //                 {dayjs(x.context.received_at)
            //                   .tz(accountCtx.account?.account.timezone as string)
            //                   .format('lll')}
            //                 <br />
            //                 Updated at:{' '}
            //                 {dayjs(x.db_updated_at)
            //                   .tz(accountCtx.account?.account.timezone as string)
            //                   .format('lll')}
            //                 <br />
            //                 In {accountCtx.account?.account.timezone}
            //               </span>
            //             }
            //           >
            //             <span className={CSS.font_size_xs}>
            //               {dayjs.duration(dayjs(x.db_created_at).diff(x.db_updated_at)).humanize()}
            //             </span>
            //           </Tooltip>
            //         </>
            //       )
            //     }
            //   }
            // },
            // {
            //   title: 'Errors?',
            //   key: 'errors',
            //   render: (x: DataLog) => {
            //     if (x.errors && Object.keys(x.errors).length > 0) {
            //       return map(x.errors, (value: string, key: number) => {
            //         return (
            //           <div
            //             style={{ wordBreak: 'break-all' }}
            //             className={CSS.text_red + ' ' + CSS.font_size_xs}
            //             key={key}
            //           >
            //             <p>
            //               <b>{key}</b>: {value}
            //             </p>
            //           </div>
            //         )
            //       })
            //     }
            //     return '-'
            //   }
            // },
            {
              title: '',
              key: 'actions',
              className: 'actions',
              width: 200,
              render: (row: DataLog) => (
                <div className={CSS.text_right}>
                  {/* check if row has been received more than 5 minutes ago, we can reprocess it */}
                  {row.checkpoint !== 100 &&
                    (dayjs().diff(dayjs(row.db_created_at), 'minutes') > 5 ||
                      window.Config.ENV === 'development') && (
                      <ReprocessDataLogButton
                        dataLogId={row.id}
                        workspaceId={workspaceCtx.workspace.id}
                        apiPOST={workspaceCtx.apiPOST}
                        onComplete={() => {
                          refetch()
                        }}
                        btnSize="small"
                        className={CSS.margin_r_xs}
                      />
                    )}
                  <ButtonDataLogState
                    workspaceId={workspaceCtx.workspace.id}
                    apiGET={workspaceCtx.apiGET}
                    dataImport={row}
                    accountTimezone={accountCtx.account?.account.timezone as string}
                  />
                </div>
              )
            }
          ]}
        />
      </Fullscreenable>
    </Layout>
  )
}

export default RouteDataLogs
