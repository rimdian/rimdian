import { Button, Space, Tooltip } from 'antd'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { DataLog, CartItem } from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'
import FormatCurrency from 'utils/format_currency'
import TableTag from 'components/common/partial_table_tag'
import Block from 'components/common/block'
import { Preview, ProductsTable } from './block_user_timeline'

const UserTimelineCartItem = (props: {
  //   carts: Cart[]
  cartItems: CartItem[]
  isLoading: boolean
  workspaceCtx: CurrentWorkspaceCtxValue
  line: DataLog
}) => {
  //   console.log('data', data)
  const [isOpen, setIsOpen] = useState(false)

  if (props.isLoading) return Preview({ notFound: false, line: props.line })
  const cartItem = props.cartItems.find((o) => o.id === props.line.item_id)
  if (!cartItem) return Preview({ notFound: true, line: props.line })

  //   const cart = props.carts.find((o) => o.id === cartItem.cart_id)

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
            {FormatCurrency(cartItem.converted_price, props.workspaceCtx.workspace.currency)}
          </Tooltip>
        </b>
        <span>
          {cartItem.quantity}x&nbsp;
          <b>{cartItem.name}</b>
        </span>
      </Space>
    </div>
  )

  if (!isOpen) return header

  return (
    <div>
      {header}
      <Block classNames={[CSS.margin_l_l, CSS.margin_t_m, CSS.padding_a_m]}>
        {/* <Property label="Cart">{cart ? cart.external_id : cartItem.cart_id}</Property> */}
        <ProductsTable
          items={[cartItem]}
          currency={cartItem.currency}
          workspaceCtx={props.workspaceCtx}
        />
      </Block>
    </div>
  )
}

export default UserTimelineCartItem
