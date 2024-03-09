import { Alert, Form, FormInstance, Select, Tag } from 'antd'
import { Rule } from 'antd/lib/form'
import Messages from 'utils/formMessages'
import { DimensionFilter, FieldTypeValue, IOperator, Operator } from './interfaces'
import { Currencies, Currency } from 'utils/currencies'
import CSS from 'utils/css'
import { CountriesFormOptions, Timezones } from 'utils/countries_timezones'
import { Languages } from 'utils/languages'

export type OperatorContainsProps = {
  value: string | undefined
}

export class OperatorContains implements IOperator {
  type: Operator = 'contains'
  label = 'contains'

  constructor(overrideType?: Operator, overrideLabel?: string) {
    if (overrideType) this.type = overrideType
    if (overrideLabel) this.label = overrideLabel
  }

  render(filter: DimensionFilter) {
    const values = filter.string_values || []
    return (
      <>
        <span className={CSS.opacity_60 + ' ' + CSS.padding_t_xxs}>{this.label}</span>
        <span>
          {values.map((value, i) => {
            return (
              <>
                <Tag color="blue" key={value}>
                  {value}
                </Tag>
                {i < values.length - 1 && <span className={CSS.padding_r_xs}>or</span>}
              </>
            )
          }) || 'no values'}
        </span>
      </>
    )
  }

  renderFormItems(fieldType: FieldTypeValue, fieldName: string, _form: FormInstance) {
    let rule: Rule = { required: true, type: 'array', min: 1, message: Messages.RequiredField }
    let input = <Select mode="tags" placeholder="press enter to add a value" />

    switch (fieldType) {
      case 'string':
        if (fieldName === 'gender') {
          input = (
            <Select
              // size="small"
              showSearch
              mode="multiple"
              placeholder="Select a gender"
              optionFilterProp="children"
              filterOption={(input: any, option: any) =>
                option.value.toLowerCase().includes(input.toLowerCase())
              }
              options={[
                { value: 'male', label: 'Male' },
                { value: 'female', label: 'Female' }
              ]}
            />
          )
        }
        if (fieldName === 'currency') {
          input = (
            <Select
              // size="small"
              showSearch
              mode="multiple"
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
        }
        if (fieldName === 'country') {
          input = (
            <Select
              // size="small"
              mode="multiple"
              // style={{ width: '200px' }}
              showSearch
              placeholder="Select a country"
              filterOption={(input: any, option: any) =>
                option.label.toLowerCase().includes(input.toLowerCase())
              }
              options={CountriesFormOptions}
            />
          )
        }
        if (fieldName === 'language') {
          input = (
            <Select
              // size="small"
              mode="multiple"
              placeholder="Select a value"
              // style={{ width: '200px' }}
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
        }
        if (fieldName === 'timezone') {
          input = (
            <Select
              // size="small"
              mode="multiple"
              // style={{ width: '200px' }}
              placeholder="Select a time zone"
              allowClear={false}
              showSearch={true}
              filterOption={(searchText: any, option: any) => {
                return (
                  searchText !== '' && option.name.toLowerCase().includes(searchText.toLowerCase())
                )
              }}
              options={Timezones}
              fieldNames={{
                label: 'name',
                value: 'name'
              }}
            />
          )
        }
        break
      case 'number':
        // TODO input
        break
      case 'time':
        // TODO input
        break
      default:
        return (
          <Alert
            type="error"
            message={'contains form item not implemented for type: ' + fieldType}
          />
        )
    }

    return (
      <Form.Item name="string_values" dependencies={['operator']} rules={[rule]}>
        {input}
      </Form.Item>
    )
  }
}
