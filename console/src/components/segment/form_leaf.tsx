import { Dispatch, SetStateAction } from 'react'
import { cloneDeep } from 'lodash'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faClose } from '@fortawesome/free-solid-svg-icons'
import { Button, Input, Form, Select, InputNumber, Space, DatePicker } from 'antd'
import { useForm } from 'antd/lib/form/Form'
import { TreeNode, EditingNodeLeaf, TreeNodeLeaf } from './interfaces'
import CSS from 'utils/css'
import TableTag from 'components/common/partial_table_tag'
import dayjs from 'dayjs'
import { CubeSchema } from 'interfaces'
import { InputDimensionFilters } from './input_dimension_filters'
import Messages from 'utils/formMessages'

export type LeafFormProps = {
  value?: TreeNode
  onChange?: (updatedLeaf: TreeNode) => void
  table: string
  schema: CubeSchema
  editingNodeLeaf: EditingNodeLeaf
  setEditingNodeLeaf: Dispatch<SetStateAction<EditingNodeLeaf | undefined>>
  cancelOrDeleteNode: () => void
}

export const LeafUserForm = (props: LeafFormProps) => {
  const [form] = useForm()

  const onSubmit = () => {
    form
      .validateFields()
      .then((values) => {
        console.log('values', values)
        if (!props.value) return

        // convert dayjs values into strings
        // if (values.field_type === 'time') {
        //   values.string_values.forEach((value: any, index: number) => {
        //     values.string_values[index] = value.format('YYYY-MM-DD HH:mm:ss')
        //   })
        // }

        const clonedLeaf = cloneDeep(props.value)
        clonedLeaf.leaf = Object.assign(clonedLeaf.leaf as TreeNodeLeaf, values)

        props.setEditingNodeLeaf(undefined)

        if (props.onChange) props.onChange(clonedLeaf)
      })
      .catch((_e) => {})
  }

  console.log('props', props)

  return (
    <Form component="div" layout="inline" form={form} initialValues={props.editingNodeLeaf.leaf}>
      <Form.Item style={{ margin: 0 }} name="table" colon={false} label={<TableTag table="user" />}>
        <Input hidden />
      </Form.Item>
      <Form.Item
        style={{ margin: 0, width: 500 }}
        name="filters"
        colon={false}
        rules={[{ required: true, type: 'array', min: 1, message: Messages.RequiredField }]}
      >
        <InputDimensionFilters schema={props.schema} />
      </Form.Item>

      {/* CONFIRM / CANCEL */}
      <Form.Item noStyle shouldUpdate>
        {(funcs) => {
          const filters = funcs.getFieldValue('filters')

          return (
            <Form.Item style={{ position: 'absolute', right: 0, top: 16 }}>
              <Button size="small" onClick={() => props.cancelOrDeleteNode()}>
                <FontAwesomeIcon icon={faClose} />
              </Button>
              {filters && filters.length > 0 && (
                <Button type="primary" size="small" className={CSS.margin_l_s} onClick={onSubmit}>
                  Confirm
                </Button>
              )}
            </Form.Item>
          )
        }}
      </Form.Item>
    </Form>
  )
}

export const LeafActionForm = (props: LeafFormProps) => {
  const [form] = useForm()

  const onSubmit = () => {
    form
      .validateFields()
      .then((values) => {
        console.log('values', values)
        if (!props.value) return

        // convert dayjs values into strings
        // if (values.field_type === 'time') {
        //   values.string_values.forEach((value: any, index: number) => {
        //     values.string_values[index] = value.format('YYYY-MM-DD HH:mm:ss')
        //   })
        // }

        const clonedLeaf = cloneDeep(props.value)
        clonedLeaf.leaf = Object.assign(clonedLeaf.leaf as TreeNodeLeaf, values)

        props.setEditingNodeLeaf(undefined)

        if (props.onChange) props.onChange(clonedLeaf)
      })
      .catch((e) => {
        console.log(e)
      })
  }

  // console.log('props', props)

  return (
    <Space style={{ alignItems: 'start' }}>
      <TableTag table={props.table} />
      <Form
        component="div"
        layout="vertical"
        form={form}
        initialValues={props.editingNodeLeaf.leaf}
      >
        <Form.Item name="table" noStyle>
          <Input hidden />
        </Form.Item>

        <Space>
          <span className={CSS.opacity_60}>happened</span>
          <Form.Item noStyle name={['action', 'count_operator']} colon={false}>
            <Select
              style={{}}
              size="small"
              options={[
                { value: 'at_least', label: 'at least' },
                { value: 'at_most', label: 'at most' },
                { value: 'exactly', label: 'exactly' }
              ]}
            />
          </Form.Item>
          <Form.Item
            noStyle
            name={['action', 'count_value']}
            colon={false}
            rules={[{ required: true, type: 'number', min: 1, message: Messages.RequiredField }]}
          >
            <InputNumber style={{ width: 70 }} size="small" />
          </Form.Item>
          <span className={CSS.opacity_60}>times</span>
        </Space>

        <div className={CSS.margin_t_m}>
          <Space>
            <span className={CSS.opacity_60}>timeframe</span>
            <Form.Item noStyle name={['action', 'timeframe_operator']} colon={false}>
              <Select
                style={{ width: 130 }}
                size="small"
                options={[
                  { value: 'anytime', label: 'anytime' },
                  { value: 'in_date_range', label: 'in date range' },
                  { value: 'before_date', label: 'before date' },
                  { value: 'after_date', label: 'after date' }
                ]}
              />
            </Form.Item>
            <Form.Item noStyle shouldUpdate>
              {(funcs) => {
                const timeframe_operator = funcs.getFieldValue(['action', 'timeframe_operator'])

                if (timeframe_operator === 'in_date_range') {
                  return (
                    <Form.Item
                      noStyle
                      name={['action', 'timeframe_values']}
                      colon={false}
                      rules={[
                        { required: true, type: 'array', min: 2, message: Messages.RequiredField }
                      ]}
                      dependencies={['action', 'timeframe_operator']}
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
                        style={{ width: 370 }}
                        size="small"
                        showTime={{
                          defaultValue: [dayjs().startOf('day'), dayjs().startOf('day')]
                        }}
                      />
                    </Form.Item>
                  )
                } else if (
                  timeframe_operator === 'before_date' ||
                  timeframe_operator === 'after_date'
                ) {
                  return (
                    <Form.Item
                      noStyle
                      name={['action', 'timeframe_values', 0]}
                      colon={false}
                      dependencies={['action', 'timeframe_operator']}
                      rules={[{ required: true, type: 'string', message: Messages.RequiredField }]}
                      getValueProps={(value: any) => {
                        return { value: value ? dayjs(value) : undefined }
                      }}
                      getValueFromEvent={(_date: any, dateString: string) => dateString}
                    >
                      <DatePicker
                        style={{ width: 180 }}
                        size="small"
                        showTime={{ defaultValue: dayjs().startOf('day') }}
                      />
                    </Form.Item>
                  )
                } else {
                  return null
                }
              }}
            </Form.Item>
            {/* <Form.Item
            noStyle
            name={['action', 'timeframe_values']}
            colon={false}
            rules={[{ required: true, type: 'number', min: 1, message: Messages.RequiredField }]}
          >
            <InputNumber style={{ width: 70 }} size="small" />
          </Form.Item> */}
          </Space>
        </div>

        <div className={CSS.margin_t_m}>
          <Space style={{ alignItems: 'start' }}>
            <span className={CSS.opacity_60}>with filters</span>
            <Form.Item
              name="filters"
              noStyle
              colon={false}
              className={CSS.margin_t_s}
              rules={[{ required: false, type: 'array', min: 0, message: Messages.RequiredField }]}
            >
              <InputDimensionFilters schema={props.schema} btnType="link" btnGhost={true} />
            </Form.Item>
          </Space>
        </div>

        {/* CONFIRM / CANCEL */}
        <div style={{ position: 'absolute', top: 16, right: 0 }}>
          <Button size="small" onClick={() => props.cancelOrDeleteNode()}>
            <FontAwesomeIcon icon={faClose} />
          </Button>
          <Button type="primary" size="small" className={CSS.margin_l_s} onClick={onSubmit}>
            Confirm
          </Button>
        </div>
      </Form>
    </Space>
  )
}
