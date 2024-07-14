import { useState } from 'react'
import { cloneDeep, get, set } from 'lodash'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPenToSquare, faTrashCan } from '@fortawesome/free-regular-svg-icons'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { Button, Select, Tag, Alert, Popconfirm, Cascader } from 'antd'
import {
  FieldsDictionary,
  Condition,
  ConditionBranch,
  ConditionLeaf,
  FieldDefinition,
  StringType,
  BooleanType,
  NumberType,
  DatetimeType,
  CountryType,
  LanguageType,
  TimezoneType,
  EditingConditionLeaf,
  FieldTypeRendererDictionary
} from './interfaces'
import { LeafForm } from './form_leaf'
import { css } from '@emotion/css'
import CSS from 'utils/css'

export const HasLeaf = (node: Condition): boolean => {
  if (node.kind === 'leaf') return true
  if (!node.branch) return false

  return node.branch.conditions.some((child: Condition) => {
    return HasLeaf(child)
  })
}

export type ConditionInputProps = {
  value?: Condition
  onChange?: (updatedValue: Condition) => void
  fieldsDictionary: FieldsDictionary[]
  fieldTypeRendererDictionary: FieldTypeRendererDictionary
}

const inputCSS: any = {
  self: css({
    '&:not(:first-child)': {
      paddingTop: CSS.M
    },
    '& .ant-form-inline .ant-form-item-with-help': {
      margin: 0
    },
    '& .ant-form-item-explain-error': {
      margin: 0
    }
  }),
  inputGroup: css({
    borderBottom: '1px solid #CFD8DC',
    paddingBottom: CSS.M
  }),
  'color-1': {
    borderColor: '#03A9F4'
  },
  'color-2': {
    borderColor: '#4CAF50'
  },
  'color-3': {
    borderColor: '#E91E63'
  },
  'color-4': {
    borderColor: '#795548'
  },
  'color-5': {
    borderColor: '#FF9800'
  },
  'color-6': {
    borderColor: '#009688'
  },
  'color-7': {
    borderColor: '#9C27B0'
  },
  'color-8': {
    borderColor: '#FFC107'
  },
  'color-9': {
    borderColor: '#607D8B'
  },
  'color-10': {
    borderColor: '#CDDC39'
  },
  'color-11': {
    borderColor: '#3F51B5'
  },
  condition: css({
    marginLeft: '60px',
    position: 'relative',
    '&:not(:last-child)': {
      borderBottom: '1px solid rgba(0, 0, 100, 0.1)'
    }
  }),
  conditionSeparator: css({
    position: 'absolute',
    left: '-35px',
    borderRight: '1px solid #CFD8DC',
    top: 0,
    bottom: 0
  }),
  conditionSeparatorHalf: css({
    bottom: '25px'
  }),
  conditionOperatorAndOr: css(
    {
      position: 'absolute',
      left: '-55px',
      top: '16px',
      height: '24px',
      lineHeight: '24px',
      backgroundColor: '#fafafa',
      border: '1px solid #CFD8DC',
      borderRadius: '3px',
      fontSize: '12px',
      fontWeight: 600,
      textAlign: 'center',
      width: '42px'
    },
    CSS.padding_h_xs
  )
}

export const ConditionInput = (props: ConditionInputProps) => {
  const [editingConditionLeaf, setEditingConditionLeaf] = useState<
    EditingConditionLeaf | undefined
  >(undefined)

  // borderColor incrementer
  let currentColorID = 0
  const getColorID = () => {
    currentColorID++
    return currentColorID
  }

  const cancelOrDeleteCondition = (path: string, key: number) => {
    // console.log('path', path);
    // console.log('key', key);
    const clonedTree = cloneDeep(props.value) as Condition

    // cancel if edit, and is not new
    if (editingConditionLeaf && !editingConditionLeaf.is_new) {
      setEditingConditionLeaf(undefined)
      props.onChange?.(clonedTree)
      return
    }

    // condition is new, and not yet confirmed, we remove it from the tree
    const target = get(clonedTree, path)

    if (target && target.length) {
      set(
        clonedTree,
        path,
        target.filter((_x: any, i: number) => i !== key)
      )
    }

    // reset possible edit mode on current field
    if (
      editingConditionLeaf &&
      editingConditionLeaf.path === path &&
      editingConditionLeaf.key === key
    ) {
      setEditingConditionLeaf(undefined)
    }

    if (props.onChange) props.onChange(clonedTree)
  }

  const addCondition = (path: string, key: number, values: any[], selectedOptions: any) => {
    // console.log('values', values);
    // console.log('selectedOptions', selectedOptions);
    // console.log('path', path);
    // console.log('key', key);
    if (!props.value) return

    const clonedTree = cloneDeep(props.value) as Condition
    if (!clonedTree.branch) return

    const setPath = path + '[' + key + ']'

    // Add branch
    if (values[0] === 'and') {
      const node: Condition = {
        kind: 'branch',
        branch: {
          operator: 'and',
          conditions: []
        } as ConditionBranch
      }

      // node path, if non root
      if (path === '') {
        clonedTree.branch.conditions.push(node)
      } else {
        const target = get(clonedTree, setPath)
        target.branch.conditions.push(node)
      }

      // console.log('tree is', JSON.stringify(clonedTree, undefined, 2))
      props.onChange?.(clonedTree)
      return
    }

    // Add leaf

    const fieldDefinition = selectedOptions[selectedOptions.length - 1] as FieldDefinition

    const leaf = {
      field: fieldDefinition.field,
      field_type: fieldDefinition.type
    } as ConditionLeaf

    switch (fieldDefinition.type) {
      case 'string':
        leaf.string_type = {
          operator: fieldDefinition.defaultOperator
        } as StringType
        break
      case 'boolean':
        leaf.boolean_type = {
          operator: fieldDefinition.defaultOperator
        } as BooleanType
        break
      case 'number':
        leaf.number_type = {
          operator: fieldDefinition.defaultOperator
        } as NumberType
        break
      case 'datetime':
        leaf.datetime_type = {
          operator: fieldDefinition.defaultOperator
        } as DatetimeType
        break
      case 'country':
        leaf.country_type = {
          operator: fieldDefinition.defaultOperator
        } as CountryType
        break
      case 'language':
        leaf.language_type = {
          operator: fieldDefinition.defaultOperator
        } as LanguageType
        break
      case 'timezone':
        leaf.timezone_type = {
          operator: fieldDefinition.defaultOperator
        } as TimezoneType
        break
      default: {
        console.error('operator type ' + fieldDefinition.type + ' is not implemented')
        return
      }
    }

    const node: Condition = {
      kind: 'leaf',
      leaf: leaf
    }

    let editingNodeKey = 0

    // node path, if non root
    if (path === '') {
      clonedTree.branch.conditions.push(node)
      editingNodeKey = clonedTree.branch.conditions.length - 1
    } else {
      const target = get(clonedTree, setPath)
      target.branch.conditions.push(node)
      editingNodeKey = target.branch.conditions.length - 1
    }

    const editingConditionLeaf = Object.assign(
      {
        path: path === '' ? 'branch.conditions' : setPath + '.branch.conditions',
        key: editingNodeKey,
        is_new: true
      },
      node as object
    ) as EditingConditionLeaf

    setEditingConditionLeaf(editingConditionLeaf)

    // console.log('tree is', JSON.stringify(clonedTree, undefined, 2))
    props.onChange?.(clonedTree)
  }

  const deleteButton = (path: string, pathKey: number, isBranch: boolean) => {
    return (
      <Popconfirm
        placement="left"
        title={'Do you really want to remove this ' + (isBranch ? 'branch' : 'condition') + '?'}
        onConfirm={cancelOrDeleteCondition.bind(null, path, pathKey)}
        okText="Delete"
        okButtonProps={{ danger: true }}
        cancelText="Cancel"
      >
        <Button size="small">
          <FontAwesomeIcon icon={faTrashCan} />
        </Button>
      </Popconfirm>
    )
  }

  const editCondition = (path: string, key: number) => {
    if (!props.value) return

    const condition = get(props.value, path + '[' + key + ']')

    const editingConditionLeaf = Object.assign(
      {
        path: path,
        key: key
      },
      condition
    ) as EditingConditionLeaf

    setEditingConditionLeaf(editingConditionLeaf)
  }

  const changeBranchOperator = (path: string, pathKey: number, value: any) => {
    const clonedTree = cloneDeep(props.value) as Condition
    if (!clonedTree.branch) return

    if (path === '') {
      clonedTree.branch.operator = value
    } else {
      set(clonedTree, path + '[' + pathKey + '].branch.operator', value)
    }

    // console.log('new tree', JSON.stringify(clonedTree, undefined, 2))
    props.onChange?.(clonedTree)
  }

  const onUpdateCondition = (updatedCondition: Condition, path: string, pathKey: number) => {
    const fullPath = path + '[' + pathKey + ']'
    // console.log('fullPath', fullPath)
    // const condition = get(props.value, path + '[' + pathKey + ']')

    const clonedTree = cloneDeep(props.value) as Condition
    set(clonedTree, fullPath, updatedCondition)
    console.log('tree is', JSON.stringify(clonedTree, undefined, 2))
    props.onChange?.(clonedTree)
  }

  const renderLeaf = (condition: Condition, path: string, pathKey: number) => {
    const isEditingCurrent =
      editingConditionLeaf &&
      editingConditionLeaf.path === path &&
      editingConditionLeaf.key === pathKey
        ? true
        : false
    let fieldDefinition: FieldDefinition | undefined

    props.fieldsDictionary.forEach((folder) => {
      fieldDefinition = folder.fields.find((def) => def.field === condition.leaf?.field)
    })

    if (!fieldDefinition)
      return (
        <Alert type="error" message={'field definition ' + condition.leaf?.field + ' not found'} />
      )

    const fieldTypeRenderer =
      props.fieldTypeRendererDictionary[condition.leaf?.field_type as string]

    if (isEditingCurrent && editingConditionLeaf) {
      return (
        <div className={css([CSS.padding_v_m, CSS.padding_l_m])}>
          <LeafForm
            value={condition}
            onChange={(updatedLeaf: Condition) => {
              onUpdateCondition(updatedLeaf, path, pathKey)
            }}
            fieldDefinition={fieldDefinition as FieldDefinition}
            fieldTypeRenderer={fieldTypeRenderer}
            editingConditionLeaf={editingConditionLeaf as EditingConditionLeaf}
            setEditingConditionLeaf={setEditingConditionLeaf}
            cancelOrDeleteCondition={cancelOrDeleteCondition.bind(null, path, pathKey)}
          />
        </div>
      )
    }

    // const fieldDefinition = getFieldDefinitionFromCondition(condition)
    // console.log('fieldDefinition', fieldDefinition)
    // console.log('condition', condition)

    return (
      <div className={css([CSS.padding_v_m, CSS.padding_l_m])}>
        <Button.Group className={CSS.pull_right}>
          {deleteButton(path, pathKey, false)}
          <Button size="small" onClick={editCondition.bind(null, path, pathKey)}>
            <FontAwesomeIcon icon={faPenToSquare} />
          </Button>
        </Button.Group>

        <div>
          <b>
            <Tag color="purple">{fieldDefinition.label}</Tag>
            {fieldTypeRenderer.render(condition)}
          </b>
        </div>
      </div>
    )
  }

  const renderBranch = (node: Condition, path: string, pathKey: number) => {
    if (!node.branch) return <span>A branch condition is required...</span>

    const conditionPath =
      path === '' ? 'branch.conditions' : path + '[' + pathKey + '].branch.conditions'
    // console.log('conditionPath', conditionPath)
    const isEditing = editingConditionLeaf ? true : false
    const borderColorID = getColorID()

    return (
      <div className={inputCSS.self}>
        <div className={css([inputCSS.inputGroup, inputCSS['color-' + borderColorID]])}>
          {/* DELETE GROUP BUTTON */}
          {path !== '' && !isEditing && (
            <Button.Group className={CSS.pull_right}>
              {deleteButton(path, pathKey, true)}
            </Button.Group>
          )}
          {/* SELECT GROUP AND OR */}
          <Select
            size="small"
            className={CSS.margin_r_xs}
            style={{ width: '80px' }}
            onChange={changeBranchOperator.bind(null, path, pathKey)}
            value={node.branch.operator}
          >
            <Select.Option value="and">ALL</Select.Option>
            <Select.Option value="or">ANY</Select.Option>
          </Select>{' '}
          <span className={CSS.opacity_60}>of the following conditions match:</span>
        </div>

        {/* LOOP OVER CONDITIONS */}

        {node.branch.conditions.map((cond: Condition, i: number) => {
          return (
            <div key={i} className={inputCSS.condition}>
              <div
                className={css([inputCSS.conditionSeparator, inputCSS['color-' + borderColorID]])}
              ></div>
              {i !== 0 && (
                <div
                  className={css([
                    inputCSS.conditionOperatorAndOr,
                    inputCSS['color-' + borderColorID]
                  ])}
                >
                  {node.branch?.operator}
                </div>
              )}

              {/* recursive call to draw the tree */}
              {cond.leaf && renderLeaf(cond, conditionPath, i)}
              {cond.branch && renderBranch(cond, conditionPath, i)}
            </div>
          )
        })}

        {/* ADD CONDITION BUTTON */}

        <div className={inputCSS.condition}>
          <div
            className={css([
              inputCSS.conditionSeparator,
              inputCSS.conditionSeparatorHalf,
              inputCSS['color-' + borderColorID]
            ])}
          ></div>
          {node.branch.conditions.length > 0 && (
            <div
              className={css([inputCSS.conditionOperatorAndOr, inputCSS['color-' + borderColorID]])}
            >
              {node.branch.operator}
            </div>
          )}

          <div className={CSS.padding_v_m}>
            <Cascader
              defaultValue={undefined}
              value={undefined}
              // popupClassName="cascader-wide"
              onChange={addCondition.bind(null, path, pathKey)}
              expandTrigger="hover"
              fieldNames={
                {
                  children: 'fields'
                } as any
              }
              options={[
                { value: 'and', label: 'AND | OR' }, // AND by default, user can switch to OR after
                // { value: 'or', label: "Group OR", },
                ...props.fieldsDictionary
              ]}
            >
              <Button
                size="small"
                type="primary"
                ghost={node.branch.conditions.length > 0}
                disabled={editingConditionLeaf ? true : false}
              >
                <FontAwesomeIcon icon={faPlus} />
                &nbsp; Condition
              </Button>
            </Cascader>
          </div>
        </div>
      </div>
    )
  }

  if (!props.value) {
    return <span>A value is required...</span>
  }

  return <div className={CSS.padding_t_xs}>{renderBranch(props.value, '', 0)}</div>
}
