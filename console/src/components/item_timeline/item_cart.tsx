import { Button, Popover, Space, Table, Tooltip } from 'antd'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { Cart, CartItem, DataLog } from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'
import FormatCurrency from 'utils/format_currency'
import TableTag from 'components/common/partial_table_tag'
import Block from 'components/common/block'
import { Preview } from './block_user_timeline'
import { useQuery } from '@tanstack/react-query'
import Attribute from 'components/common/partial_attribute'

const UserTimelineCart = (props: {
  carts: Cart[]
  isLoading: boolean
  workspaceCtx: CurrentWorkspaceCtxValue
  line: DataLog
}) => {
  const [isOpen, setIsOpen] = useState(false)

  if (props.isLoading) return Preview({ notFound: false, line: props.line })
  const cart = props.carts.find((o) => o.id === props.line.item_id)
  if (!cart) return Preview({ notFound: true, line: props.line })

  const items = cart.items ? JSON.parse(cart.items) : []
  const total = items.reduce((acc: number, item: CartItem) => acc + item.price * item.quantity, 0)
  const quantity = items.reduce((acc: number, item: CartItem) => acc + item.quantity, 0)

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
        {quantity > 0 && <span>{quantity} items</span>}
        {total > 0 && (
          <b>{FormatCurrency(total, cart.currency || props.workspaceCtx.workspace.currency)}</b>
        )}
      </Space>
    </div>
  )

  if (!isOpen) return header

  return (
    <div>
      {header}
      <Block classNames={[CSS.margin_l_l, CSS.margin_t_m, CSS.padding_a_m]}>
        <ItemsTable cart={cart} workspaceCtx={props.workspaceCtx} />
      </Block>
    </div>
  )
}

export default UserTimelineCart

const ItemsTable = (props: { cart: Cart; workspaceCtx: CurrentWorkspaceCtxValue }) => {
  const { isLoading, data } = useQuery<CartItem[]>(
    ['cart_items', props.cart.id],
    (): Promise<CartItem[]> => {
      return new Promise((resolve, reject) => {
        props.workspaceCtx
          .apiPOST('/db.select', {
            workspace_id: props.workspaceCtx.workspace.id,
            from: 'cart_item',
            columns: ['*'],
            where: 'cart_id = ?',
            args: [props.cart.id]
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
          render: (_text: string, item: CartItem) => (
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
          render: (_text: string, item: CartItem) => (
            <Space>{FormatCurrency(item.price, props.cart.currency)}</Space>
          )
        }
      ]}
    />
  )
}
