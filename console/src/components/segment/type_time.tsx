import { FormInstance, Form, Select, Alert } from 'antd'
import Messages from 'utils/formMessages'
import { DimensionFilter, FieldTypeRenderer, FieldTypeValue, IOperator } from './interfaces'
import { OperatorSet, OperatorNotSet } from './operator_set_not_set'
import {
  OperatorBeforeDate,
  OperatorAfterDate,
  OperatorInDateRange,
  OperatorNotInDateRange
} from './operator_time'

export class FieldTypeTime implements FieldTypeRenderer {
  operators: IOperator[] = [
    new OperatorSet(),
    new OperatorNotSet(),
    new OperatorBeforeDate(),
    new OperatorAfterDate(),
    new OperatorInDateRange(),
    new OperatorNotInDateRange()
  ]

  render(filter: DimensionFilter) {
    const operator = this.operators.find((x) => x.type === filter.operator)
    if (!operator)
      return <Alert type="error" message={'operator not found for: {filter.operator'} />
    return <>{operator.render(filter)}</>
  }

  renderFormItems(fieldType: FieldTypeValue, fieldName: string, form: FormInstance) {
    return (
      <>
        <Form.Item name="operator" rules={[{ required: true, message: Messages.RequiredField }]}>
          <Select
            // size="small"
            placeholder="select a value"
            // style={{ width: '150px' }}
            dropdownMatchSelectWidth={false}
            options={this.operators.map((op: IOperator) => {
              return {
                value: op.type,
                label: op.label
              }
            })}
          />
        </Form.Item>

        <Form.Item noStyle shouldUpdate>
          {(funcs) => {
            const operator = this.operators.find((x) => x.type === funcs.getFieldValue('operator'))
            if (operator) return operator.renderFormItems(fieldType, fieldName, form)
          }}
        </Form.Item>
      </>
    )
  }
}
