import { Form, FormInstance, InputNumber, Tag } from 'antd'
import Messages from 'utils/formMessages'
import { DimensionFilter, FieldTypeValue, IOperator, Operator } from './interfaces'
import CSS from 'utils/css'

export type OperatorNumberProps = {
  value: string | undefined
}

export class OperatorNumber implements IOperator {
  type: Operator = 'gt'
  label = 'greater than'

  constructor(overrideType?: Operator, overrideLabel?: string) {
    if (overrideType) this.type = overrideType
    if (overrideLabel) this.label = overrideLabel
  }

  render(filter: DimensionFilter) {
    return (
      <>
        <span className={CSS.opacity_60 + ' ' + CSS.padding_t_xxs}>{this.label}</span>
        <span>
          <Tag color="blue">{filter.number_values?.[0]}</Tag>
        </span>
      </>
    )
  }

  renderFormItems(_fieldType: FieldTypeValue, _fieldName: string, _form: FormInstance) {
    return (
      <Form.Item
        name={['number_values', 0]}
        dependencies={['operator']}
        rules={[{ required: true, type: 'number', message: Messages.RequiredField }]}
      >
        <InputNumber style={{ width: '100%' }} placeholder="enter a value" />
      </Form.Item>
    )
  }
}
