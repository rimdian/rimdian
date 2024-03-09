import { Button, Space, Tooltip } from 'antd'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { DataLog, Order, OrderItem } from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'
import FormatCurrency from 'utils/format_currency'
import TableTag from 'components/common/partial_table_tag'
import Block from 'components/common/block'
import { Preview, ProductsTable } from './block_user_timeline'
import Property from 'components/common/partial_property'

const UserTimelineOrderItem = (props: {
  orders: Order[]
  orderItems: OrderItem[]
  isLoading: boolean
  workspaceCtx: CurrentWorkspaceCtxValue
  line: DataLog
}) => {
  //   console.log('data', data)
  const [isOpen, setIsOpen] = useState(false)

  if (props.isLoading) return Preview({ notFound: false, line: props.line })
  const orderItem = props.orderItems.find((o) => o.id === props.line.item_id)
  if (!orderItem) return Preview({ notFound: true, line: props.line })

  const order = props.orders.find((o) => o.id === orderItem.order_id)

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
        <b>
          <Tooltip title="Total price (converted in workspace currency)">
            {FormatCurrency(orderItem.converted_price, props.workspaceCtx.workspace.currency)}
          </Tooltip>
        </b>
        <span>
          {orderItem.quantity}x&nbsp;
          <b>{orderItem.name}</b>
        </span>
      </Space>
    </div>
  )

  if (!isOpen) return header

  return (
    <div>
      {header}
      <Block classNames={[CSS.margin_l_l, CSS.margin_t_m, CSS.padding_a_m]}>
        <Property label="Order">{order ? order.external_id : orderItem.order_id}</Property>
        <ProductsTable
          items={[orderItem]}
          currency={orderItem.currency}
          workspaceCtx={props.workspaceCtx}
        />
      </Block>
    </div>
  )
}

export default UserTimelineOrderItem
