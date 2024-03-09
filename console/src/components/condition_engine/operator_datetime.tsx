import { FormInstance } from 'antd'
import CSS from 'utils/css'
import { Condition, IOperator, OperatorType } from './interfaces'

export class OperatorBefore implements IOperator {
  type: OperatorType = 'before'
  label = 'before'

  render() {
    return <span className={CSS.text_blue}>{this.label}</span>
  }

  renderFormItems(_condition: Condition, _form: FormInstance) {
    return <>todo</>
  }
}

export class OperatorAfter implements IOperator {
  type: OperatorType = 'after'
  label = 'after'

  render() {
    return <span className={CSS.text_blue}>{this.label}</span>
  }

  renderFormItems(_condition: Condition, _form: FormInstance) {
    return <>todo</>
  }
}

export class OperatorBetween implements IOperator {
  type: OperatorType = 'between'
  label = 'between'

  render() {
    return <span className={CSS.text_blue}>{this.label}</span>
  }

  renderFormItems(_condition: Condition, _form: FormInstance) {
    return <>todo</>
  }
}
