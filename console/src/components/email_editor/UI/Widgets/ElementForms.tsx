import React from 'react'
import { Form, InputNumber, Select, Input, Radio } from 'antd'
import PaddingInputs from '../Widgets/PaddingInputs'
import { ColorPickerLight } from '../Widgets/ColorPicker'

export const Fonts = [
  { label: 'Arial, sans-serif', value: 'Arial, sans-serif' },
  { label: 'Verdana, sans-serif', value: 'Verdana, sans-serif' },
  { label: 'Helvetica, sans-serif', value: 'Helvetica, sans-serif' },
  { label: 'Georgia, serif', value: 'Georgia, serif' },
  { label: 'Tahoma, sans-serif', value: 'Tahoma, sans-serif' },
  { label: 'Lucida, sans-serif', value: 'Lucida, sans-serif' },
  { label: 'Trebuchet MS, sans-serif', value: 'Trebuchet MS, sans-serif' },
  { label: 'Times New Roman, serif', value: 'Times New Roman, serif' }
]

const fontWeights = [
  { label: <span style={{ fontWeight: 100 }}>100</span>, value: 100 },
  { label: <span style={{ fontWeight: 200 }}>200</span>, value: 200 },
  { label: <span style={{ fontWeight: 300 }}>300</span>, value: 300 },
  { label: <span style={{ fontWeight: 400 }}>400</span>, value: 400 },
  { label: <span style={{ fontWeight: 500 }}>500</span>, value: 500 },
  { label: <span style={{ fontWeight: 600 }}>600</span>, value: 600 },
  { label: <span style={{ fontWeight: 700 }}>700</span>, value: 700 },
  { label: <span style={{ fontWeight: 800 }}>800</span>, value: 800 },
  { label: <span style={{ fontWeight: 900 }}>900</span>, value: 900 }
]

const labelProps: any = {
  labelAlign: 'left',
  className: 'cmeditor-form-item-align-right',
  labelCol: { span: 10 },
  wrapperCol: { span: 14 }
}

const ElementForms = (props: any) => {
  // console.log('SQQSD', block)
  return (
    <>
      <Form.Item label="Font style" {...labelProps}>
        <Input.Group style={{ width: '100%', marginBottom: '8px' }} size="small" compact>
          <ColorPickerLight
            style={{ width: '13%' }}
            size="small"
            value={props.block.data.styles[props.element].color}
            onChange={(newColor) => {
              props.block.data.styles[props.element].color = newColor
              props.updateTree(props.block.path, props.block)
            }}
          />
          <InputNumber
            style={{ width: '42%' }}
            value={parseInt(props.block.data.styles[props.element].fontSize || '13px')}
            onChange={(value) => {
              props.block.data.styles[props.element].fontSize = value + 'px'
              props.updateTree(props.block.path, props.block)
            }}
            size="small"
            step={1}
            min={0}
            parser={(value: string | undefined) => {
              if (value === undefined) return 0
              return parseInt(value.replace('px', ''))
            }}
            formatter={(value) => value + 'px'}
          />
          <Select
            size="small"
            style={{ width: '45%' }}
            value={props.block.data.styles[props.element].fontWeight}
            onChange={(value) => {
              props.block.data.styles[props.element].fontWeight = value
              props.updateTree(props.block.path, props.block)
            }}
            defaultValue={props.block.data.styles[props.element].fontWeight}
            options={fontWeights}
          />
        </Input.Group>
      </Form.Item>

      <Form.Item label="Font family" {...labelProps}>
        <Select
          size="small"
          options={Fonts}
          value={props.block.data.styles[props.element].fontFamily}
          onChange={(newColor) => {
            props.block.data.styles[props.element].fontFamily = newColor
            props.updateTree(props.block.path, props.block)
          }}
        />
      </Form.Item>

      <Form.Item label="Padding control" {...labelProps}>
        <Radio.Group
          style={{ width: '100%' }}
          onChange={(e) => {
            props.block.data.styles[props.element].paddingControl = e.target.value
            props.updateTree(props.block.path, props.block)
          }}
          value={props.block.data.styles[props.element].paddingControl}
          optionType="button"
          size="small"
          // buttonStyle="solid"
        >
          <Radio.Button style={{ width: '50%', textAlign: 'center' }} value="all">
            All
          </Radio.Button>
          <Radio.Button style={{ width: '50%', textAlign: 'center' }} value="separate">
            Separate
          </Radio.Button>
        </Radio.Group>
      </Form.Item>

      {props.block.data.styles[props.element].paddingControl === 'all' && (
        <>
          <Form.Item label="Paddings" {...labelProps}>
            <InputNumber
              style={{ width: '100%' }}
              value={parseInt(props.block.data.styles[props.element].padding || '0px')}
              onChange={(value) => {
                props.block.data.styles[props.element].padding = value + 'px'
                props.updateTree(props.block.path, props.block)
              }}
              size="small"
              step={1}
              min={0}
              parser={(value: string | undefined) => {
                // if (['▭'].indexOf(value)) {
                //     value = value.substring(1)
                // }
                if (value === undefined) return 0
                return parseInt(value.replace('px', ''))
              }}
              // formatter={value => '▭  ' + value + 'px'}
              formatter={(value) => value + 'px'}
            />
          </Form.Item>
        </>
      )}

      {props.block.data.styles[props.element].paddingControl === 'separate' && (
        <>
          <Form.Item label="Paddings" {...labelProps}>
            <PaddingInputs
              styles={props.block.data.styles[props.element]}
              onChange={(updatedStyles: any) => {
                props.block.data.styles[props.element] = updatedStyles
                props.updateTree(props.block.path, props.block)
              }}
            />
          </Form.Item>
        </>
      )}
    </>
  )
}

export default ElementForms
