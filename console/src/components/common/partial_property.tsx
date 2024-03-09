import { Space } from 'antd'
import CSS from 'utils/css'

interface PropertyProps {
  label: React.ReactNode
  children: React.ReactNode
}

const Property = (props: PropertyProps) => {
  return (
    <div className={CSS.padding_b_xs}>
      <Space>
        <span className={CSS.font_weight_semibold}>{props.label}:</span>
        {props.children}
      </Space>
    </div>
  )
}

export default Property
