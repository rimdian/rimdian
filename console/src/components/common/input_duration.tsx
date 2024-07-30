import { InputNumber, Select } from 'antd'

type InputDurationProps = {
  onChange?: (value: Array<number | string | null>) => void
  value?: Array<any>
  size?: 'small' | 'middle' | 'large'
  disabled?: boolean
}

const InputDuration = (props: InputDurationProps) => {
  let inputValue: number | null = null
  if (props.value && typeof props.value[0] === 'number') {
    inputValue = props.value[0] as number
  }
  const unitValue = props.value ? (props.value[1] as string) : null

  const onNumberChange = (value: number | null) => {
    if (value === null) {
      props.onChange?.([null, unitValue])
      return
    }
    const newValue = value + ' ' + (unitValue || 'days')
    console.log('newValue', newValue)
    props.onChange?.([value, unitValue || 'days'])
  }

  const onUnitChange = (value: string) => {
    props.onChange?.([inputValue || 1, value])
  }

  return (
    <InputNumber
      size={props.size}
      step={1}
      min={1}
      style={{ width: 200 }}
      onChange={onNumberChange}
      value={inputValue}
      addonAfter={
        <>
          <Select
            size={props.size}
            defaultValue="days"
            style={{ width: 100 }}
            disabled={props.disabled}
            value={unitValue}
            onChange={onUnitChange}
            options={[
              { value: 'hours', label: 'hours' },
              { value: 'days', label: 'days' }
            ]}
          />
        </>
      }
    />
  )
}

export default InputDuration
