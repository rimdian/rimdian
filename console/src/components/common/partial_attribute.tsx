import { css } from '@emotion/css'
import CSS from 'utils/css'

const attributeCss = {
  self: css({
    borderBottom: 'dashed 1px #CFD8DC',
    fontSize: '12px',
    padding: `${CSS.XS} ${CSS.S}`,

    '& :hover': {
      backgroundColor: '@table-row-hover-bg-custom'
    },

    '& ::before': {
      content: '" "',
      display: 'table'
    },

    '& ::after': {
      clear: 'both',
      content: '" "',
      display: 'table'
    }
  }),

  label: css({
    float: 'left',
    fontWeight: 600,
    color: '#455A64',
    fontSize: '13px',
    marginRight: CSS.M
  })
}
export type AttributeProps = {
  classNames?: string[]
  label: React.ReactNode
  children: React.ReactNode
}

const Attribute = (props: AttributeProps) => {
  return (
    <div className={css([attributeCss.self, ...(props.classNames || [])])}>
      <div className={attributeCss.label}>{props.label}</div>
      <div>{props.children}</div>
    </div>
  )
}

export default Attribute
