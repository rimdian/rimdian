import { FormInstance } from 'antd'
import CSS from 'utils/css'
import { FieldTypeValue, IOperator, Operator } from './interfaces'

export class OperatorSet implements IOperator {
  type: Operator = 'is_set'
  label = 'is set'

  render() {
    return <span className={CSS.opacity_60 + ' ' + CSS.padding_t_xxs}>{this.label}</span>
  }

  renderFormItems(_fieldType: FieldTypeValue, _fieldName: string, _form: FormInstance) {
    return <></>
  }
}

export class OperatorNotSet implements IOperator {
  type: Operator = 'is_not_set'
  label = 'is not set'

  render() {
    return <span className={CSS.opacity_60 + ' ' + CSS.padding_t_xxs}>{this.label}</span>
  }

  renderFormItems(_fieldType: FieldTypeValue, _fieldName: string, _form: FormInstance) {
    return <></>
  }
}
