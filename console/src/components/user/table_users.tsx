import { Tag, Table, TablePaginationConfig, Tooltip, Button, Popover } from 'antd'
import {
  SubscriptionList,
  SubscriptionListUserActive,
  SubscriptionListUserPaused,
  SubscriptionListUserUnsubscribed,
  User
} from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { FilterValue } from 'antd/lib/table/interface'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faChevronLeft,
  faChevronRight,
  faFemale,
  faLocationDot,
  faMale,
  faRefresh,
  faSearch
} from '@fortawesome/free-solid-svg-icons'
import dayjs from 'dayjs'
import CSS from 'utils/css'
import { blockCss } from 'components/common/block'
import { Segment } from 'components/segment/interfaces'
import { useCallback, useMemo } from 'react'
import FormatCurrency from 'utils/format_currency'

interface UserList {
  users: User[]
  next_token?: string // next page: older rows = created before
  previous_token?: string // previous page: newer rows = created after
}

interface BlockUsersProps {
  timezone: string
  segments: Segment[]
  currentSegment?: Segment
  currentList?: SubscriptionList
  limit: number
}

const BlockUsers = (props: BlockUsersProps) => {
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
              props.limit +
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
                dayjs(x.signed_up_at).tz(props.timezone).format('lll') + ' in ' + props.timezone
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
                .tz(props.timezone as string)
                .format('lll') +
              ' in ' +
              props.timezone
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

            {/* <ButtonUserState user={row} accountTimezone={props.timezone as string} /> */}
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
        title: 'Orders & LTV',
        key: 'orders',
        render: (x: User) => {
          return (
            <span>
              {x.orders_count}
              {' - '}
              {FormatCurrency(x.orders_ltv, workspaceCtx.workspace.currency, { light: true })}
            </span>
          )
        }
      }
    ]

    if (props.currentSegment) {
      cols.push(signedUp)
      cols.push(createdAt)
    } else if (props.currentList) {
      cols.push(subscription)
      cols.push(createdAt)
    }

    cols.push(actions)

    return cols
  }, [
    data,
    isFetching,
    refetch,
    searchParams,
    showUser,
    props.currentList,
    props.currentSegment,
    props.timezone,
    workspaceCtx.workspace.currency,
    nextPage,
    previousPage
  ])

  return (
    <>
      <Table
        pagination={false}
        dataSource={data?.users}
        loading={isLoading}
        onChange={onTableChange}
        rowKey="id"
        className={blockCss.self}
        columns={columns}
        size="middle"
      />
    </>
  )
}

export default BlockUsers
