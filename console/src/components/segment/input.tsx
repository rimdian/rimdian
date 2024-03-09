import { useMemo, useState } from 'react'
import { cloneDeep, forEach, get, set } from 'lodash'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPenToSquare, faTrashCan } from '@fortawesome/free-regular-svg-icons'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { Button, Select, Tag, Alert, Popconfirm, Cascader, Space, Popover } from 'antd'
import {
  EditingNodeLeaf,
  FieldTypeRendererDictionary,
  DimensionFilter,
  ActionCondition,
  TreeNode,
  TreeNodeBranch,
  TreeNodeLeaf
} from './interfaces'
import { css } from '@emotion/css'
import CSS from 'utils/css'
import { CubeSchema } from 'interfaces'
import TableTag from 'components/common/partial_table_tag'
import { FieldTypeString } from './type_string'
import { FieldTypeTime } from './type_time'
import { LeafActionForm, LeafUserForm } from './form_leaf'
import { FieldTypeNumber } from './type_number'

export const HasLeaf = (node: TreeNode): boolean => {
  if (node.kind === 'leaf') return true
  if (!node.branch) return false

  return node.branch.leaves.some((child: TreeNode) => {
    return HasLeaf(child)
  })
}

export type SegmentSchemas = {
  [key: string]: CubeSchema
}

export type TreeNodeInputProps = {
  value?: TreeNode
  onChange?: (updatedValue: TreeNode) => void
  schemas: SegmentSchemas
}

const fieldTypeRendererDictionary: FieldTypeRendererDictionary = {
  string: new FieldTypeString(),
  time: new FieldTypeTime(),
  number: new FieldTypeNumber()
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

// const typeIcon = css({
//   width: '25px',
//   textAlign: 'center',
//   display: 'inline-block',
//   marginRight: CSS.M,
//   fontSize: '9px',
//   lineHeight: '23px',
//   borderRadius: '3px',
//   backgroundColor: '#eee',
//   color: '#666'
// })

export const TreeNodeInput = (props: TreeNodeInputProps) => {
  const [editingNodeLeaf, setEditingNodeLeaf] = useState<EditingNodeLeaf | undefined>(undefined)

  const cascaderOptions = useMemo(() => {
    const options: any[] = [
      { value: 'and', label: 'AND | OR' } // AND by default, user can switch to OR after
    ]

    forEach(props.schemas, (_schema: CubeSchema, tableName: string) => {
      options.push({
        value: tableName,
        label: <TableTag table={tableName} />
      })
    })

    // forEach(props.schemas, (schema: CubeSchema, tableName: string) => {
    //   const measures: any[] = []

    //   // forEach(schema.measures, (measure, key) => {
    //   //   if (measure.shown === false || measure.meta?.hide_from_segmentation === true) {
    //   //     return
    //   //   }

    //   //   // consider count/count_distinct/sum/avg/max... as number type
    //   //   const type = ['string', 'time', 'number'].includes(measure.type) ? measure.type : 'number'

    //   //   measures.push({
    //   //     value: key,
    //   //     label: (
    //   //       <Tooltip title={measure.description}>
    //   //         <span className={typeIcon}>123</span>
    //   //         {measure.title}
    //   //       </Tooltip>
    //   //     ),
    //   //     type: type
    //   //   })
    //   // })

    //   const dimensions: any[] = []

    //   forEach(schema.dimensions, (dimension, key) => {
    //     if (dimension.shown === false || dimension.meta?.hide_from_segmentation === true) {
    //       return
    //     }

    //     let icon = <span className={typeIcon}>123</span>

    //     switch (dimension.type) {
    //       case 'string':
    //         icon = <span className={typeIcon}>Abc</span>
    //         break
    //       case 'number':
    //         if (key.indexOf('is_') !== -1) {
    //           icon = <span className={typeIcon}>0/1</span>
    //         }
    //         break
    //       case 'time':
    //         icon = (
    //           <span className={typeIcon}>
    //             <FontAwesomeIcon icon={faCalendar} />
    //           </span>
    //         )
    //         break
    //       default:
    //     }

    //     dimensions.push({
    //       value: key,
    //       label: (
    //         <Tooltip title={dimension.description}>
    //           {icon} {dimension.title}
    //         </Tooltip>
    //       ),
    //       type: dimension.type
    //     })
    //   })

    //   options.push({
    //     value: tableName,
    //     label: <TableTag table={tableName} />,
    //     children: [...measures, ...dimensions]
    //   })
    // })

    return options
  }, [props.schemas])

  // borderColor incrementer
  let currentColorID = 0
  const getColorID = () => {
    currentColorID++
    return currentColorID
  }

  const cancelOrDeleteNode = (path: string, key: number) => {
    // console.log('path', path);
    // console.log('key', key);
    const clonedTree = cloneDeep(props.value) as TreeNode

    // cancel if edit, and is not new
    if (editingNodeLeaf && !editingNodeLeaf.is_new) {
      setEditingNodeLeaf(undefined)
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
    if (editingNodeLeaf && editingNodeLeaf.path === path && editingNodeLeaf.key === key) {
      setEditingNodeLeaf(undefined)
    }

    if (props.onChange) props.onChange(clonedTree)
  }

  const addTreeNode = (path: string, key: number, values: any[], selectedOptions: any) => {
    // console.log('values', values);
    // console.log('selectedOptions', selectedOptions);
    // console.log('path', path);
    // console.log('key', key);
    if (!props.value) return

    const clonedTree = cloneDeep(props.value) as TreeNode
    if (!clonedTree.branch) return

    const setPath = path + '[' + key + ']'

    // Add branch
    if (values[0] === 'and') {
      const node: TreeNode = {
        kind: 'branch',
        branch: {
          operator: 'and',
          leaves: []
        } as TreeNodeBranch
      }

      // node path, if non root
      if (path === '') {
        clonedTree.branch.leaves.push(node)
      } else {
        const target = get(clonedTree, setPath)
        target.branch.leaves.push(node)
      }

      // console.log('tree is', JSON.stringify(clonedTree, undefined, 2))
      props.onChange?.(clonedTree)
      return
    }

    // Add leaf
    const leaf = {
      table: selectedOptions[0].value,
      filters: [] as DimensionFilter[]
    } as TreeNodeLeaf

    // actions
    if (leaf.table !== 'user') {
      leaf.action = {
        timeframe_operator: 'anytime',
        timeframe_values: [],
        count_operator: 'at_least',
        count_value: 1
      } as ActionCondition
    }

    console.log('leaf', leaf)

    // // https://cube.dev/docs/product/apis-integrations/rest-api/query-format#filters-operators
    // switch (selectedOptions[1].type) {
    //   case 'string':
    //     leaf.operator = 'equals'
    //     leaf.string_values = []
    //     break
    //   case 'number':
    //     leaf.operator = 'equals'
    //     leaf.number_values = []
    //     break
    //   case 'time':
    //     leaf.operator = 'beforeDate'
    //     leaf.string_values = []
    //     break
    //   default: {
    //     console.error('operator type ' + selectedOptions[1].type + ' is not implemented')
    //     return
    //   }
    // }

    const node: TreeNode = {
      kind: 'leaf',
      leaf: leaf
    }

    let editingNodeKey = 0

    // node path, if non root
    if (path === '') {
      clonedTree.branch.leaves.push(node)
      editingNodeKey = clonedTree.branch.leaves.length - 1
    } else {
      const target = get(clonedTree, setPath)
      target.branch.leaves.push(node)
      editingNodeKey = target.branch.leaves.length - 1
    }

    const editingNodeLeaf = Object.assign(
      {
        path: path === '' ? 'branch.leaves' : setPath + '.branch.leaves',
        key: editingNodeKey,
        is_new: true
      },
      node as object
    ) as EditingNodeLeaf

    setEditingNodeLeaf(editingNodeLeaf)

    // console.log('tree is', JSON.stringify(clonedTree, undefined, 2))
    props.onChange?.(clonedTree)
  }

  const deleteButton = (path: string, pathKey: number, isBranch: boolean) => {
    return (
      <Popconfirm
        placement="left"
        title={'Do you really want to remove this ' + (isBranch ? 'branch' : 'condition') + '?'}
        onConfirm={cancelOrDeleteNode.bind(null, path, pathKey)}
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

  const editNode = (path: string, key: number) => {
    if (!props.value) return

    const condition = get(props.value, path + '[' + key + ']')

    const editingNodeLeaf = Object.assign(
      {
        path: path,
        key: key
      },
      condition
    ) as EditingNodeLeaf

    setEditingNodeLeaf(editingNodeLeaf)
  }

  const changeBranchOperator = (path: string, pathKey: number, value: any) => {
    const clonedTree = cloneDeep(props.value) as TreeNode
    if (!clonedTree.branch) return

    if (path === '') {
      clonedTree.branch.operator = value
    } else {
      set(clonedTree, path + '[' + pathKey + '].branch.operator', value)
    }

    // console.log('new tree', JSON.stringify(clonedTree, undefined, 2))
    props.onChange?.(clonedTree)
  }

  const onUpdateNode = (updatedNode: TreeNode, path: string, pathKey: number) => {
    const fullPath = path + '[' + pathKey + ']'
    // console.log('fullPath', fullPath)
    // const condition = get(props.value, path + '[' + pathKey + ']')

    const clonedTree = cloneDeep(props.value) as TreeNode
    set(clonedTree, fullPath, updatedNode)
    console.log('tree is', JSON.stringify(clonedTree, undefined, 2))
    props.onChange?.(clonedTree)
  }

  const renderLeaf = (node: TreeNode, path: string, pathKey: number) => {
    const isEditingCurrent =
      editingNodeLeaf && editingNodeLeaf.path === path && editingNodeLeaf.key === pathKey
        ? true
        : false

    const schema = props.schemas[node.leaf?.table as string]

    if (!schema) {
      return (
        <div className={css([CSS.padding_v_m, CSS.padding_l_m])}>
          <Button.Group className={CSS.pull_right}>
            {deleteButton(path, pathKey, false)}
          </Button.Group>
          <div>
            <Alert type="error" message={'table ' + node.leaf?.table + ' not found'} />
          </div>
        </div>
      )
    }

    if (isEditingCurrent && editingNodeLeaf) {
      return (
        <div className={css([CSS.padding_v_m, CSS.padding_l_m])}>
          {node.leaf?.table === 'user' && (
            <LeafUserForm
              value={node}
              onChange={(updatedLeaf: TreeNode) => {
                onUpdateNode(updatedLeaf, path, pathKey)
              }}
              table={node.leaf?.table as string}
              schema={schema}
              editingNodeLeaf={editingNodeLeaf as EditingNodeLeaf}
              setEditingNodeLeaf={setEditingNodeLeaf}
              cancelOrDeleteNode={cancelOrDeleteNode.bind(null, path, pathKey)}
            />
          )}
          {node.leaf?.table !== 'user' && (
            <LeafActionForm
              value={node}
              onChange={(updatedLeaf: TreeNode) => {
                onUpdateNode(updatedLeaf, path, pathKey)
              }}
              table={node.leaf?.table as string}
              schema={schema}
              editingNodeLeaf={editingNodeLeaf as EditingNodeLeaf}
              setEditingNodeLeaf={setEditingNodeLeaf}
              cancelOrDeleteNode={cancelOrDeleteNode.bind(null, path, pathKey)}
            />
          )}
        </div>
      )
    }

    // console.log('node', node)

    return (
      <div style={{ lineHeight: '32px' }} className={css([CSS.padding_v_m, CSS.padding_l_m])}>
        <Button.Group className={CSS.pull_right}>
          {deleteButton(path, pathKey, false)}
          <Button size="small" onClick={editNode.bind(null, path, pathKey)}>
            <FontAwesomeIcon icon={faPenToSquare} />
          </Button>
        </Button.Group>

        <div>
          <Space style={{ alignItems: 'start' }}>
            <TableTag table={node.leaf?.table as string} />
            <div>
              {node.leaf?.action && (
                <>
                  <span className={CSS.opacity_60 + ' ' + CSS.padding_r_s}>
                    happened&nbsp;
                    {node.leaf?.action.count_operator === 'at_least' && 'at least'}
                    {node.leaf?.action.count_operator === 'at_most' && 'at most'}
                    {node.leaf?.action.count_operator === 'exactly' && 'exactly'}
                  </span>
                  <Tag color="blue">{node.leaf?.action.count_value}</Tag>

                  <span className={CSS.opacity_60}>times</span>

                  {node.leaf?.action.timeframe_operator !== 'anytime' && (
                    <div>
                      <Space>
                        {node.leaf?.action.timeframe_operator === 'in_date_range' && (
                          <>
                            <span className={CSS.opacity_60}>between</span>
                            <Tag color="blue">{node.leaf?.action.timeframe_values?.[0]}</Tag>
                            &rarr;
                            <Tag className={CSS.margin_l_s} color="blue">
                              {node.leaf?.action.timeframe_values?.[1]}
                            </Tag>
                          </>
                        )}
                        {node.leaf?.action.timeframe_operator === 'before_date' && (
                          <>
                            <span className={CSS.opacity_60}>before</span>
                            <Tag color="blue">{node.leaf?.action.timeframe_values?.[0]}</Tag>
                          </>
                        )}
                        {node.leaf?.action.timeframe_operator === 'after_date' && (
                          <>
                            <span className={CSS.opacity_60}>after</span>
                            <Tag color="blue">{node.leaf?.action.timeframe_values?.[0]}</Tag>
                          </>
                        )}
                      </Space>
                    </div>
                  )}
                </>
              )}
              {node.leaf?.filters && node.leaf?.filters.length > 0 && (
                <Space style={{ alignItems: 'start' }}>
                  <span className={CSS.opacity_60}>with filters</span>
                  <table>
                    <tbody>
                      {node.leaf?.filters.map((filter, key) => {
                        const dimension = schema.dimensions[filter.field_name]
                        const fieldTypeRenderer = fieldTypeRendererDictionary[filter.field_type]

                        return (
                          <tr key={key}>
                            <td>
                              {!fieldTypeRenderer && (
                                <Alert
                                  type="error"
                                  message={'type ' + filter.field_type + ' is not implemented'}
                                />
                              )}
                              {fieldTypeRenderer && (
                                <Space key={key}>
                                  <Popover
                                    title={'field: ' + filter.field_name}
                                    content={dimension.description}
                                  >
                                    <b>{dimension.title}</b>
                                  </Popover>
                                  {fieldTypeRenderer.render(filter, dimension)}
                                </Space>
                              )}
                            </td>
                          </tr>
                        )
                      })}
                    </tbody>
                  </table>
                </Space>
              )}
            </div>
          </Space>
        </div>
      </div>
    )
  }

  const renderBranch = (node: TreeNode, path: string, pathKey: number) => {
    if (!node.branch) return <span>A branch condition is required...</span>

    const conditionPath = path === '' ? 'branch.leaves' : path + '[' + pathKey + '].branch.leaves'
    // console.log('conditionPath', conditionPath)
    const isEditing = editingNodeLeaf ? true : false
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

        {node.branch.leaves.map((leaf: TreeNode, i: number) => {
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
              {leaf.leaf && renderLeaf(leaf, conditionPath, i)}
              {leaf.branch && renderBranch(leaf, conditionPath, i)}
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
          {node.branch.leaves.length > 0 && (
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
              popupClassName={CSS.cascaderWide}
              onChange={addTreeNode.bind(null, path, pathKey)}
              expandTrigger="hover"
              options={cascaderOptions}
            >
              <Button
                size="small"
                type="primary"
                ghost={node.branch.leaves.length > 0}
                disabled={editingNodeLeaf ? true : false}
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
