import { FormInstance } from 'antd'
import CSS from 'utils/css'
import { Condition, IOperator, OperatorType } from './interfaces'

export class OperatorExists implements IOperator {
  type: OperatorType = 'exists'
  label = 'exists'

  render() {
    return <span className={CSS.text_blue}>{this.label}</span>
  }

  renderFormItems(_condition: Condition, _form: FormInstance) {
    return <></>
  }
}

export class OperatorNotExists implements IOperator {
  type: OperatorType = 'not_exists'
  label = "doesn't exist"

  render() {
    return <span className={CSS.text_blue}>{this.label}</span>
  }

  renderFormItems(_condition: Condition, _form: FormInstance) {
    return <></>
  }
}
