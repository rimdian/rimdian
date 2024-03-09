import { Button, Popover, Space, Table, Tag, Tooltip } from 'antd'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { DataLog, Order, OrderItem } from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'
import FormatCurrency from 'utils/format_currency'
import TableTag from 'components/common/partial_table_tag'
import Block from 'components/common/block'
import { AppColumns, Preview } from './block_user_timeline'
import dayjs from 'dayjs'
import Property from 'components/common/partial_property'
import FormatDuration from 'utils/format_duration'
import { useQuery } from '@tanstack/react-query'
import Attribute from 'components/common/partial_attribute'

const UserTimelineOrder = (props: {
  orders: Order[]
  isLoading: boolean
  workspaceCtx: CurrentWorkspaceCtxValue
  line: DataLog
}) => {
  //   console.log('data', data)
  const [isOpen, setIsOpen] = useState(false)

  if (props.isLoading) return Preview({ notFound: false, line: props.line })
  const order = props.orders.find((o) => o.id === props.line.item_id)
  if (!order) return Preview({ notFound: true, line: props.line })
  const domain = props.workspaceCtx.workspace.domains.find((d) => d.id === order.domain_id)

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
          <>
            <TableTag table={props.line.kind} />
          </>
        </Tooltip>
        {domain && domain.type !== 'web' && (
          <>
            <Tooltip title={domain.name}>{domain.type}</Tooltip>
          </>
        )}
        <b>
          <Tooltip title="Total price (converted in workspace currency)">
            {FormatCurrency(order.total_price, props.workspaceCtx.workspace.currency)}
          </Tooltip>
        </b>
      </Space>
    </div>
  )

  if (!isOpen) return header
  const discountCodes = order.discount_codes ? JSON.parse(order.discount_codes) : []
  // const items = order.items ? JSON.parse(order.items) : []

  return (
    <div>
      {header}
      <Block classNames={[CSS.margin_l_l, CSS.margin_t_m, CSS.padding_a_m]}>
        <ItemsTable order={order} workspaceCtx={props.workspaceCtx} />

        <table>
          <tbody>
            <tr>
              <td colSpan={2}>
                <Property label="Ext. ID">{order.external_id}</Property>
              </td>
            </tr>

            {order.currency === props.workspaceCtx.workspace.currency && (
              <>
                <tr>
                  <td>
                    <Property label="Subtotal price">
                      {FormatCurrency(
                        order.converted_subtotal_price,
                        props.workspaceCtx.workspace.currency
                      )}
                    </Property>
                  </td>
                  <td>
                    <Property label="Total price">
                      {FormatCurrency(
                        order.converted_total_price,
                        props.workspaceCtx.workspace.currency
                      )}
                    </Property>
                  </td>
                </tr>
              </>
            )}
            {order.currency !== props.workspaceCtx.workspace.currency && (
              <>
                <tr>
                  <td>
                    <Property label="Converted subtotal price">
                      {FormatCurrency(
                        order.converted_subtotal_price,
                        props.workspaceCtx.workspace.currency
                      )}
                    </Property>
                  </td>
                  <td>
                    <Property label="Converted total price">
                      {FormatCurrency(
                        order.converted_total_price,
                        props.workspaceCtx.workspace.currency
                      )}
                    </Property>
                  </td>
                </tr>
                <tr>
                  <td>
                    <Property label="Subtotal price">
                      {FormatCurrency(order.subtotal_price, props.workspaceCtx.workspace.currency)}
                    </Property>
                  </td>
                  <td>
                    <Property label="Total price">
                      {FormatCurrency(order.total_price, props.workspaceCtx.workspace.currency)}
                    </Property>
                  </td>
                </tr>
                <tr>
                  <td>
                    <Property label="Order currency">{order.currency}</Property>
                  </td>
                  <td>
                    <Property label="FX rate">
                      {order.fx_rate ? order.fx_rate.toFixed(4) : 'n/a'}
                    </Property>
                  </td>
                </tr>
              </>
            )}
            <tr>
              <td>
                <Property label="#1 order?">{order.is_first_conversion ? 'yes' : 'no'}</Property>
              </td>
              <td>
                <Property label="Time to conversion">
                  {FormatDuration(order.time_to_conversion)}
                </Property>
              </td>
            </tr>
            {discountCodes.length > 0 && (
              <tr>
                <td colSpan={2}>
                  <Property label="Discount codes">
                    {discountCodes.map((code: string) => (
                      <Tag key={code} color="magenta">
                        {code}
                      </Tag>
                    ))}
                  </Property>
                </td>
              </tr>
            )}
            {order.cancelled_at && (
              <tr>
                <td>
                  <Property label="Cancelled at">
                    {dayjs(order.cancelled_at).format('lll')}
                  </Property>
                </td>
                <td>
                  <Property label="Cancel reason">{order.cancel_reason || 'n/a'}</Property>
                </td>
              </tr>
            )}
            <AppColumns kind="order" item={order} apps={props.workspaceCtx.workspace.apps} />
          </tbody>
        </table>
      </Block>
    </div>
  )
}

export default UserTimelineOrder

const ItemsTable = (props: { order: Order; workspaceCtx: CurrentWorkspaceCtxValue }) => {
  const { isLoading, data } = useQuery<OrderItem[]>(
    ['order_items', props.order.id],
    (): Promise<OrderItem[]> => {
      return new Promise((resolve, reject) => {
        props.workspaceCtx
          .apiPOST('/db.select', {
            workspace_id: props.workspaceCtx.workspace.id,
            from: 'order_item',
            columns: ['*'],
            where: 'order_id = ?',
            args: [props.order.id]
          })
          .then(resolve)
          .catch(reject)
      })
    }
  )

  return (
    <Table
      size="small"
      dataSource={data || []}
      className={CSS.margin_b_l}
      pagination={false}
      loading={isLoading}
      rowKey="id"
      columns={[
        {
          title: 'Products',
          key: 'product_title',
          render: (_text: string, item: OrderItem) => (
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
          render: (_text: string, item: OrderItem) => (
            <Space>{FormatCurrency(item.price, props.order.currency)}</Space>
          )
        }
      ]}
    />
  )
}
