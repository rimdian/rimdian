import { css, CSSInterpolation } from '@emotion/css'
import CSS, { baseBorderRadius, borderColorSecondary } from 'utils/css'

export type BlockProps = {
  key?: string
  title?: React.ReactNode
  extra?: React.ReactNode
  small?: boolean
  grid?: boolean
  classNames?: CSSInterpolation[]
  style?: React.CSSProperties
  children: React.ReactNode
}

export const blockCss = {
  self: css({
    borderRadius: baseBorderRadius,
    backgroundColor: 'white',
    boxShadow: '1px 1px 1px 0px rgba(0, 0, 0, 0.1)',
    marginBottom: CSS.M
  }),

  head: css(
    {
      display: 'flex',
      justifyContent: 'center',
      flexDirection: 'column',
      minHeight: 56,
      fontSize: 16,
      borderBottom: '1px solid ' + borderColorSecondary,
      '& ::before': {
        content: '""',
        display: 'table'
      },
      '& ::after': {
        content: '""',
        display: 'table',
        clear: 'both'
      }
    },
    CSS.padding_h_m
  ),

  headSmall: css(
    {
      fontSize: 14,
      minHeight: 48
    },
    CSS.padding_h_s
  ),

  headWrapper: css({
    width: '100%',
    display: 'flex',
    justifyContent: 'center'
  }),

  title: css({
    display: 'inline-block',
    flex: 1,
    fontWeight: 600,
    lineHeight: '32px',
    textOverflow: 'ellipsis',
    overflow: 'hidden',
    whiteSpace: 'nowrap'
  }),

  extra: css({
    fontSize: 14,
    lineHeight: '22px',
    marginInlineStart: 'auto',
    fontWeight: 400,
    color: 'inherit'
  }),

  extraSmall: css({
    fontSize: 14,
    lineHeight: '18px',
    marginInlineStart: 'auto',
    fontWeight: 400,
    color: 'inherit'
  })
}

const Block = (props: BlockProps) => {
  const hasHeader = props.title || props.extra

  return (
    <div className={css([blockCss.self, ...(props.classNames || [])])} style={props.style}>
      {hasHeader && (
        <div className={css([blockCss.head, props.small && blockCss.headSmall])}>
          <div className={blockCss.headWrapper}>
            {props.title && <div className={blockCss.title}>{props.title}</div>}
            {props.extra && (
              <div className={css([blockCss.extra, props.small && blockCss.extraSmall])}>
                {props.extra}
              </div>
            )}
          </div>
        </div>
      )}
      <div className={props.grid ? CSS.grid : undefined}>{props.children}</div>
    </div>
  )
}

export default Block
