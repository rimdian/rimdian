import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { QueryFunctionContext, useInfiniteQuery, useQuery } from '@tanstack/react-query'
import { Button, Popover, Space, Spin, Table, Tag, Tooltip } from 'antd'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import {
  DataLog,
  Session,
  Order,
  User,
  Device,
  OrderItem,
  Pageview,
  CartItem,
  ProductItem,
  App,
  CustomEvent
} from 'interfaces'
import { useMemo, useState } from 'react'
import CSS from 'utils/css'
import { paramsToQueryString } from 'utils/searchParams'
import dayjs from 'dayjs'
import FormatCurrency from 'utils/format_currency'
import TableTag from 'components/common/partial_table_tag'
import Property from 'components/common/partial_property'
import Attribute from 'components/common/partial_attribute'
import Block from 'components/common/block'
import UserTimelineSession from './item_session'
import UserTimelineOrder from './item_order'
import UserTimelinePageview from './item_pageview'
import UserTimelineDevice from './item_device'
import UserTimelineOrderItem from './item_order_item'
import UserTimelineCartItem from './item_cart_item'
import UserTimelineCustomEvent from './item_custom_event'

type BlockUserTimelineProps = {
  user: User
  devices: Device[]
  workspaceCtx: CurrentWorkspaceCtxValue
  timezone: string
  limit?: number
}

interface DataLogResult {
  data_logs: DataLog[]
  next_token: string
}

export const BlockUserTimeline = (props: BlockUserTimelineProps) => {
  const queryKey = ['data_log', props.workspaceCtx.workspace.id, props.user.id, props.limit]
  //   const queryClient = useQueryClient()

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery(
    queryKey,
    (ctx: QueryFunctionContext): Promise<DataLogResult> => {
      // console.log('fetching', ctx)
      const params: any = {}
      if (props.user.id) params['user_id'] = props.user.id
      if (props.limit) params['limit'] = props.limit

      let path =
        '/dataLog.list?' +
        paramsToQueryString(params) +
        '&workspace_id=' +
        props.workspaceCtx.workspace.id
      if (ctx.pageParam) {
        path += '&next_token=' + ctx.pageParam
      }
      return new Promise((resolve, reject) => {
        props.workspaceCtx
          .apiGET(path)
          .then((data: any) => {
            //   console.log('data', data)
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

  const sessionIDs = useMemo(() => {
    if (!data || !data.pages.length) return []
    const sessionIDsMap: any = {}
    data.pages.forEach((page) => {
      page.data_logs.forEach((line) => {
        if (line.kind === 'session') {
          sessionIDsMap[line.item_id] = true
        }
      })
    })
    return Object.keys(sessionIDsMap)
  }, [data])

  const pageviewIDs = useMemo(() => {
    if (!data || !data.pages.length) return []
    const pageviewIDsMap: any = {}
    data.pages.forEach((page) => {
      page.data_logs.forEach((line) => {
        if (line.kind === 'pageview') {
          pageviewIDsMap[line.item_id] = true
        }
      })
    })
    return Object.keys(pageviewIDsMap)
  }, [data])

  const customEventIDs = useMemo(() => {
    if (!data || !data.pages.length) return []
    const customEventIDsMap: any = {}
    data.pages.forEach((page) => {
      page.data_logs.forEach((line) => {
        if (line.kind === 'custom_event') {
          customEventIDsMap[line.item_id] = true
        }
      })
    })
    return Object.keys(customEventIDsMap)
  }, [data])

  const orderIDs = useMemo(() => {
    if (!data || !data.pages.length) return []
    const orderIDsMap: any = {}
    data.pages.forEach((page) => {
      page.data_logs.forEach((line) => {
        if (line.kind === 'order') {
          orderIDsMap[line.item_id] = true
        }
      })
    })
    return Object.keys(orderIDsMap)
  }, [data])

  const orderItemIDs = useMemo(() => {
    if (!data || !data.pages.length) return []
    const orderItemIDsMap: any = {}
    data.pages.forEach((page) => {
      page.data_logs.forEach((line) => {
        if (line.kind === 'order_item') {
          orderItemIDsMap[line.item_id] = true
        }
      })
    })
    return Object.keys(orderItemIDsMap)
  }, [data])

  // const cartIDs = useMemo(() => {
  //   if (!data || !data.pages.length) return []
  //   const cartIDsMap: any = {}
  //   data.pages.forEach((page) => {
  //     page.data_logs.forEach((line) => {
  //       if (line.kind === 'cart') {
  //         cartIDsMap[line.item_id] = true
  //       }
  //     })
  //   })
  //   return Object.keys(cartIDsMap)
  // }, [data])

  const cartItemIDs = useMemo(() => {
    if (!data || !data.pages.length) return []
    const cartItemIDsMap: any = {}
    data.pages.forEach((page) => {
      page.data_logs.forEach((line) => {
        if (line.kind === 'cart_item') {
          cartItemIDsMap[line.item_id] = true
        }
      })
    })
    return Object.keys(cartItemIDsMap)
  }, [data])

  const { isLoading: isLoadingOrders, data: orders } = useQuery<Order[]>(
    ['order_key', ...orderIDs],
    (): Promise<Order[]> => {
      return new Promise((resolve, reject) => {
        if (!orderIDs.length) return resolve([])
        const placeholders = orderIDs.map(() => '?').join(',')
        props.workspaceCtx
          .apiPOST('/db.select', {
            workspace_id: props.workspaceCtx.workspace.id,
            from: 'order',
            columns: ['*'],
            where: 'id IN (' + placeholders + ')',
            args: orderIDs
          })
          .then(resolve)
          .catch(reject)
      })
    }
  )

  // const { isLoading: isLoadingCarts, data: carts } = useQuery<Cart[]>(
  //   cartIDs,
  //   (): Promise<Cart[]> => {
  //     return new Promise((resolve, reject) => {
  //       if (!cartIDs.length) return resolve([])
  //       const placeholders = cartIDs.map(() => '?').join(',')
  //       props.workspaceCtx
  //         .apiPOST('/db.select', {
  //           workspace_id: props.workspaceCtx.workspace.id,
  //           from: 'cart',
  //           columns: ['*'],
  //           where: 'id IN (' + placeholders + ')',
  //           args: cartIDs
  //         })
  //         .then(resolve)
  //         .catch(reject)
  //     })
  //   }
  // )

  const { isLoading: isLoadingOrderItems, data: orderItems } = useQuery<OrderItem[]>(
    ['order_item_key', ...orderItemIDs],
    (): Promise<OrderItem[]> => {
      return new Promise((resolve, reject) => {
        if (!orderItemIDs.length) return resolve([])
        const placeholders = orderItemIDs.map(() => '?').join(',')
        props.workspaceCtx
          .apiPOST('/db.select', {
            workspace_id: props.workspaceCtx.workspace.id,
            from: 'order_item',
            columns: ['*'],
            where: 'id IN (' + placeholders + ')',
            args: orderItemIDs
          })
          .then(resolve)
          .catch(reject)
      })
    }
  )

  const { isLoading: isLoadingCartItems, data: cartItems } = useQuery<CartItem[]>(
    ['cart_item_key', ...cartItemIDs],
    (): Promise<CartItem[]> => {
      return new Promise((resolve, reject) => {
        if (!cartItemIDs.length) return resolve([])
        const placeholders = cartItemIDs.map(() => '?').join(',')
        props.workspaceCtx
          .apiPOST('/db.select', {
            workspace_id: props.workspaceCtx.workspace.id,
            from: 'cart_item',
            columns: ['*'],
            where: 'id IN (' + placeholders + ')',
            args: cartItemIDs
          })
          .then(resolve)
          .catch(reject)
      })
    }
  )

  const { isLoading: isLoadingSessions, data: sessions } = useQuery<Session[]>(
    ['session_key', ...sessionIDs],
    (): Promise<Session[]> => {
      return new Promise((resolve, reject) => {
        if (!sessionIDs.length) return resolve([])
        const placeholders = sessionIDs.map(() => '?').join(',')
        props.workspaceCtx
          .apiPOST('/db.select', {
            workspace_id: props.workspaceCtx.workspace.id,
            from: 'session',
            columns: ['*'],
            where: 'id IN (' + placeholders + ')',
            args: sessionIDs
          })
          .then(resolve)
          .catch(reject)
      })
    }
  )

  const { isLoading: isLoadingPageviews, data: pageviews } = useQuery<Pageview[]>(
    ['pageview_key', ...pageviewIDs],
    (): Promise<Pageview[]> => {
      return new Promise((resolve, reject) => {
        if (!pageviewIDs.length) return resolve([])
        const placeholders = pageviewIDs.map(() => '?').join(',')
        props.workspaceCtx
          .apiPOST('/db.select', {
            workspace_id: props.workspaceCtx.workspace.id,
            from: 'pageview',
            columns: ['*'],
            where: 'id IN (' + placeholders + ')',
            args: pageviewIDs
          })
          .then(resolve)
          .catch(reject)
      })
    }
  )

  const { isLoading: isLoadingCustomEvents, data: customEvents } = useQuery<CustomEvent[]>(
    ['custom_event_key', ...customEventIDs],
    (): Promise<CustomEvent[]> => {
      return new Promise((resolve, reject) => {
        if (!customEventIDs.length) return resolve([])
        const placeholders = customEventIDs.map(() => '?').join(',')
        props.workspaceCtx
          .apiPOST('/db.select', {
            workspace_id: props.workspaceCtx.workspace.id,
            from: 'custom_event',
            columns: ['*'],
            where: 'id IN (' + placeholders + ')',
            args: customEventIDs
          })
          .then(resolve)
          .catch(reject)
      })
    }
  )

  const lines = useMemo(() => {
    if (!data || !data.pages.length) return []
    const lines: DataLog[] = []
    data.pages.forEach((page) => {
      page.data_logs.forEach((line) => {
        // skip segment enter events if they are merged with a user
        if (line.kind === 'segment' && line.merged_from_user_external_id) {
          return
        }
        // dont show cart events, only cart_items
        if (line.kind === 'cart') return
        lines.push(line)
      })
    })
    // console.log('lines', lines)
    return lines
  }, [data])

  const RenderItem = (line: DataLog) => {
    if (line.action === 'update') {
      return <UserTimelineItemUpdate line={line} />
    }

    if (line.action === 'noop') {
      return (
        <Space size="large">
          <Tooltip title={<>Ext. ID: {line.item_external_id}</>}>
            <span>
              <TableTag table={line.kind} />
            </span>
          </Tooltip>
          updated without changes
        </Space>
      )
    }

    switch (line.kind) {
      case 'segment':
        // hide segment enter events if they are merged with a user
        return (
          <Space size="large">
            <TableTag table={line.kind} />
            <span>
              {line.action === 'enter' && 'entered'}
              {line.action === 'exit' && 'exited'}
              <Tag
                className={CSS.margin_l_s}
                color={props.workspaceCtx.segmentsMap[line.item_external_id].color}
              >
                {props.workspaceCtx.segmentsMap[line.item_external_id].name}
              </Tag>
            </span>
          </Space>
        )
      case 'session':
        return (
          <UserTimelineSession
            devices={props.devices}
            sessions={sessions || []}
            isLoading={isLoadingSessions}
            workspaceCtx={props.workspaceCtx}
            line={line}
          />
        )
      case 'device':
        return (
          <UserTimelineDevice
            devices={props.devices}
            workspaceCtx={props.workspaceCtx}
            line={line}
          />
        )
      case 'pageview':
        return (
          <UserTimelinePageview
            pageviews={pageviews || []}
            isLoading={isLoadingPageviews}
            workspaceCtx={props.workspaceCtx}
            line={line}
          />
        )
      case 'custom_event':
        return (
          <UserTimelineCustomEvent
            customEvents={customEvents || []}
            isLoading={isLoadingCustomEvents}
            workspaceCtx={props.workspaceCtx}
            line={line}
          />
        )
      case 'order':
        return (
          <UserTimelineOrder
            orders={orders || []}
            isLoading={isLoadingOrders}
            workspaceCtx={props.workspaceCtx}
            line={line}
          />
        )
      case 'order_item':
        return (
          <UserTimelineOrderItem
            orderItems={orderItems || []}
            orders={orders || []}
            isLoading={isLoadingOrderItems}
            workspaceCtx={props.workspaceCtx}
            line={line}
          />
        )
      case 'cart':
        return ''
      // return (
      //   <UserTimelineCart
      //     carts={carts || []}
      //     isLoading={isLoadingCarts}
      //     workspaceCtx={props.workspaceCtx}
      //     line={line}
      //   />
      // )
      case 'cart_item':
        return (
          <UserTimelineCartItem
            cartItems={cartItems || []}
            isLoading={isLoadingCartItems}
            workspaceCtx={props.workspaceCtx}
            line={line}
          />
        )
      case 'user':
        return (
          <Space size="large">
            <TableTag table={line.kind} />
            <span>
              <span className={CSS.font_weight_semibold + ' ' + CSS.padding_r_s}>
                {line.item_external_id}
              </span>{' '}
              is created
            </span>
          </Space>
        )
      case 'user_alias':
        return (
          <Space size="large">
            <TableTag table={line.kind} />
            <span>
              merged with
              <span className={CSS.font_weight_semibold + ' ' + CSS.padding_l_s}>
                {line.updated_fields.find((f: any) => f.field === 'user_external_id')?.previous}
              </span>
            </span>
          </Space>
        )
      default:
        return (
          <div>
            <Space size="large">
              <Tooltip title={<>Ext. ID: {line.item_external_id}</>}>
                <TableTag table={line.kind} />
              </Tooltip>
              is created
            </Space>
          </div>
        )
    }
  }

  return (
    <div>
      <table>
        <tbody>
          {lines.map((line: DataLog, index: number) => {
            return (
              <tr key={line.id} className="tr-hover">
                <td
                  width={150}
                  className={
                    CSS.opacity_70 +
                    ' ' +
                    CSS.padding_v_s +
                    ' ' +
                    CSS.padding_l_s +
                    ' ' +
                    CSS.text_right +
                    ' ' +
                    CSS.padding_r_xl
                  }
                  style={{ position: 'relative' }}
                >
                  <Tooltip
                    title={
                      <span>
                        {dayjs(line.event_at).tz(props.timezone).format('lll')} in {props.timezone}
                      </span>
                    }
                  >
                    {dayjs(line.event_at).fromNow()}
                  </Tooltip>
                </td>
                <td style={{ borderLeft: '1px solid #90A4AE', position: 'relative' }} width={30}>
                  {/* mask the timeline vertical bar for the first row */}
                  {index === 0 && (
                    <div
                      style={{
                        top: 0,
                        left: -1,
                        position: 'absolute',
                        backgroundColor: 'rgb(243, 246, 252)',
                        width: 3
                      }}
                    >
                      &nbsp;
                    </div>
                  )}
                  {/* bullet point */}
                  <div
                    style={{
                      position: 'absolute',
                      marginLeft: '-5px',
                      width: 8,
                      height: 8,
                      borderRadius: '50%',
                      backgroundColor: '#90A4AE',
                      marginTop: '20px'
                    }}
                  ></div>
                </td>
                <td className={CSS.padding_r_l + ' ' + CSS.padding_v_s}>{RenderItem(line)}</td>
              </tr>
            )
          })}
        </tbody>
      </table>

      {hasNextPage && (
        <div className={CSS.text_center}>
          <Button
            loading={isFetchingNextPage}
            size="small"
            type="primary"
            ghost
            onClick={() => fetchNextPage()}
          >
            <Space>
              <FontAwesomeIcon icon={faPlus} />
              Load more
            </Space>
          </Button>
        </div>
      )}
    </div>
  )
}

export const AppColumns = (props: { kind: string; item: any; apps: App[] }) => {
  // loop over workspace apps and extract extra fields
  const appColumns = useMemo(() => {
    const cols: { name: string; iconURL: string }[] = []

    for (const app of props.apps) {
      for (const extraColumn of app.manifest.extra_columns || []) {
        if (extraColumn.kind === props.kind) {
          for (const column of extraColumn.columns || []) {
            // add column only if it has a value
            if (props.item[column.name])
              cols.push({ name: column.name, iconURL: app.manifest.icon_url })
          }
        }
      }
    }

    return cols
  }, [props.apps, props.kind, props.item])

  return (
    <>
      {appColumns.map((extraColumn: any, index: number) => {
        if (index % 2 === 0) {
          return (
            <tr key={index}>
              <td>
                <Property
                  label={
                    <>
                      <img
                        alt=""
                        src={extraColumn.iconURL}
                        height={16}
                        className={CSS.padding_r_s}
                      />{' '}
                      {extraColumn.name}
                    </>
                  }
                >
                  {props.item[extraColumn.name]}
                </Property>
              </td>
              <td>
                <Property
                  label={
                    <>
                      <img
                        alt=""
                        src={appColumns[index + 1].iconURL}
                        height={16}
                        className={CSS.padding_r_s}
                      />{' '}
                      {appColumns[index + 1].name}
                    </>
                  }
                >
                  {props.item[appColumns[index + 1].name]}
                </Property>
              </td>
            </tr>
          )
        }
        return null
      })}
    </>
  )
}

export const Preview = (props: { notFound: boolean; line: DataLog }) => {
  return (
    <Space size="large">
      <Tooltip title={<>Ext. ID: {props.line.item_external_id}</>}>
        <span>
          <TableTag table={props.line.kind} />
        </span>
      </Tooltip>
      {!props.notFound && <Spin size="small" />}
      {props.notFound && <span>(not found)</span>}
    </Space>
  )
}

const UserTimelineItemUpdate = (props: { line: DataLog }) => {
  const [isOpen, setIsOpen] = useState(false)

  const header = (
    <div onClick={() => setIsOpen(!isOpen)} style={{ cursor: 'pointer' }}>
      <span className={CSS.pull_right}>
        {isOpen ? (
          <Button size="small" type="link" onClick={() => setIsOpen(false)}>
            hide -
          </Button>
        ) : (
          <Button size="small" type="link" onClick={() => setIsOpen(true)}>
            details +
          </Button>
        )}
      </span>

      <Space size="large">
        <Tooltip
          title={
            <>
              <p>Data log ID: {props.line.id}</p>
              {props.line.merged_from_user_external_id && (
                <p>Merged from user: {props.line.merged_from_user_external_id}</p>
              )}
              Ext. ID: {props.line.item_external_id}
            </>
          }
        >
          <span>
            <TableTag table={props.line.kind} />
          </span>
        </Tooltip>
        updated
      </Space>
    </div>
  )

  if (!isOpen) return header

  return (
    <>
      {header}
      <Block classNames={[CSS.margin_l_xl, CSS.margin_t_m, CSS.padding_a_m]}>
        {props.line.updated_fields.map((field: any, i: number) => {
          return (
            <span key={i}>
              <Property label={field.field}>
                <Space>
                  {field.previous === '' && '(empty)'}
                  {field.previous === null && '(null)'}
                  {field.previous !== '' && field.previous !== null && field.previous}→
                  {field.new === '' && '(empty)'}
                  {field.new === null && '(null)'}
                  {field.new !== '' && field.new !== null && field.new}
                </Space>
              </Property>
            </span>
          )
        })}
      </Block>
    </>
  )
}

export const ProductsTable = (props: {
  items: ProductItem[]
  currency: string
  workspaceCtx: CurrentWorkspaceCtxValue
  loading?: boolean
}) => {
  return (
    <Table
      size="small"
      dataSource={props.items}
      className={CSS.margin_b_l}
      pagination={false}
      loading={props.loading}
      rowKey="id"
      columns={[
        {
          title: 'Products',
          key: 'product_title',
          render: (_text: string, item: ProductItem) => (
            <Space>
              <span>{item.quantity}x</span>
              {item.image_url && <Popover content={<img alt="" src={item.image_url} />}></Popover>}
              <Popover
                content={
                  <div style={{ width: 400 }}>
                    {item.brand && <Attribute label="Brand">{item.brand}</Attribute>}
                    {item.category && <Attribute label="Category">{item.category}</Attribute>}
                    {item.product_external_id && (
                      <Attribute label="Product ext. ID">{item.product_external_id}</Attribute>
                    )}
                    {item.sku && <Attribute label="SKU">{item.sku}</Attribute>}
                    {item.variant_title && (
                      <Attribute label="Variant title">{item.variant_title}</Attribute>
                    )}
                    {item.variant_external_id && (
                      <Attribute label="Variant ext. ID">{item.variant_external_id}</Attribute>
                    )}
                  </div>
                }
              >
                <Button type="link">
                  {item.name}
                  {item.variant_title && ' - ' + item.variant_title}
                </Button>
              </Popover>
            </Space>
          )
        },
        {
          title: 'Unit price',
          key: 'price',
          render: (_text: string, item: ProductItem) => (
            <Space>{FormatCurrency(item.price, props.currency)}</Space>
          )
        }
      ]}
    />
  )
}
