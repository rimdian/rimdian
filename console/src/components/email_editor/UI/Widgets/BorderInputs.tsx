import React from 'react'
import { ColorPickerLight } from './ColorPicker'
import { InputNumber, Select, Input } from 'antd'

interface BorderInputsProps {
  styles: any
  propertyPrefix: string
  onChange: (styles: any) => void
  required?: boolean
}

const BorderInputs = (props: BorderInputsProps) => {
  const options = []
  const defaultOption = props.required ? 'solid' : 'none'

  if (!props.required) {
    options.push({ label: 'None', value: 'none' })
  }

  options.push({
    label: (
      <>
        <span
          style={{
            display: 'inline-block',
            borderTop: '2px solid black',
            marginTop: '10px',
            width: '32px'
          }}
        >
          &nbsp;
        </span>
      </>
    ),
    value: 'solid'
  })
  options.push({
    label: (
      <>
        <span
          style={{
            display: 'inline-block',
            borderTop: '2px dashed black',
            marginTop: '10px',
            width: '32px'
          }}
        >
          &nbsp;
        </span>
      </>
    ),
    value: 'dashed'
  })
  options.push({
    label: (
      <>
        <span
          style={{
            display: 'inline-block',
            borderTop: '2px dotted black',
            marginTop: '10px',
            width: '32px'
          }}
        >
          &nbsp;
        </span>
      </>
    ),
    value: 'dotted'
  })
  options.push({
    label: (
      <>
        <span
          style={{
            display: 'inline-block',
            borderTop: '4px double black',
            marginTop: '10px',
            width: '32px'
          }}
        >
          &nbsp;
        </span>
      </>
    ),
    value: 'double'
  })

  return (
    <Input.Group style={{ width: '100%' }} size="small" compact>
      <ColorPickerLight
        style={{ width: '13%' }}
        size="small"
        value={props.styles[props.propertyPrefix + 'Color']}
        onChange={(newColor) => {
          props.styles[props.propertyPrefix + 'Color'] = newColor
          props.onChange(props.styles)
        }}
      />
      <InputNumber
        style={{ width: '42%' }}
        value={parseInt(props.styles[props.propertyPrefix + 'Width'] || '0px')}
        onChange={(value) => {
          props.styles[props.propertyPrefix + 'Width'] = value + 'px'
          props.onChange(props.styles)
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
        style={{ width: '44%' }}
        value={props.styles[props.propertyPrefix + 'Style']}
        onChange={(value) => {
          props.styles[props.propertyPrefix + 'Style'] = value
          props.onChange(props.styles)
        }}
        defaultValue={defaultOption}
        options={options}
      />
    </Input.Group>
  )
}

export default BorderInputs
