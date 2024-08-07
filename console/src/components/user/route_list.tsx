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
  Space,
  Popover
} from 'antd'
import {
  SubscriptionList,
  SubscriptionListUserActive,
  SubscriptionListUserPaused,
  SubscriptionListUserUnsubscribed,
  User
} from 'interfaces'
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
import { useCallback, useMemo } from 'react'
import { faTrashAlt } from '@fortawesome/free-regular-svg-icons'
import ButtonUpsertSubscriptionList from 'components/subscription_list/button_upsert'
import ButtonImportSubscriptionListUsers from 'components/subscription_list/button_import_users'

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
      position: 'relative',
      cursor: 'pointer',
      padding: CSS.S + ' ' + CSS.L,
      borderLeft: '2px solid rgba(0,0,0,0)',
      '.chevron': {
        position: 'absolute',
        top: 16,
        left: 7,
        fontSize: '10px',
        color: '#4e6cff',
        opacity: 0
      },
      '&:hover': {
        '.chevron': {
          opacity: 1
        }
      }
    },

    '& > li.active': {
      '.chevron': {
        opacity: 1
      }
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
            '/user.list?with_subscription_lists=true&with_segments=true&limit=' +
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

  const nextPage = useCallback(
    (next_token?: string) => {
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
    },
    [searchParams, setSearchParams]
  )

  const previousPage = useCallback(
    (previous_token?: string) => {
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
    },
    [searchParams, setSearchParams]
  )

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
  const showUser = useCallback(
    (userExternalId: string) => {
      searchParams.append('showUser', userExternalId)
      setSearchParams(searchParams)
    },
    [searchParams, setSearchParams]
  )

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
    if (
      searchParams.get('segment_id') === '_all' ||
      (!searchParams.get('segment_id') && !searchParams.get('list_id'))
    ) {
      return workspaceCtx.segmentsMap['_all']
    }

    if (searchParams.get('segment_id'))
      return workspaceCtx.segmentsMap[searchParams.get('segment_id') as string]
    else return undefined
  }, [workspaceCtx.segmentsMap, searchParams])

  const currentList = useMemo(() => {
    if (!searchParams.get('list_id')) {
      return undefined
    }

    return workspaceCtx.subscriptionLists.find(
      (list: SubscriptionList) => list.id === searchParams.get('list_id')
    )
  }, [searchParams, workspaceCtx.subscriptionLists])

  // columns are different if we are showing a segment or a subscription list
  const columns = useMemo(() => {
    const signedUp: any = {
      title: 'Signed up',
      key: 'signed_up_at',
      render: (x: User) => {
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
    }

    const createdAt: any = {
      title: 'Created at',
      key: 'created_at',
      render: (x: User) => (
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
    }

    const actions: any = {
      title: (
        <div className={CSS.text_right}>
          <Button.Group>
            <Button
              size="small"
              disabled={!data?.previous_token}
              className={CSS.pull_right}
              onClick={() => previousPage(data?.previous_token)}
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
              onClick={() => nextPage(data?.next_token)}
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
              <Button size="small" type="text" onClick={() => showUser(user.external_id)}>
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

    const subscription: any = {
      title: 'Subscription',
      key: 'subscription',
      render: (user: User) => {
        const subscription = user.subscription_lists?.find(
          (sub) => sub.subscription_list_id === searchParams.get('list_id')
        )
        if (!subscription) return '-'
        return (
          <Popover
            title={
              <>
                {subscription.status === SubscriptionListUserActive && 'Subscription active'}
                {subscription.status === SubscriptionListUserPaused && 'Subscription paused'}
                {subscription.status === SubscriptionListUserUnsubscribed &&
                  'Subscription cancelled'}
              </>
            }
            content={
              <>
                <p>
                  <b>Created: </b>
                  {dayjs(subscription.created_at).tz(user.timezone).format('lll')} in{' '}
                  {user.timezone}
                </p>
                <p>
                  <b>Updated: </b>
                  {dayjs(subscription.db_updated_at).tz(user.timezone).format('lll')} in{' '}
                  {user.timezone}
                </p>
                {subscription.comment && subscription.comment.length > 0 && (
                  <p>
                    <b>Comment: </b> {subscription.comment}
                  </p>
                )}
              </>
            }
          >
            {subscription.status === SubscriptionListUserActive && <Tag color="green">Active</Tag>}
            {subscription.status === SubscriptionListUserPaused && <Tag color="orange">Paused</Tag>}
            {subscription.status === SubscriptionListUserUnsubscribed && (
              <Tag color="default">Unsubscribed</Tag>
            )}
          </Popover>
        )
      }
    }

    const cols = [
      {
        title: 'External ID',
        key: 'id',
        render: (x: User) => <Tooltip title={'Internal ID: ' + x.id}>{x.external_id}</Tooltip>
      },
      {
        key: 'from',
        title: <FontAwesomeIcon icon={faLocationDot} style={{ color: 'rgba(0,0,0, 0.5)' }} />,
        render: (user: User) => user.country || ''
      },
      {
        title: 'Name',
        key: 'name',
        render: (x: User) => (
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
                  color: x.gender === 'male' ? 'rgba(178,235,242 ,1)' : 'rgba(248,187,208 ,1)'
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
      }
    ]

    if (currentSegment) {
      cols.push(signedUp)
      cols.push(createdAt)
    } else if (currentList) {
      cols.push(subscription)
      cols.push(createdAt)
    }

    cols.push(actions)

    return cols
  }, [
    accountCtx.account,
    data,
    isFetching,
    refetch,
    searchParams,
    showUser,
    currentList,
    currentSegment,
    nextPage,
    previousPage
  ])

  return (
    <Layout currentOrganization={workspaceCtx.organization} currentWorkspaceCtx={workspaceCtx}>
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
                className={currentSegment && currentSegment.id === '_all' ? 'active' : ''}
              >
                <FontAwesomeIcon className="chevron" icon={faChevronRight} />
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
                    className={currentSegment && currentSegment.id === segment.id ? 'active' : ''}
                  >
                    {segment.status === 'building' && (
                      <Tooltip className={CSS.pull_right} title="Building...">
                        <Badge status="processing" />
                      </Tooltip>
                    )}
                    <FontAwesomeIcon className="chevron" icon={faChevronRight} />
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
                      <FontAwesomeIcon className="chevron" icon={faChevronRight} />
                      <Tag color={list.color}>{list.name}</Tag>
                      <span className={secondMenuCSS.counter}>
                        {numbro(list.active_users).format({
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
            {currentSegment && currentSegment.id !== '_all' && (
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
                {currentSegment &&
                  currentSegment.id !== 'anonymous' &&
                  currentSegment.id !== 'authenticated' && (
                    <Space>
                      <Popconfirm
                        title="Do you really want to delete this segment?"
                        okText="Delete"
                        okButtonProps={{ danger: true }}
                        cancelText="No"
                        placement="left"
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

            {currentList && (
              <>
                <Space
                  size="large"
                  style={{ marginTop: 2, lineHeight: '20px', height: '20px', marginLeft: 20 }}
                >
                  <Tag color={currentList.color}>{currentList.name}</Tag>

                  <Tooltip title="Active users">
                    <Badge status="success" text={<>{currentList.active_users}</>} />
                  </Tooltip>
                  <Tooltip title="Paused users">
                    <Badge status="warning" text={<>{currentList.paused_users}</>} />
                  </Tooltip>
                  <Tooltip title="Unsubscribed users">
                    <Badge status="default" text={<>{currentList.unsubscribed_users}</>} />
                  </Tooltip>
                </Space>

                <div className={CSS.topSeparator}></div>
                {currentList && (
                  <Space>
                    <ButtonImportSubscriptionListUsers
                      btnProps={{
                        type: 'primary',
                        ghost: true
                      }}
                      segments={segments}
                      subscriptionList={currentList}
                    />
                    {/* <Popconfirm
                      title="Do you really want to delete this subscription list?"
                      okText="Delete"
                      okButtonProps={{ danger: true }}
                      cancelText="No"
                    >
                      <Button type="text" size="small">
                        <FontAwesomeIcon icon={faTrashAlt} />
                      </Button>
                    </Popconfirm> */}
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
            columns={columns}
          />
        </Col>
      </Row>
    </Layout>
  )
}

export default RouteUsers
