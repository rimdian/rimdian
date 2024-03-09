import { Form, FormInstance, Input } from 'antd'
import CSS from 'utils/css'
import Messages from 'utils/formMessages'
import { Condition, IOperator, OperatorType } from './interfaces'

export type OperatorRegexProps = {
  value: string | undefined
}

export class OperatorRegex implements IOperator {
  type: OperatorType = 'regex'
  label = 'regex'

  render(condition: Condition) {
    return (
      <>
        regex matches&nbsp;{' '}
        <span className={CSS.text_blue}>{condition.leaf?.string_type?.value}</span>
      </>
    )
  }

  renderFormItems(condition: Condition, _form: FormInstance) {
    return (
      <Form.Item
        name={[condition.leaf?.field_type + '_type', 'value']}
        dependencies={[condition.leaf?.field_type + '_type', 'operator']}
        rules={[{ required: true, type: 'regexp', message: Messages.RequiredField }]}
      >
        <Input size="small" placeholder="i.e: /football/i" style={{ width: '250px' }} />
      </Form.Item>
    )
  }
}
