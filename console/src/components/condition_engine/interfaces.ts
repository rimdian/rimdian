import { FormInstance } from 'antd'

export type OperatorType =
  | 'and'
  | 'or'
  | 'exists'
  | 'not_exists'
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'not_contains'
  | 'regex'
  | 'is_true'
  | 'is_false'
  | 'before'
  | 'after'
  | 'between'

export type FieldTypeValue =
  | 'string'
  | 'number'
  | 'boolean'
  | 'datetime'
  | 'currency'
  | 'country'
  | 'language'
  | 'timezone'

export interface FieldTypeRenderer {
  type: string
  operators: IOperator[]
  render: (condition: Condition) => JSX.Element
  renderFormItems: (condition: Condition, form: FormInstance) => JSX.Element
}

export interface FieldTypeRendererDictionary {
  [key: string]: FieldTypeRenderer
}

export interface IOperator {
  type: OperatorType
  label: string
  render: (condition: Condition) => JSX.Element
  renderFormItems: (condition: Condition, form: FormInstance) => JSX.Element
  // operators: Operator[]
  // renderForm: (operator: OperatorValue, value: string | undefined, onChange: (newValue: any) => void) => JSX.Element
}

// the Condition(Leaf|Branch) is persisted as JSON tree in DB
export interface Condition {
  kind: 'branch' | 'leaf'
  branch?: ConditionBranch
  leaf?: ConditionLeaf
}

// export const IsBranch = (toBeDetermined: Condition): toBeDetermined is ConditionBranch => {
//     if ((toBeDetermined as ConditionBranch).kind === ) {
//         return true
//     }
//     return false
// }

export interface ConditionBranch {
  operator: 'and' | 'or'
  conditions: Condition[]
}

export interface ConditionLeaf {
  field: string
  field_type: FieldTypeValue
  string_type?: StringType
  boolean_type?: BooleanType
  number_type?: NumberType
  datetime_type?: DatetimeType
  currency_type?: CurrencyType
  country_type?: CountryType
  language_type?: LanguageType
  timezone_type?: TimezoneType
}

// current editing condition
export interface EditingConditionLeaf extends Condition {
  is_new?: boolean // flag to remove node from tree if cancel a new condition without confirm
  path: string
  key: number
}

export interface DatetimeType {
  operator: 'exists' | 'not_exists' | 'before' | 'after' | 'between'
  values?: Date[] // between requires 2 values
  timezone?: string
}

export interface NumberType {
  operator:
    | 'exists'
    | 'not_exists'
    | 'equals'
    | 'not_equals'
    | 'greater_than'
    | 'lower_than'
    | 'between'
  values?: number[] // between requires 2 values
}

export type StringTypeOperator =
  | 'exists'
  | 'notExists'
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'not_contains'
  | 'regex'

export interface StringType {
  operator: StringTypeOperator
  value?: string
}

export type CountryType = StringType
export type CurrencyType = StringType
export type LanguageType = StringType
export type TimezoneType = StringType

export interface BooleanType {
  operator: 'exists' | 'not_exists' | 'is_true' | 'is_false'
  value?: string
}

export interface FieldsDictionary {
  field: string
  label: string
  fields: FieldDefinition[]
}

export interface FieldDefinition {
  field: string
  label: string
  type: FieldTypeValue
  defaultOperator: OperatorType
}

// export interface ConditionLeaf extends Condition {
//     type: FieldType
//     // each type has its own object to hold value(s)
//     typeString?: IFieldTypeString
// }

// export const FieldTypeOperators: any = {
//     string: ['exists', 'notExists', 'equals', 'notEquals', 'contains', 'notContains', 'regex'],
//     number: [], // TODO
//     boolean: ['isTrue', 'isFalse'],
//     datetime: [], // TODO
//     currency: ['exists', 'notExists', 'equals', 'notEquals'],
//     country: ['exists', 'notExists', 'equals', 'notEquals'],
//     language: ['exists', 'notExists', 'equals', 'notEquals'],
//     timezone: ['exists', 'notExists', 'equals', 'notEquals'],
// }

// export interface ConditionBranch extends Condition {
//     operator: 'and' | 'or'
//     conditions: Condition[]
// }

// export interface ConditionLeaf extends Condition {
//     field: string
//     type: FieldTypeValue
//     operator: OperatorValue
//     // each type has its own object to hold value(s)
//     typeString?: IFieldTypeString
// }

// export type ConditionValue = {
//     isLeaf: boolean
//     operator: OperatorValue
// }

// export interface ConditionValueLeaf extends ConditionValue {
//     operator: OperatorValue
//     field?: string
//     value?: string
// }

// export interface ConditionValueBranch extends ConditionValue {
//     operator: 'and' | 'or'
//     conditions: ConditionValue[]
// }

// type FocusedCondition = {
//     isLeaf: boolean
//     field: string
//     value?: string
//     operator: OperatorValue
//     fieldDefinition: LeafFieldDefinition
//     path: string
//     key: number
//     isEditing: boolean
// }

// // generate tree from definition
// export type LeafDefinition = {
//     value: string,
//     label: string,
//     children: LeafFieldDefinition[]
// }

// export type LeafFieldDefinition = {
//     value: string,
//     label: string,
//     fieldType: FieldTypeValue,
//     // operators: OperatorType[]
//     valueFormater?: (value: any) => JSX.Element
// }
