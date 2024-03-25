import React from 'react'
import { ColorPickerLight } from './ColorPicker'
import { InputNumber, Select, Input } from 'antd'

interface FontStyleInputsProps {
  styles: any
  onChange: (styles: any) => void
  required?: boolean
}

const FontStyleInputs = (props: FontStyleInputsProps) => {
  return (
    <>
      <Input.Group style={{ width: '100%', marginBottom: '8px' }} size="small" compact>
        <ColorPickerLight
          style={{ width: '13%' }}
          size="small"
          value={props.styles.color}
          onChange={(newColor) => {
            props.styles.color = newColor
            props.onChange(props.styles)
          }}
        />
        <InputNumber
          style={{ width: '42%' }}
          value={parseInt(props.styles.fontSize || '13px')}
          onChange={(value) => {
            props.styles.fontSize = value + 'px'
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
          style={{ width: '45%' }}
          value={props.styles.fontWeight}
          onChange={(value) => {
            props.styles.fontWeight = value
            props.onChange(props.styles)
          }}
          defaultValue={props.styles.fontWeight}
          options={[
            { label: <span style={{ fontWeight: 100 }}>100</span>, value: 100 },
            { label: <span style={{ fontWeight: 200 }}>200</span>, value: 200 },
            { label: <span style={{ fontWeight: 300 }}>300</span>, value: 300 },
            { label: <span style={{ fontWeight: 400 }}>400</span>, value: 400 },
            { label: <span style={{ fontWeight: 500 }}>500</span>, value: 500 },
            { label: <span style={{ fontWeight: 600 }}>600</span>, value: 600 },
            { label: <span style={{ fontWeight: 700 }}>700</span>, value: 700 },
            { label: <span style={{ fontWeight: 800 }}>800</span>, value: 800 },
            { label: <span style={{ fontWeight: 900 }}>900</span>, value: 900 }
          ]}
        />
      </Input.Group>

      <Input.Group style={{ width: '100%' }} size="small" compact>
        <Select
          size="small"
          style={{ width: '55%' }}
          value={props.styles.textTransform}
          onChange={(value) => {
            props.styles.textTransform = value
            props.onChange(props.styles)
          }}
          defaultValue={props.styles.textTransform}
          options={[
            { label: 'None', value: 'none' },
            { label: 'Uppercase', value: 'uppercase' },
            { label: 'Capitalize', value: 'capitalize' },
            { label: 'Lowercase', value: 'lowercase' }
          ]}
        />
        <Select
          size="small"
          style={{ width: '45%' }}
          value={props.styles.fontStyle}
          onChange={(value) => {
            props.styles.fontStyle = value
            props.onChange(props.styles)
          }}
          defaultValue={props.styles.fontStyle}
          options={[
            { label: <span style={{ fontStyle: 'normal' }}>Normal</span>, value: 'normal' },
            { label: <span style={{ fontStyle: 'italic' }}>Italic</span>, value: 'italic' }
            // { label: <span style={{ fontStyle: 'oblique' }}>Oblique</span>, value: 'oblique' },
          ]}
        />
      </Input.Group>
    </>
  )
}

export default FontStyleInputs
