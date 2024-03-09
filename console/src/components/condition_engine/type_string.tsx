import { FormInstance, Form, Select, Alert } from 'antd'
import Messages from 'utils/formMessages'
import { Condition, FieldTypeRenderer, IOperator } from './interfaces'
import { OperatorEquals } from './operator_equals'
import { OperatorExists, OperatorNotExists } from './operator_exists'
import { OperatorRegex } from './operator_regex'

export class FieldTypeString implements FieldTypeRenderer {
  type = 'string'

  operators: IOperator[] = [
    new OperatorExists(),
    new OperatorNotExists(),
    new OperatorEquals(),
    new OperatorEquals('not_equals', "doesn't equal"),
    new OperatorRegex()
  ]

  // operator = this.operators[0]

  constructor(overrideType?: string) {
    // override string type for country/language/timezone...
    this.type = overrideType ? overrideType : this.type
  }

  render(condition: Condition) {
    const operator = this.operators.find((x) => x.type === condition.leaf?.string_type?.operator)
    if (!operator)
      return (
        <Alert
          type="error"
          message={'operator not found for: {condition.leaf?.string_type?.operator'}
        />
      )
    return <>{operator.render(condition)}</>
  }

  renderFormItems(condition: Condition, form: FormInstance) {
    return (
      <>
        <Form.Item
          name={['string_type', 'operator']}
          rules={[{ required: true, message: Messages.RequiredField }]}
        >
          <Select
            size="small"
            placeholder="select a value"
            style={{ width: '150px' }}
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
            const operator = this.operators.find(
              (x) => x.type === funcs.getFieldValue(['string_type', 'operator'])
            )
            if (operator) return operator.renderFormItems(condition, form)
          }}
        </Form.Item>
      </>
    )
  }
}
