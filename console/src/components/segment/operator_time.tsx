import { DatePicker, Form, FormInstance, Tag } from 'antd'
import CSS from 'utils/css'
import { DimensionFilter, FieldTypeValue, IOperator, Operator } from './interfaces'
import Messages from 'utils/formMessages'
import dayjs from 'dayjs'

const formItemDatetime = (
  <Form.Item
    name={['string_values', 0]}
    dependencies={['operator']}
    rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
    getValueProps={(value: any) => {
      return { value: value ? dayjs(value) : undefined }
    }}
    getValueFromEvent={(_date: any, dateString: string) => dateString}
  >
    <DatePicker showTime={{ defaultValue: dayjs().startOf('day') }} />
  </Form.Item>
)

const formItemDatetimeRange = (
  <Form.Item
    name="string_values"
    dependencies={['operator']}
    rules={[{ required: true, type: 'array', message: Messages.RequiredField }]}
    getValueProps={(values: any[]) => {
      return {
        value: values?.map((value) => {
          return value ? dayjs(value) : undefined
        })
      }
    }}
    getValueFromEvent={(_date: any, dateStrings: string[]) => dateStrings}
  >
    <DatePicker.RangePicker
      showTime={{
        defaultValue: [dayjs().startOf('day'), dayjs().startOf('day')]
      }}
    />
  </Form.Item>
)

export class OperatorBeforeDate implements IOperator {
  type: Operator = 'before_date'
  label = 'before date'

  render(filter: DimensionFilter) {
    return (
      <>
        <span className={CSS.opacity_60 + ' ' + CSS.padding_t_xxs}>{this.label}</span>
        <span>
          <Tag color="blue">{filter.string_values?.[0]}</Tag>
        </span>
      </>
    )
  }

  renderFormItems(_fieldType: FieldTypeValue, _fieldName: string, _form: FormInstance) {
    return formItemDatetime
  }
}

export class OperatorAfterDate implements IOperator {
  type: Operator = 'after_date'
  label = 'after date'

  render(filter: DimensionFilter) {
    return (
      <>
        <span className={CSS.opacity_60 + ' ' + CSS.padding_t_xxs}>{this.label}</span>
        <span>
          <Tag color="blue">{filter.string_values?.[0]}</Tag>
        </span>
      </>
    )
  }

  renderFormItems(_fieldType: FieldTypeValue, _fieldName: string, _form: FormInstance) {
    return formItemDatetime
  }
}

export class OperatorInDateRange implements IOperator {
  type: Operator = 'in_date_range'
  label = 'in date range'

  render(filter: DimensionFilter) {
    return (
      <>
        <span className={CSS.opacity_60 + ' ' + CSS.padding_t_xxs}>{this.label}</span>
        <span>
          <Tag color="blue">{filter.string_values?.[0]}</Tag>
          &rarr;
          <Tag className={CSS.margin_l_s} color="blue">
            {filter.string_values?.[1]}
          </Tag>
        </span>
      </>
    )
  }

  renderFormItems(_fieldType: FieldTypeValue, _fieldName: string, _form: FormInstance) {
    return formItemDatetimeRange
  }
}

export class OperatorNotInDateRange implements IOperator {
  type: Operator = 'not_in_date_range'
  label = 'not in date range'

  render(filter: DimensionFilter) {
    return (
      <>
        <span className={CSS.opacity_60 + ' ' + CSS.padding_t_xxs}>{this.label}</span>
        <span>
          <Tag color="blue">{filter.string_values?.[0]}</Tag>
          &rarr;
          <Tag className={CSS.margin_l_s} color="blue">
            {filter.string_values?.[1]}
          </Tag>
        </span>
      </>
    )
  }

  renderFormItems(_fieldType: FieldTypeValue, _fieldName: string, _form: FormInstance) {
    return formItemDatetimeRange
  }
}
