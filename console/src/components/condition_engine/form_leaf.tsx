import { Dispatch, SetStateAction } from 'react'
import { cloneDeep } from 'lodash'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faClose } from '@fortawesome/free-solid-svg-icons'
import { Button, Tag, Input, Form } from 'antd'
import { useForm } from 'antd/lib/form/Form'
import {
  Condition,
  ConditionLeaf,
  EditingConditionLeaf,
  FieldDefinition,
  FieldTypeRenderer
} from './interfaces'
import CSS from 'utils/css'

export type LeafFormProps = {
  value?: Condition
  onChange?: (updatedLeaf: Condition) => void
  fieldTypeRenderer: FieldTypeRenderer
  fieldDefinition: FieldDefinition
  editingConditionLeaf: EditingConditionLeaf
  setEditingConditionLeaf: Dispatch<SetStateAction<EditingConditionLeaf | undefined>>
  cancelOrDeleteCondition: () => void
}

export const LeafForm = (props: LeafFormProps) => {
  const [form] = useForm()

  const onSubmit = () => {
    form
      .validateFields()
      .then((values) => {
        // console.log('values', values)
        if (!props.value) return

        const clonedLeaf = cloneDeep(props.value)
        clonedLeaf.leaf = Object.assign(clonedLeaf.leaf as ConditionLeaf, values)

        props.setEditingConditionLeaf(undefined)

        if (props.onChange) props.onChange(clonedLeaf)
      })
      .catch((_e) => {})
  }

  return (
    <Form
      component="div"
      layout="inline"
      form={form}
      initialValues={props.editingConditionLeaf.leaf}
    >
      <Form.Item
        style={{ margin: 0 }}
        name="field"
        colon={false}
        label={<Tag color="purple">{props.fieldDefinition.label}</Tag>}
      >
        <Input hidden />
      </Form.Item>

      {props.fieldTypeRenderer.renderFormItems(props.editingConditionLeaf, form)}

      {/* CONFIRM / CANCEL */}
      <Form.Item>
        <Button type="primary" size="small" className={CSS.margin_r_s} onClick={onSubmit}>
          Confirm
        </Button>
        <Button size="small" onClick={() => props.cancelOrDeleteCondition()}>
          <FontAwesomeIcon icon={faClose} />
        </Button>
      </Form.Item>
    </Form>
  )
}
