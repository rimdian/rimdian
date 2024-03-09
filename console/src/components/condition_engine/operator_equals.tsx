import { Alert, Form, FormInstance, Input, Select } from 'antd'
import { Rule } from 'antd/lib/form'
import Messages from 'utils/formMessages'
import { Condition, IOperator, OperatorType } from './interfaces'
import { Currencies, Currency } from 'utils/currencies'
import CSS from 'utils/css'
import { Languages } from 'utils/languages'

export type OperatorEqualsProps = {
  value: string | undefined
}

export class OperatorEquals implements IOperator {
  type: OperatorType = 'equals'
  label = 'equals'

  constructor(overrideType?: OperatorType, overrideLabel?: string) {
    if (overrideType) this.type = overrideType
    if (overrideLabel) this.label = overrideLabel
  }

  render(condition: Condition) {
    let value: any
    switch (condition.leaf?.field_type) {
      case 'string':
        value = condition.leaf?.string_type?.value
        break
      case 'number':
        value = condition.leaf?.number_type?.values?.[0]
        break
      // case 'datetime':
      //     value = condition.leaf?.datetime_type?.value
      // break
      case 'currency':
        value = condition.leaf?.currency_type?.value
        break
      case 'country':
        value = condition.leaf?.country_type?.value
        break
      case 'language':
        value = condition.leaf?.language_type?.value
        break
      case 'timezone':
        value = condition.leaf?.timezone_type?.value
        break
      default:
        value = (
          <Alert
            type="error"
            message={'equals operator not implemented for type: ' + condition.leaf?.field_type}
          />
        )
    }
    return (
      <>
        {this.label}&nbsp; <span className={CSS.text_blue}>{value}</span>
      </>
    )
  }

  renderFormItems(condition: Condition, _form: FormInstance) {
    let fieldName = 'value'
    let rule: Rule = { required: true, type: 'string', message: Messages.RequiredField }
    let input = <Input size="small" placeholder="enter a value" style={{ width: '150px' }} />

    switch (condition.leaf?.field_type) {
      case 'string':
        break
      case 'currency':
        input = (
          <Select
            showSearch
            placeholder="Select a currency"
            optionFilterProp="children"
            filterOption={(input: any, option: any) =>
              option.value.toLowerCase().includes(input.toLowerCase())
            }
            options={Currencies.map((c: Currency) => {
              return { value: c.code, label: c.code + ' - ' + c.currency }
            })}
          />
        )
        break
      case 'country':
        // TODO input
        break
      case 'language':
        input = (
          <Select
            size="small"
            placeholder="Select a value"
            allowClear={false}
            showSearch={true}
            filterOption={(searchText: any, option: any) => {
              return (
                searchText !== '' && option.name.toLowerCase().includes(searchText.toLowerCase())
              )
            }}
            options={Languages}
          />
        )
        break
      case 'timezone':
        // TODO input
        break
      case 'number':
        fieldName = 'values[0]'
        // TODO input
        break
      case 'datetime':
        // TODO input
        break
      default:
        return (
          <Alert
            type="error"
            message={'equals form item not implemented for type: ' + condition.leaf?.field_type}
          />
        )
    }

    return (
      <Form.Item
        name={[condition.leaf?.field_type + '_type', fieldName]}
        dependencies={[condition.leaf?.field_type + '_type', 'operator']}
        rules={[rule]}
      >
        {input}
      </Form.Item>
    )
  }
}
