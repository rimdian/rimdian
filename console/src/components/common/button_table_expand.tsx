export interface ButtonExpandProps {
  depth: number
  onClick: (e: any) => void
  expanded: boolean
}
export const ButtonExpand = (props: ButtonExpandProps) => {
  return (
    <button
      style={{ marginLeft: props.depth > 2 ? props.depth * 20 - 40 : 0 }}
      type="button"
      className={
        'ant-table-row-expand-icon ant-table-row-expand-icon-' +
        (props.expanded ? 'expanded' : 'collapsed')
      }
      onClick={props.onClick}
    ></button>
  )
}
