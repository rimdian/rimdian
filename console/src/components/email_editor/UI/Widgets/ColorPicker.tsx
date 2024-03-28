import { useState, useEffect } from 'react'
import { SketchPicker } from 'react-color'
import { Input, Dropdown } from 'antd'

type ColorPickerValue = string | undefined

interface ColorPickerInputProps {
  size?: 'small' | 'large' | undefined
  value: ColorPickerValue
  style?: object
  className?: string
  onChange: (value: ColorPickerValue) => void
  onMouseDown?: (event: any) => void
  onBlur?: (event: any) => void
}

const presetColors = [
  'transparent',
  '#F44336',
  '#E91E63',
  '#9C27B0',
  '#673AB7',
  '#3F51B5',
  '#2196F3',
  '#03A9F4',
  '#00BCD4',
  '#009688',
  '#4CAF50',
  '#8BC34A',
  '#CDDC39',
  '#FFEB3B',
  '#FFC107',
  '#FF9800',
  '#FF5722',
  '#795548',
  '#9E9E9E',
  '#546E7A',
  '#FFFFFF',
  '#000000'
]

const ColorPickerInput = (props: ColorPickerInputProps) => {
  const [hexColor, setHexColor] = useState(props.value)

  useEffect(() => {
    setHexColor(props.value)
  }, [props.value])

  // wrap Input with a span otherwise dropdown click doesnt trigger on prefix
  return (
    <span
      onMouseDown={props.onMouseDown ? props.onMouseDown : () => {}}
      onBlur={props.onBlur ? props.onBlur : () => {}}
    >
      <Dropdown
        overlayStyle={{ minWidth: 'none' }}
        onOpenChange={(visible: boolean) => {
          // trigger update when Input has been manually updated
          if (visible === false && hexColor !== props.value) {
            props.onChange(hexColor)
          }
        }}
        trigger={['click']}
        dropdownRender={() => {
          return (
            <span>
              <div style={{ color: 'black' }} onClick={(e) => e.stopPropagation()}>
                <SketchPicker
                  color={hexColor}
                  disableAlpha={true}
                  onChange={(color: any) => {
                    setHexColor(color.hex === 'transparent' ? undefined : color.hex)
                  }}
                  onChangeComplete={(color: any) => {
                    // called after drag knobs
                    props.onChange(color.hex === 'transparent' ? undefined : color.hex)
                  }}
                  presetColors={presetColors}
                />
              </div>
            </span>
          )
        }}
      >
        <span>
          <Input
            onMouseDown={props.onMouseDown ? props.onMouseDown : () => {}}
            onBlur={props.onBlur ? props.onBlur : () => {}}
            placeholder="None"
            size={props.size}
            allowClear={true}
            value={hexColor}
            onChange={(e) => {
              setHexColor(e.target.value === '' ? undefined : e.target.value)
            }}
            prefix={
              <span
                style={{
                  // border: '1px solid rgba(0,0,0,0.6)',
                  boxShadow: '0px 0px 3px 0px rgba(0,0,0,0.5)',
                  borderRadius: '3px',
                  background:
                    hexColor === undefined
                      ? 'url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAMUlEQVQ4T2NkYGAQYcAP3uCTZhw1gGGYhAGBZIA/nYDCgBDAm9BGDWAAJyRCgLaBCAAgXwixzAS0pgAAAABJRU5ErkJggg==) #FFFFFF'
                      : props.value,
                  width: props.size === 'small' ? '14px' : '20px',
                  height: props.size === 'small' ? '14px' : '20px',
                  marginRight: props.size === 'small' ? '6px' : '8px',
                  cursor: 'pointer'
                }}
              ></span>
            }
          />
        </span>
      </Dropdown>
    </span>
  )
}

export const ColorPickerLight = (props: ColorPickerInputProps) => {
  const [hexColor, setHexColor] = useState(props.value)

  useEffect(() => {
    setHexColor(props.value)
  }, [props.value])

  const styles = Object.assign({}, props.style || {}, {
    background:
      hexColor === undefined
        ? 'url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAMUlEQVQ4T2NkYGAQYcAP3uCTZhw1gGGYhAGBZIA/nYDCgBDAm9BGDWAAJyRCgLaBCAAgXwixzAS0pgAAAABJRU5ErkJggg==) #FFFFFF'
        : hexColor,
    height: props.size === 'small' ? '24px' : '32px'
  })

  // wrap Input with a span otherwise dropdown click doesnt trigger on prefix
  return (
    <Dropdown
      overlayStyle={{ minWidth: 'none' }}
      onOpenChange={(visible: boolean) => {
        // trigger update when Input has been manually updated
        if (visible === false && hexColor !== props.value) {
          props.onChange(hexColor)
        }
        if (visible === false && props.onBlur) props.onBlur(null)
      }}
      trigger={['click']}
      overlay={
        <span>
          <div style={{ color: 'black' }} onClick={(e) => e.stopPropagation()}>
            <SketchPicker
              color={hexColor}
              disableAlpha={true}
              onChange={(color: any) => {
                setHexColor(color.hex === 'transparent' ? undefined : color.hex)
              }}
              onChangeComplete={(color: any) => {
                // called after drag knobs
                // console.log('complete')
                props.onChange(color.hex === 'transparent' ? undefined : color.hex)
              }}
              presetColors={presetColors}
            />
          </div>
        </span>
      }
    >
      <span
        onMouseDown={props.onMouseDown || undefined}
        className={'rmdeditor-color-picker-input-light ' + props.className}
        style={styles}
      ></span>
    </Dropdown>
  )
}

export default ColorPickerInput
