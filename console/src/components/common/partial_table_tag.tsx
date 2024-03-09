import { Tag } from 'antd'

export interface TableTagProps {
  table: string
}
const TableTag = (props: TableTagProps) => {
  // magenta red volcano orange gold lime green cyan blue geekblue purple
  const table = props.table.toLowerCase()
  let color = 'geekblue'

  if (table === 'user') color = 'lime'
  if (table === 'session') color = 'magenta'
  if (table === 'order') color = 'purple'
  if (table === 'custom_event') color = 'volcano'
  if (table === 'pageview') color = 'cyan'

  return (
    <Tag style={{ margin: 0 }} color={color}>
      {props.table}
    </Tag>
  )
}

export default TableTag
