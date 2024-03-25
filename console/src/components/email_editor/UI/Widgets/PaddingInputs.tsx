import React from 'react'
import { InputNumber, Input } from 'antd'

interface PaddingInputsProps {
  styles: any
  onChange: (styles: any) => void
}

const parser = (value: string | undefined) => {
  if (value === undefined) return 0
  value = value.replace('↑', '')
  value = value.replace('↓', '')
  value = value.replace('←', '')
  value = value.replace('→', '')
  value = value.replace('px', '')
  return parseInt(value)
}

const PaddingInputs = (props: PaddingInputsProps) => {
  return (
    <>
      <InputNumber
        style={{ width: '50%', right: '25%' }}
        value={parseInt(props.styles.paddingTop || '0px')}
        onChange={(value) => {
          props.styles.paddingTop = value + 'px'
          props.onChange(props.styles)
        }}
        defaultValue={parseInt(props.styles.paddingTop || '0px')}
        size="small"
        step={1}
        min={0}
        parser={parser}
        formatter={(value) => '↑  ' + value + 'px'}
      />

      <Input.Group style={{ width: '100%' }} size="small" compact>
        <InputNumber
          style={{ width: '50%' }}
          value={parseInt(props.styles.paddingLeft || '0px')}
          onChange={(value) => {
            props.styles.paddingLeft = value + 'px'
            props.onChange(props.styles)
          }}
          defaultValue={parseInt(props.styles.paddingLeft || '0px')}
          size="small"
          step={1}
          min={0}
          parser={parser}
          formatter={(value) => '←  ' + value + 'px'}
        />
        <InputNumber
          style={{ width: '50%' }}
          value={parseInt(props.styles.paddingRight || '0px')}
          onChange={(value) => {
            props.styles.paddingRight = value + 'px'
            props.onChange(props.styles)
          }}
          defaultValue={parseInt(props.styles.paddingRight || '0px')}
          size="small"
          step={1}
          min={0}
          parser={parser}
          formatter={(value) => '→  ' + value + 'px'}
        />
      </Input.Group>
      <InputNumber
        style={{ width: '50%', right: '25%' }}
        value={parseInt(props.styles.paddingBottom || '0px')}
        onChange={(value) => {
          props.styles.paddingBottom = value + 'px'
          props.onChange(props.styles)
        }}
        defaultValue={parseInt(props.styles.paddingBottom || '0px')}
        size="small"
        step={1}
        min={0}
        parser={parser}
        formatter={(value) => '↓  ' + value + 'px'}
      />
    </>
  )
}

export default PaddingInputs
