import { css } from '@emotion/css'
import { Space, Row, Col } from 'antd'
import CSS, { baseBorderRadius, borderColorBase, colorPrimary } from 'utils/css'

type InfoRadioItem = {
  key: string
  icon?: JSX.Element
  title: JSX.Element
  content: JSX.Element
  disabled?: boolean
}
type InfoRadioGroupProps = {
  options: InfoRadioItem[]
  layout: 'horizontal' | 'vertical'
  onChange: (selectedKey: string) => void
  value?: string
  span?: undefined | number | string
  disabled?: boolean
}

const radioCSS = {
  item: css({
    border: '1px solid ' + borderColorBase,
    borderRadius: baseBorderRadius,
    padding: '16px 24px 24px 16px',
    backgroundColor: 'white',
    // color: @radio-button-color,
    height: '100%',

    '&:hover': {
      cursor: 'pointer',
      borderColor: colorPrimary
    },

    '&.disabled.checked': {
      backgroundColor: '#f5f5f5',
      color: '#bfbfbf'
    }
  }),

  checked: css({
    color: colorPrimary,
    backgroundColor: 'white',
    borderColor: colorPrimary
  }),

  disabled: css({
    cursor: 'not-allowed !important',
    backgroundColor: '#f5f5f5',

    '&:hover': {
      borderColor: borderColorBase + ' !important'
    }
  }),

  icon: css({
    fontSize: '24px',
    opacity: 0.6
  }),

  title: css({
    fontWeight: 600,
    fontSize: '14px',
    paddingBottom: '12px'
  }),

  content: css({
    opacity: 0.7,
    fontSize: '13px'
  })
}

const InfoRadioGroup = (props: InfoRadioGroupProps) => {
  // console.log('value', value)
  const onClick = (selectedKey: string) => {
    const item = props.options.find((x) => x.key === selectedKey)
    if (props.disabled || item?.disabled) return
    props.onChange(selectedKey)
  }

  return (
    <Row gutter={24}>
      {props.options.map((item: InfoRadioItem) => (
        <Col
          className={props.layout === 'horizontal' ? CSS.margin_b_m : ''}
          span={props.layout === 'horizontal' ? 24 : props.span}
          key={item.key}
        >
          <div
            onClick={onClick.bind(null, item.key)}
            className={css([
              radioCSS.item,
              props.value === item.key && radioCSS.checked,
              (props.disabled || item.disabled) && radioCSS.disabled
            ])}
          >
            <Space size={24} align="start">
              {item.icon && <div className={radioCSS.icon}>{item.icon}</div>}
              <div>
                <div className={radioCSS.title}>{item.title}</div>
                <div className={radioCSS.content}>{item.content}</div>
              </div>
            </Space>
          </div>
        </Col>
      ))}
    </Row>
  )
}

export default InfoRadioGroup
