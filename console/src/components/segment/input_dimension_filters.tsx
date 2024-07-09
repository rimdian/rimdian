import { CubeSchema } from 'interfaces'
import { DimensionFilter, FieldTypeRendererDictionary } from './interfaces'
import { Alert, Button, Form, Modal, Popconfirm, Popover, Select, Space, Tooltip } from 'antd'
import { useState } from 'react'
import { clone, map } from 'lodash'
import { css } from '@emotion/css'
import CSS from 'utils/css'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCalendar, faTrashAlt } from '@fortawesome/free-regular-svg-icons'
import { FieldTypeString } from './type_string'
import { FieldTypeTime } from './type_time'
import { FieldTypeNumber } from './type_number'

const typeIcon = css({
  width: '25px',
  textAlign: 'center',
  display: 'inline-block',
  marginRight: CSS.M,
  fontSize: '9px',
  lineHeight: '23px',
  borderRadius: '3px',
  backgroundColor: '#eee',
  color: '#666'
})

const fieldTypeRendererDictionary: FieldTypeRendererDictionary = {
  string: new FieldTypeString(),
  time: new FieldTypeTime(),
  number: new FieldTypeNumber()
}

export const InputDimensionFilters = (props: {
  value?: DimensionFilter[]
  onChange?: (updatedValue: DimensionFilter[]) => void
  schema: CubeSchema
  btnType?: string
  btnGhost?: boolean
}) => {
  const hasFilter = props.value && props.value.length > 0 ? true : false

  return (
    <span>
      {hasFilter && (
        <table>
          <tbody>
            {(props.value || []).map((filter, key) => {
              const dimension = props.schema.dimensions[filter.field_name]
              const fieldTypeRenderer = fieldTypeRendererDictionary[filter.field_type]

              return (
                <tr key={key}>
                  <td className={CSS.padding_b_s}>
                    {!fieldTypeRenderer && (
                      <Alert
                        type="error"
                        message={'type ' + filter.field_type + ' is not implemented'}
                      />
                    )}
                    {fieldTypeRenderer && (
                      <Space>
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
                  <td>
                    <Button.Group>
                      <Popconfirm
                        title="Do you really want to remove this filter?"
                        onConfirm={() => {
                          if (!props.onChange) return
                          const clonedValue = props.value ? [...props.value] : []
                          clonedValue.splice(clonedValue.indexOf(filter), 1)
                          props.onChange(clonedValue)
                        }}
                      >
                        <Button size="small" type="link">
                          <FontAwesomeIcon icon={faTrashAlt} />
                        </Button>
                      </Popconfirm>
                    </Button.Group>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      )}

      <AddFilterButton
        schema={props.schema}
        existingFilters={props.value}
        btnType={props.btnType}
        btnGhost={props.btnGhost || hasFilter}
        onComplete={(values: DimensionFilter) => {
          if (!props.onChange) return
          const clonedValue = props.value ? [...props.value] : []
          clonedValue.push(values)
          props.onChange(clonedValue)
        }}
      />
    </span>
  )
}

const AddFilterButton = (props: {
  existingFilters?: DimensionFilter[]
  onComplete: any
  schema: CubeSchema
  btnType?: any
  btnGhost?: boolean
}) => {
  const [form] = Form.useForm()
  const [modalVisible, setModalVisible] = useState(false)

  const onClicked = () => {
    setModalVisible(true)
  }

  // clone dimensions, and remove existing filters
  const availableDimensions = clone(props.schema.dimensions)
  if (props.existingFilters) {
    props.existingFilters.forEach((filter) => {
      delete availableDimensions[filter.field_name]
    })
  }

  return (
    <>
      <Button
        className={props.existingFilters && props.existingFilters.length > 0 ? CSS.margin_t_s : ''}
        type={props.btnType || 'primary'}
        size="small"
        ghost={props.btnGhost}
        onClick={onClicked}
      >
        + Add filter
      </Button>

      {modalVisible && (
        <Modal
          open={true}
          title="Add a filter"
          okText="Confirm"
          width={400}
          cancelText="Cancel"
          onCancel={() => {
            form.resetFields()
            setModalVisible(false)
          }}
          onOk={() => {
            form
              .validateFields()
              .then((values: any) => {
                form.resetFields()
                setModalVisible(false)
                values.field_type = props.schema.dimensions[values.field_name].type
                props.onComplete(values)
              })
              .catch(console.error)
          }}
        >
          <Form form={form} name="form_add_filter" layout="vertical" className={CSS.margin_v_l}>
            <Form.Item
              name="field_name"
              rules={[{ required: true, type: 'string', message: 'Please select a dimension' }]}
            >
              <Select
                // style={{ width: 200 }}
                listHeight={500}
                showSearch
                dropdownMatchSelectWidth={false}
                placeholder="Select a dimension"
                options={map(availableDimensions, (dimension, field) => {
                  // console.log('dimension', dimension)

                  let icon = <span className={typeIcon}>123</span>

                  switch (dimension.type) {
                    case 'string':
                      icon = <span className={typeIcon}>Abc</span>
                      break
                    case 'number':
                      if (field.indexOf('is_') !== -1 || field.indexOf('consent_') !== -1) {
                        icon = <span className={typeIcon}>0/1</span>
                      }
                      break
                    case 'time':
                      icon = (
                        <span className={typeIcon}>
                          <FontAwesomeIcon icon={faCalendar} />
                        </span>
                      )
                      break
                    default:
                  }

                  return {
                    label: (
                      <Tooltip title={dimension.description}>
                        {icon} {dimension.title}
                      </Tooltip>
                    ),
                    value: field
                  }
                })}
              />
            </Form.Item>

            <Form.Item noStyle shouldUpdate>
              {(funcs) => {
                const field_name = funcs.getFieldValue('field_name')
                if (!field_name) return null

                const fieldTypeRenderer =
                  fieldTypeRendererDictionary[props.schema.dimensions[field_name].type]

                if (!fieldTypeRenderer)
                  return (
                    <Alert
                      type="error"
                      message={
                        'type ' + props.schema.dimensions[field_name].type + ' is not implemented'
                      }
                    />
                  )

                return fieldTypeRenderer.renderFormItems(
                  props.schema.dimensions[field_name].type as any,
                  field_name,
                  form
                )
              }}
            </Form.Item>
          </Form>
        </Modal>
      )}
    </>
  )
}
