import { Button, Popover, Space, Tooltip } from 'antd'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { DataLog, Pageview } from 'interfaces'
import { useState } from 'react'
import CSS from 'utils/css'
import FormatCurrency from 'utils/format_currency'
import TableTag from 'components/common/partial_table_tag'
import Block from 'components/common/block'
import { AppColumns, Preview } from './block_user_timeline'
import Property from 'components/common/partial_property'
import FormatDuration from 'utils/format_duration'
import KeepURLPath from 'utils/keep_url_path'
import { truncate } from 'lodash'

const UserTimelinePageview = (props: {
  workspaceCtx: CurrentWorkspaceCtxValue
  line: DataLog
  pageviews: Pageview[]
  isLoading: boolean
}) => {
  const [isOpen, setIsOpen] = useState(false)

  if (props.isLoading) return Preview({ notFound: false, line: props.line })

  const pageview = props.pageviews.find((p) => p.id === props.line.item_id)
  if (!pageview) return Preview({ notFound: true, line: props.line })

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
        {pageview.duration && FormatDuration(pageview.duration)}
        {pageview.page_id && (
          <Tooltip title={pageview.page_id}>
            <b>{KeepURLPath(pageview.page_id)}</b>
          </Tooltip>
        )}
      </Space>
    </div>
  )

  if (!isOpen) return header

  const properties: any = []

  if (pageview.product_brand)
    properties.push(<Property label="Product brand">{pageview.product_brand}</Property>)
  if (pageview.product_category)
    properties.push(<Property label="Product category">{pageview.product_category}</Property>)
  if (pageview.product_external_id)
    properties.push(<Property label="Product ext. ID">{pageview.product_external_id}</Property>)
  if (pageview.product_name)
    properties.push(<Property label="Product name">{pageview.product_name}</Property>)
  if (pageview.product_price)
    properties.push(
      <Property label="Product price">
        {FormatCurrency(
          pageview.product_price,
          pageview.product_currency || props.workspaceCtx.workspace.currency
        )}
      </Property>
    )
  if (pageview.product_converted_price)
    properties.push(
      <Property label="Converted product price">
        {FormatCurrency(pageview.product_converted_price, props.workspaceCtx.workspace.currency)}
      </Property>
    )
  if (pageview.product_sku)
    properties.push(<Property label="Product SKU">{pageview.product_sku}</Property>)
  if (pageview.product_variant_external_id)
    properties.push(
      <Property label="Product variant ext. ID">{pageview.product_variant_external_id}</Property>
    )
  if (pageview.product_variant_title)
    properties.push(
      <Property label="Product variant title">{pageview.product_variant_title}</Property>
    )

  return (
    <div>
      {header}
      <Block classNames={[CSS.margin_l_xl, CSS.margin_t_m, CSS.padding_a_m]}>
        <table>
          <tbody>
            <tr>
              <td colSpan={2}>
                <Property label="Title">
                  <>{pageview.title}</>
                </Property>
              </td>
            </tr>
            {pageview.referrer && (
              <tr>
                <td colSpan={2}>
                  <Property label="Referrer">
                    <a href={pageview.referrer} rel="noreferrer" target="_blank">
                      {truncate(pageview.referrer, { length: 50 })}
                    </a>
                  </Property>
                </td>
              </tr>
            )}
            {/* split properties in two columns */}
            {properties.map((property: any, index: number) => {
              if (index % 2 === 0) {
                return (
                  <tr key={index}>
                    <td>{property}</td>
                    <td>{properties[index + 1]}</td>
                  </tr>
                )
              }
              return null
            })}
            {pageview.image_url && (
              <tr>
                <td colSpan={2}>
                  <Property label="Image">
                    <Popover content={<img alt="" src={pageview.image_url} />}>
                      <img alt="" src={pageview.image_url} height={50} />
                    </Popover>
                  </Property>
                </td>
              </tr>
            )}
            <AppColumns kind="pageview" item={pageview} apps={props.workspaceCtx.workspace.apps} />
          </tbody>
        </table>
      </Block>
    </div>
  )
}

export default UserTimelinePageview
