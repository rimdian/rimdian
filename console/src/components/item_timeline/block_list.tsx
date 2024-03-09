import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { QueryFunctionContext, useInfiniteQuery } from '@tanstack/react-query'
import { Button, Space, Table, Tooltip } from 'antd'
import TableTag from 'components/common/partial_table_tag'
import { DataLog } from 'interfaces'
import { truncate } from 'lodash'
import { useMemo } from 'react'
import CSS from 'utils/css'
import { paramsToQueryString } from 'utils/searchParams'

type BlockDataLogProps = {
  workspaceId: string
  apiGET: (endpoint: string) => Promise<any>
  userId?: string
  origin?: number
  originId?: string
  parentDataLog?: DataLog
  kind?: string
  itemId?: string
  limit?: number
}

interface DataLogResult {
  data_logs: DataLog[]
  next_token: string
}

export const BlockDataLog = (props: BlockDataLogProps) => {
  const queryKey = [
    'data_log',
    props.workspaceId,
    props.userId,
    props.origin,
    props.originId,
    props.itemId,
    props.kind,
    props.limit
  ]
  //   const queryClient = useQueryClient()

  const { data, fetchNextPage, hasNextPage, isFetching, isFetchingNextPage } = useInfiniteQuery(
    queryKey,
    (ctx: QueryFunctionContext): Promise<DataLogResult> => {
      // console.log('fetching', ctx)
      const params: any = {}
      if (props.userId) params['user_id'] = props.userId
      if (props.origin) params['origin'] = props.origin
      if (props.originId) params['origin_id'] = props.originId
      if (props.parentDataLog) params['origin_id'] = props.parentDataLog.id
      if (props.itemId) params['item_id'] = props.itemId
      if (props.kind) params['kind'] = props.kind
      if (props.limit) params['limit'] = props.limit

      let path =
        '/dataLog.list?' + paramsToQueryString(params) + '&workspace_id=' + props.workspaceId
      if (ctx.pageParam) {
        path += '&next_token=' + ctx.pageParam
      }
      return new Promise((resolve, reject) => {
        props
          .apiGET(path)
          .then((data: any) => {
            //   console.log('data', data)

            // add parent data log to the first page
            if (props.parentDataLog) data.data_logs.push(props.parentDataLog)

            resolve(data as DataLogResult)
          })
          .catch(reject)
      })
    },
    {
      getNextPageParam: (lastPage: DataLogResult) => {
        return lastPage?.next_token || undefined
      }
    }
  )

  const lines = useMemo(() => {
    if (!data || !data.pages.length) return []
    const lines: DataLog[] = []
    data.pages.forEach((page) => {
      lines.push(...page.data_logs)
    })
    return lines
  }, [data])

  return (
    <div>
      <Table
        size="small"
        showHeader={false}
        rowKey="id"
        loading={isFetching || isFetchingNextPage}
        dataSource={lines}
        pagination={false}
        summary={() => {
          if (!hasNextPage) return

          return (
            <Table.Summary.Row>
              <Table.Summary.Cell index={0} colSpan={8}>
                <div style={{ textAlign: 'center' }}>
                  <Space>
                    {/* <Button loading={isFetching} size="small" onClick={() => backToHead()}>
                      <Space>
                        <FontAwesomeIcon icon={faArrowUp} />
                        Back to head
                      </Space>
                    </Button> */}
                    {/* <Button
                      loading={isFetching || isFetchingNextPage}
                      size="small"
                      onClick={() => fetchNextPage()}
                    >
                      <Space>
                        <FontAwesomeIcon icon={faRefresh} />
                        Refresh lines
                      </Space>
                    </Button> */}
                    <Button
                      loading={isFetchingNextPage}
                      // type={hasNextPage ? 'primary' : 'default'}
                      size="small"
                      onClick={() => fetchNextPage()}
                    >
                      <Space>
                        <FontAwesomeIcon icon={faPlus} />
                        Load more
                      </Space>
                    </Button>
                  </Space>
                </div>
              </Table.Summary.Cell>
            </Table.Summary.Row>
          )
        }}
        columns={[
          {
            key: 'action',
            width: 100,
            render: (_value: any, record: DataLog) => {
              return <span>{record.action}</span>
            }
          },
          {
            key: 'kind',
            render: (_value: any, record: DataLog) => {
              return (
                <span>
                  <TableTag table={record.kind} />
                </span>
              )
            }
          },
          {
            key: 'id',
            render: (_value: any, record: DataLog) => {
              return (
                <Tooltip title={record.item_id}>{truncate(record.item_id, { length: 13 })}</Tooltip>
              )
            }
          },
          {
            key: 'update_field',
            render: (_value: any, record: DataLog) => {
              if (record.action !== 'update') return
              return record.updated_fields.map((field: any, i: number) => {
                return <div key={i}>{field.field}</div>
              })
            }
          },
          {
            key: 'update_prev',
            render: (_value: any, record: DataLog) => {
              if (record.action !== 'update') return
              return record.updated_fields.map((field: any, i: number) => {
                return (
                  <div key={i} className={CSS.text_right}>
                    {field.previous === '' && '(empty)'}
                    {field.previous === null && '(null)'}
                    {field.previous !== null && field.previous !== '' && field.previous}
                  </div>
                )
              })
            }
          },
          {
            key: 'arrow',
            render: (_value: any, record: DataLog) => {
              if (record.action !== 'update') return
              return record.updated_fields.map((_field: any, i: number) => {
                return <div key={i}>→</div>
              })
            }
          },
          {
            key: 'update_new',
            render: (_value: any, record: DataLog) => {
              if (record.action !== 'update') return
              return record.updated_fields.map((field: any, i: number) => {
                return <div key={i}>{field.new === '' ? '(empty)' : field.new}</div>
              })
            }
          }
        ]}
      />
    </div>
  )
}
