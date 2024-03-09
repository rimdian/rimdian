import { FormInstance, Form, Select, Alert } from 'antd'
import Messages from 'utils/formMessages'
import { Condition, FieldTypeRenderer, IOperator } from './interfaces'
import { OperatorExists, OperatorNotExists } from './operator_exists'
import { OperatorBefore, OperatorAfter, OperatorBetween } from './operator_datetime'

export class FieldTypeDatetime implements FieldTypeRenderer {
  type = 'datetime'

  operators: IOperator[] = [
    new OperatorExists(),
    new OperatorNotExists(),
    new OperatorBefore(),
    new OperatorAfter(),
    new OperatorBetween()
  ]

  // operator = this.operators[0]

  constructor(overrideType?: string) {
    // override string type for country/language/timezone...
    this.type = overrideType ? overrideType : this.type
  }

  render(condition: Condition) {
    const operator = this.operators.find((x) => x.type === condition.leaf?.datetime_type?.operator)
    if (!operator)
      return (
        <Alert
          type="error"
          message={'operator not found for: {condition.leaf?.datetime_type?.operator'}
        />
      )
    return <>{operator.render(condition)}</>
  }

  renderFormItems(condition: Condition, form: FormInstance) {
    return (
      <>
        <Form.Item
          name={['datetime_type', 'operator']}
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
              (x) => x.type === funcs.getFieldValue(['datetime_type', 'operator'])
            )
            if (operator) return operator.renderFormItems(condition, form)
          }}
        </Form.Item>
      </>
    )
  }
}
