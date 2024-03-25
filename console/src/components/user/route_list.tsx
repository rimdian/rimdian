import {
  Tag,
  Table,
  TablePaginationConfig,
  Tooltip,
  Button,
  Row,
  Col,
  Popconfirm,
  Badge,
  Space
} from 'antd'
import { SubscriptionList, User } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { FilterValue } from 'antd/lib/table/interface'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { useAccount } from 'components/login/context_account'
import {
  faChevronLeft,
  faChevronRight,
  faFemale,
  faLocationDot,
  faMale,
  faRefresh,
  faSearch
} from '@fortawesome/free-solid-svg-icons'
import numbro from 'numbro'
import dayjs from 'dayjs'
import CSS from 'utils/css'
import Block, { blockCss } from 'components/common/block'
import { css } from '@emotion/css'
import ButtonUpsertSegment from 'components/segment/button_upsert'
import { forEach } from 'lodash'
import { Segment } from 'components/segment/interfaces'
import { useMemo } from 'react'
import { faTrashAlt } from '@fortawesome/free-regular-svg-icons'
import ButtonUpsertSubscriptionList from 'components/subscription_list/button_upsert'

interface UserList {
  users: User[]
  next_token?: string // next page: older rows = created before
  previous_token?: string // previous page: newer rows = created after
}

// interface UserListParams {
//   limit: string
//   next_token?: string
//   previous_token?: string
//   // filters:
//   user_id?: string
//   is_authenticated?: string
// }

const pageSize = 12

const secondMenuCSS = {
  title: css({
    padding: CSS.M,
    fontWeight: 'bold'
  }),

  counter: css(
    {
      fontSize: '10px',
      paddingLeft: CSS.S
    },
    CSS.font_weight_semibold
  ),

  list: css({
    listStyleType: 'none',
    margin: 0,
    padding: 0,

    '& > li': {
      cursor: 'pointer',
      padding: CSS.S + ' ' + CSS.L,
      borderLeft: '2px solid rgba(0,0,0,0)',
      '&:hover': {
        borderLeft: '2px solid rgba(78,108,255,0.4);' //rgba(225,245,254 ,0.7);
      }
    },

    '& > li.active': {
      borderLeft: '2px solid rgba(78,108,255,0.7);' //rgba(225,245,254 ,0.7);
    }
  })
}

const RouteUsers = () => {
  const accountCtx = useAccount()
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [searchParams, setSearchParams] = useSearchParams()

  // users
  const { isLoading, data, refetch, isFetching } = useQuery<UserList>(
    ['users', workspaceCtx.workspace.id, searchParams.toString()],
    (): Promise<UserList> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET(
            '/user.list?limit=' +
              pageSize +
              // paramsToQueryString(params) +
              '&workspace_id=' +
              workspaceCtx.workspace.id +
              '&' +
              searchParams.toString()
          )
          .then((data: any) => {
            resolve(data as UserList)
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
    searchParams.forEach((val, key) => {
      // console.log(key, val)
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
    searchParams.forEach((val, key) => {
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
    searchParams.forEach((val, key) => {
      if (val !== undefined) {
        newParams[key] = val
      }
    })
    newParams.page = pagination.current || 1

    if (filters.is_authenticated && filters.is_authenticated.length) {
      newParams.is_authenticated = filters.is_authenticated[0]
    } else {
      delete newParams.is_authenticated
    }

    setSearchParams(newParams)
  }

  // add showUser param to url
  const showUser = (userExternalId: string) => {
    searchParams.append('showUser', userExternalId)
    setSearchParams(searchParams)
  }

  const segments: Segment[] = useMemo(() => {
    const list = [workspaceCtx.segmentsMap.anonymous, workspaceCtx.segmentsMap.authenticated]

    forEach(workspaceCtx.segmentsMap, (segment: Segment) => {
      if (segment.id !== '_all' && segment.id !== 'anonymous' && segment.id !== 'authenticated') {
        list.push(segment)
      }
    })

    return list
  }, [workspaceCtx.segmentsMap])

  const currentSegment = useMemo(() => {
    if (searchParams.get('segment_id') === '_all' || !searchParams.get('segment_id')) {
      return workspaceCtx.segmentsMap['_all']
    }

    return workspaceCtx.segmentsMap[searchParams.get('segment_id') as string]
  }, [workspaceCtx.segmentsMap, searchParams])

  const currentList = useMemo(() => {
    if (!searchParams.get('list_id')) {
      return undefined
    }

    return workspaceCtx.subscriptionLists.find(
      (list: SubscriptionList) => list.id === searchParams.get('list_id')
    )
  }, [searchParams, workspaceCtx.subscriptionLists])

  return (
    <Layout
      currentOrganization={workspaceCtx.organization}
      currentWorkspace={workspaceCtx.workspace}
    >
      <Row gutter={[16, 16]}>
        <Col span={5}>
          <div className={CSS.top}>
            <h1>Segments</h1>
            <div className={CSS.topSeparator}></div>
          </div>

          <Block>
            <ul className={secondMenuCSS.list}>
              <li
                onClick={() => setSearchParams({})}
                className={currentSegment.id === '_all' ? 'active' : ''}
              >
                All users
                <span className={secondMenuCSS.counter}>
                  {workspaceCtx.segmentsMap._all &&
                    numbro(workspaceCtx.segmentsMap['_all'].users_count).format({
                      totalLength: 3,
                      trimMantissa: true
                    })}
                </span>
              </li>
              {segments.map((segment: Segment) => {
                return (
                  <li
                    key={segment.id}
                    onClick={() => setSearchParams({ segment_id: segment.id })}
                    className={currentSegment.id === segment.id ? 'active' : ''}
                  >
                    {segment.status === 'building' && (
                      <Tooltip className={CSS.pull_right} title="Building...">
                        <Badge status="processing" />
                      </Tooltip>
                    )}
                    <Tag color={segment.color}>{segment.name}</Tag>
                    <span className={secondMenuCSS.counter}>
                      {workspaceCtx.segmentsMap.anonymous &&
                        numbro(segment.users_count).format({
                          totalLength: 3,
                          trimMantissa: true
                        })}
                    </span>
                  </li>
                )
              })}
            </ul>
          </Block>
          <ButtonUpsertSegment />

          <div className={CSS.top + ' ' + CSS.margin_t_l}>
            <h1>Subscription lists</h1>
            <div className={CSS.topSeparator}></div>
          </div>

          {workspaceCtx.subscriptionLists.length > 0 && (
            <Block>
              <ul className={secondMenuCSS.list}>
                {workspaceCtx.subscriptionLists.map((list: SubscriptionList) => {
                  return (
                    <li
                      key={list.id}
                      onClick={() => setSearchParams({ list_id: list.id })}
                      className={currentList?.id === list.id ? 'active' : ''}
                    >
                      <Tag color={list.color}>{list.name}</Tag>
                      <span className={secondMenuCSS.counter}>
                        {numbro(list.users_count).format({
                          totalLength: 3,
                          trimMantissa: true
                        })}
                      </span>
                    </li>
                  )
                })}
              </ul>
            </Block>
          )}
          <ButtonUpsertSubscriptionList />
        </Col>

        <Col span={19}>
          <div className={CSS.top}>
            <h1>Users</h1>
            {currentSegment.id !== '_all' && (
              <>
                <Space style={{ marginTop: 2, lineHeight: '20px', height: '20px', marginLeft: 20 }}>
                  <Tag color={currentSegment.color}>{currentSegment.name}</Tag>

                  {currentSegment.status === 'building' && (
                    <Badge status="processing" text="Building" />
                  )}
                  {currentSegment.status === 'active' && <Badge status="success" text="Active" />}
                  {currentSegment.status === 'deleted' && <Badge status="error" text="Deleted" />}
                </Space>

                <div className={CSS.topSeparator}></div>
                {currentSegment.id !== 'anonymous' && currentSegment.id !== 'authenticated' && (
                  <Space>
                    <Popconfirm
                      title="Do you really want to delete this segment?"
                      okText="Delete"
                      okButtonProps={{ danger: true }}
                      cancelText="No"
                    >
                      <Button type="text" size="small">
                        <FontAwesomeIcon icon={faTrashAlt} />
                      </Button>
                    </Popconfirm>
                    <ButtonUpsertSegment segment={currentSegment} />
                  </Space>
                )}
              </>
            )}
          </div>
          <Table
            pagination={false}
            dataSource={data?.users}
            loading={isLoading}
            onChange={onTableChange}
            rowKey="id"
            className={blockCss.self}
            columns={[
              {
                title: 'External ID',
                key: 'id',
                render: (x) => <Tooltip title={'Internal ID: ' + x.id}>{x.external_id}</Tooltip>
              },
              {
                key: 'from',
                title: (
                  <FontAwesomeIcon icon={faLocationDot} style={{ color: 'rgba(0,0,0, 0.5)' }} />
                ),
                render: (user) => user.country || ''
              },
              {
                title: 'Name',
                key: 'name',
                render: (x) => (
                  <span style={{ textTransform: 'capitalize' }}>
                    {x.photo_url && (
                      <img
                        src={x.photo_url}
                        alt=""
                        style={{ width: '24px', height: '24px', borderRadius: '50%' }}
                        className={CSS.margin_r_s}
                      />
                    )}
                    {x.first_name} {x.last_name}
                  </span>
                )
              },
              {
                title: '',
                key: 'gender',
                render: (x: User) => (
                  <span>
                    {x.gender && (
                      <FontAwesomeIcon
                        icon={x.gender === 'male' ? faMale : faFemale}
                        style={{
                          color:
                            x.gender === 'male' ? 'rgba(178,235,242 ,1)' : 'rgba(248,187,208 ,1)'
                        }}
                      />
                    )}
                  </span>
                )
              },
              {
                title: 'Segments',
                key: 'segments',
                render: (x: User) => {
                  return x.is_authenticated ? (
                    <Tag color="blue">Authenticated</Tag>
                  ) : (
                    <Tag color="default">Anonnymous</Tag>
                  )
                }
              },
              {
                title: 'Signed up',
                key: 'signed_up_at',
                render: (x) => {
                  if (!x.signed_up_at) return '-'
                  return (
                    <span className={CSS.font_size_xs}>
                      <Tooltip
                        title={
                          dayjs(x.signed_up_at)
                            .tz(accountCtx.account?.account.timezone as string)
                            .format('lll') +
                          ' in ' +
                          accountCtx.account?.account.timezone
                        }
                      >
                        {dayjs(x.signed_up_at).fromNow()}
                      </Tooltip>
                    </span>
                  )
                }
              },
              {
                title: 'Created at',
                key: 'created_at',
                render: (x) => (
                  <span className={CSS.font_size_xs}>
                    <Tooltip
                      title={
                        dayjs(x.created_at)
                          .tz(accountCtx.account?.account.timezone as string)
                          .format('lll') +
                        ' in ' +
                        accountCtx.account?.account.timezone
                      }
                    >
                      {dayjs(x.created_at).fromNow()}
                    </Tooltip>
                  </span>
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
                width: 150,
                render: (user: User) => (
                  <div className={CSS.text_right}>
                    <Button.Group>
                      {user.consent_all || user.consent_personalization || user.is_authenticated ? (
                        <Button
                          size="small"
                          type="text"
                          onClick={showUser.bind(null, user.external_id)}
                        >
                          <FontAwesomeIcon icon={faSearch} />
                        </Button>
                      ) : (
                        <Tooltip title="User has not given consent to be tracked" placement="left">
                          <Button size="small" type="text" disabled>
                            <FontAwesomeIcon icon={faSearch} />
                          </Button>
                        </Tooltip>
                      )}

                      {/* <ButtonUserState user={row} accountTimezone={accountCtx.account?.account.timezone as string} /> */}
                    </Button.Group>
                  </div>
                )
              }
            ]}
          />
        </Col>
      </Row>
    </Layout>
  )
}

export default RouteUsers
