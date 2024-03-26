import React, { useCallback, useState, useMemo, useRef, useEffect, PropsWithChildren } from 'react'
import { BaseEditor, BaseRange, Node } from 'slate'
import { Slate, Editable, ReactEditor, withReact, useSlate } from 'slate-react'
import { Editor, Transforms, Text, createEditor, Element as SlateElement, Descendant } from 'slate'
import cn from 'classnames'
import { withHistory, HistoryEditor } from 'slate-history'
import { BoldOutlined, ItalicOutlined, UnderlineOutlined } from '@ant-design/icons'
import { Select } from 'antd'
import { ColorPickerLight } from './ColorPicker'

const LIST_TYPES = ['numbered-list', 'bulleted-list']

interface BaseProps {
  className: string
  [key: string]: unknown
}
// type OrNull<T> = T | null

export interface MyEditorProps {
  isFocused: boolean
  onChange: (data: any) => void
  styles: any
  value: any
  // type: string
  toolbarButtons: string[]
}

interface CustomEditor extends ReactEditor {
  savedSelection: BaseRange | null
}

type FormattedText = {
  text: string
  bold?: boolean
  italic?: boolean
  underline?: boolean
  fontSize?: number
  fontFamily?: string
  fontColor?: string
}

type CustomText = FormattedText

type ParagraphElement = {
  type: 'paragraph'
  children: CustomText[]
}

type H1Element = {
  type: 'h1'
  children: CustomText[]
}

type H2Element = {
  type: 'h2'
  children: CustomText[]
}

type H3Element = {
  type: 'h3'
  children: CustomText[]
}

type ListItemElement = {
  type: 'list-item'
  children: CustomText[]
}

type CustomElement = ParagraphElement | H1Element | H2Element | H3Element | ListItemElement

declare module 'slate' {
  interface CustomTypes {
    Editor: CustomEditor & BaseEditor & ReactEditor & HistoryEditor
    Element: CustomElement
    Text: CustomText
  }
}

const MyEditor = (props: MyEditorProps) => {
  const [value, setValue] = useState<Descendant[]>(props.value)
  const editor = useMemo(() => withHistory(withReact(createEditor())), [])

  const renderElement = useCallback(
    (cbProps: any) => <Element {...cbProps} styles={props.styles} />,
    [props.styles]
  )
  const renderLeaf = useCallback(
    (cbProps: any) => <Leaf {...cbProps} styles={props.styles} />,
    [props.styles]
  )

  // set value when using history
  useEffect(() => {
    setValue(props.value)
  }, [props.value])

  // console.log('editor', editor)
  return (
    <Slate
      editor={editor}
      initialValue={value}
      onChange={(value) => {
        setValue(value)
      }}
    >
      <MyToolbar toolbarButtons={props.toolbarButtons} isFocused={props.isFocused} />
      {/* <HoveringToolbar /> */}
      <Editable
        renderElement={renderElement}
        renderLeaf={renderLeaf}
        // renderElement={props => <Element {...props} />}
        // renderLeaf={props => <Leaf {...props} />}
        placeholder="Enter some text..."
        spellCheck
        onBlur={() => {
          props.onChange(value)
        }}
      />
    </Slate>
  )
}

const clearFormat = (editor: Editor, format: string) => {
  // console.log('clear format', format)
  Transforms.setNodes(editor, { [format]: null }, { match: Text.isText, split: true })
}

const setFontProperty = (editor: Editor, key: string, value: any) => {
  Transforms.setNodes(editor, { [key]: value }, { match: Text.isText, split: true })
}

const toggleFormat = (editor: Editor, format: string) => {
  const isActive = isFormatActive(editor, format)
  Transforms.setNodes(
    editor,
    { [format]: isActive ? null : true },
    { match: Text.isText, split: true }
  )
}

const toggleBlock = (editor: Editor, format: string) => {
  const isActive = isBlockActive(editor, format)
  const isList = LIST_TYPES.includes(format)

  Transforms.unwrapNodes(editor, {
    // match: (n: Node) =>
    //   LIST_TYPES.includes(!Editor.isEditor(n) && SlateElement.isElement(n) && n.type),
    match: (n: Node) => {
      if (!Editor.isEditor(n) && SlateElement.isElement(n) && n.type) {
        return LIST_TYPES.includes(n.type)
      }
      return false
    },
    split: true
  })

  const newProperties: Partial<SlateElement> = {
    type: isActive ? 'paragraph' : isList ? 'list-item' : (format as any)
  }

  Transforms.setNodes(editor, newProperties)

  if (!isActive && isList) {
    const block = { type: format, children: [] } as CustomElement
    Transforms.wrapNodes(editor, block)
  }
}

const isFormatActive = (editor: Editor, format: string) => {
  const generator = Editor.nodes(editor, {
    match: (n: any) => n[format] === true,
    mode: 'all'
  })
  const [match] = Array.from(generator)
  return !!match
}

const isBlockActive = (editor: Editor, format: string) => {
  const generator = Editor.nodes(editor, {
    match: (n) => !Editor.isEditor(n) && SlateElement.isElement(n) && n.type === format
  })
  const [match] = Array.from(generator)
  return !!match
}

const Element = (props: any) => {
  const elementStyle = {
    lineHeight: 1, // match with mjml
    ...props.styles[props.element.type]
  }

  if (elementStyle.paddingControl === 'all') {
    delete elementStyle.paddingTop
    delete elementStyle.paddingRight
    delete elementStyle.paddingBottom
    delete elementStyle.paddingLeft
  }

  switch (props.element.type) {
    case 'h1':
      return (
        <div {...props.attributes} style={elementStyle}>
          {props.children}
        </div>
      )
    case 'h2':
      return (
        <div {...props.attributes} style={elementStyle}>
          {props.children}
        </div>
      )
    case 'h3':
      return (
        <div {...props.attributes} style={elementStyle}>
          {props.children}
        </div>
      )
    // case 'block-quote':
    //   return <blockquote {...attributes}>{children}</blockquote>
    // case 'bulleted-list':
    //   return <ul {...attributes} style={elementStyle}>{children}</ul>
    // case 'list-item':
    //   return <li {...attributes} style={elementStyle}>{children}</li>
    // case 'numbered-list':
    //   return <ol {...attributes} style={elementStyle}>{children}</ol>
    default:
      return (
        <p {...props.attributes} style={elementStyle}>
          {props.children}
        </p>
      )
  }
}

const Leaf = (props: any) => {
  let content = props.children

  // console.log('leaf', leaf)
  if (props.leaf.bold) {
    content = <strong>{content}</strong>
  }

  if (props.leaf.italic) {
    content = <em>{content}</em>
  }

  if (props.leaf.underline) {
    content = <u>{content}</u>
  }

  if (props.leaf.fontSize) {
    content = <span style={{ fontSize: props.leaf.fontSize }}>{content}</span>
  }

  if (props.leaf.fontFamily) {
    content = <span style={{ fontFamily: props.leaf.fontFamily }}>{content}</span>
  }

  if (props.leaf.fontColor) {
    content = <span style={{ color: props.leaf.fontColor }}>{content}</span>
  }

  return <span {...props.attributes}>{content}</span>
}

const fontFamilies = [
  { label: 'Arial, sans-serif', value: 'Arial, sans-serif' },
  { label: 'Verdana, sans-serif', value: 'Verdana, sans-serif' },
  { label: 'Helvetica, sans-serif', value: 'Helvetica, sans-serif' },
  { label: 'Georgia, serif', value: 'Georgia, serif' },
  { label: 'Tahoma, sans-serif', value: 'Tahoma, sans-serif' },
  { label: 'Lucida, sans-serif', value: 'Lucida, sans-serif' },
  { label: 'Trebuchet MS, sans-serif', value: 'Trebuchet MS, sans-serif' },
  { label: 'Times New Roman, serif', value: 'Times New Roman, serif' }
]

// const fontWeights = [
//   { label: <span style={{ fontWeight: 100 }}>100</span>, value: 100 },
//   { label: <span style={{ fontWeight: 200 }}>200</span>, value: 200 },
//   { label: <span style={{ fontWeight: 300 }}>300</span>, value: 300 },
//   { label: <span style={{ fontWeight: 400 }}>400</span>, value: 400 },
//   { label: <span style={{ fontWeight: 500 }}>500</span>, value: 500 },
//   { label: <span style={{ fontWeight: 600 }}>600</span>, value: 600 },
//   { label: <span style={{ fontWeight: 700 }}>700</span>, value: 700 },
//   { label: <span style={{ fontWeight: 800 }}>800</span>, value: 800 },
//   { label: <span style={{ fontWeight: 900 }}>900</span>, value: 900 },
// ]

const MyToolbar = (props: any) => {
  const ref = useRef(null)
  const editor = useSlate()

  const sizes = []

  for (var i = 6; i <= 48; i++) {
    sizes.push({ label: i + 'px', value: i + 'px' })
  }

  useEffect(() => {
    const el: any = ref.current

    if (!el) {
      return
    }

    if (!props.isFocused) {
      el.removeAttribute('style')
      return
    }

    // if (!ReactEditor.isFocused(editor)) {
    //   el.removeAttribute('style')
    //   return
    // }

    el.style.display = 'block'
  })

  // find if current selection has font size / family applied
  let fontSizeValue = undefined
  let fontFamilyValue = undefined
  let fontColorValue = undefined

  const [matchFontSize]: any = Editor.nodes(editor, {
    match: (n) => (Text.isText(n) && n.fontSize ? true : false),
    mode: 'all'
  })

  const [matchFontFamily]: any = Editor.nodes(editor, {
    match: (n) => (Text.isText(n) && n.fontFamily ? true : false),
    mode: 'all'
  })

  const [matchFontColor]: any = Editor.nodes(editor, {
    match: (n) => (Text.isText(n) && n.fontColor ? true : false),
    mode: 'all'
  })

  if (
    matchFontSize &&
    matchFontSize[0] &&
    Text.isText(matchFontSize[0]) &&
    matchFontSize[0].fontSize
  )
    fontSizeValue = matchFontSize[0].fontSize
  if (
    matchFontFamily &&
    matchFontFamily[0] &&
    Text.isText(matchFontFamily[0]) &&
    matchFontFamily[0].fontFamily
  )
    fontFamilyValue = matchFontFamily[0].fontFamily
  if (
    matchFontColor &&
    matchFontColor[0] &&
    Text.isText(matchFontColor[0]) &&
    matchFontColor[0].fontColor
  )
    fontColorValue = matchFontColor[0].fontColor

  return (
    <div ref={ref} className="rmdeditor-toolbar">
      {props.toolbarButtons.includes('bold') && (
        <FormatButton format="bold" icon={<BoldOutlined />} />
      )}
      {props.toolbarButtons.includes('italic') && (
        <FormatButton format="italic" icon={<ItalicOutlined />} />
      )}
      {props.toolbarButtons.includes('underlined') && (
        <FormatButton format="underlined" icon={<UnderlineOutlined />} />
      )}
      {props.toolbarButtons.includes('h1') && (
        <BlockButton format="h1" icon={<span style={{ fontSize: '15px' }}>H1</span>} />
      )}
      {props.toolbarButtons.includes('h2') && (
        <BlockButton format="h2" icon={<span style={{ fontSize: '15px' }}>H2</span>} />
      )}
      {props.toolbarButtons.includes('h3') && (
        <BlockButton format="h3" icon={<span style={{ fontSize: '15px' }}>H3</span>} />
      )}
      {props.toolbarButtons.includes('fonts') && (
        <>
          <Select
            style={{ width: '120px' }}
            placeholder="Font size"
            dropdownMatchSelectWidth={true}
            defaultActiveFirstOption={false}
            autoFocus={false}
            allowClear={true}
            value={fontSizeValue}
            onMouseDown={() => {
              // save last known selection
              if (!editor.savedSelection) editor.savedSelection = editor.selection
              // console.log('down')
            }}
            onBlur={() => {
              // reset selection on exit
              editor.savedSelection = null
            }}
            onClear={() => {
              // use editor.selection instead of editor.savedSelection because
              // onClear is triggered before onMouseDown
              if (editor.selection) {
                Transforms.select(editor, editor.selection)
                clearFormat(editor, 'fontSize')
                editor.savedSelection = null
                ReactEditor.focus(editor)
              }
            }}
            onChange={(val) => {
              if (!val) return // abort onClear
              // console.log('val', val)
              if (editor.savedSelection) {
                // console.log('editor.savedSelection', editor.savedSelection)
                Transforms.select(editor, editor.savedSelection)
                if (!val) clearFormat(editor, 'fontSize')
                else setFontProperty(editor, 'fontSize', val)
                editor.savedSelection = null
                ReactEditor.focus(editor)
              } else {
                // console.log('no selection')
              }
            }}
            size="small"
            options={sizes}
          />
          <Select
            style={{ width: '150px' }}
            placeholder="Font family"
            dropdownMatchSelectWidth={false}
            defaultActiveFirstOption={false}
            autoFocus={false}
            allowClear={true}
            value={fontFamilyValue}
            onMouseDown={() => {
              // save last known selection
              if (!editor.savedSelection) editor.savedSelection = editor.selection
            }}
            onClear={() => {
              if (editor.selection) {
                Transforms.select(editor, editor.selection)
                clearFormat(editor, 'fontFamily')
                editor.savedSelection = null
                ReactEditor.focus(editor)
              }
            }}
            onBlur={() => {
              // reset selection on exit
              editor.savedSelection = null
            }}
            onChange={(val) => {
              if (!val) return // abort onClear
              // console.log('change', val)
              if (editor.savedSelection) {
                Transforms.select(editor, editor.savedSelection)
                setFontProperty(editor, 'fontFamily', val)
                editor.savedSelection = null
                ReactEditor.focus(editor)
              }
            }}
            size="small"
            options={fontFamilies}
          />
          <ColorPickerLight
            className="rmdeditor-toolbar-color"
            size="small"
            value={fontColorValue}
            onMouseDown={() => {
              // save last known selection
              if (!editor.savedSelection) editor.savedSelection = editor.selection
            }}
            onBlur={() => {
              // reset selection on exit
              editor.savedSelection = null
            }}
            onChange={(newColor) => {
              // console.log('newColor', newColor)
              if (editor.savedSelection) {
                Transforms.select(editor, editor.savedSelection)
                setFontProperty(editor, 'fontColor', newColor)
                // console.log('newColor', newColor)
                ReactEditor.focus(editor)
              }
            }}
          />
        </>
      )}
      <div className="rmdeditor-toolbar-overlay"></div>
    </div>
  )
}

const FormatButton = (props: { format: string; icon: React.ReactNode }) => {
  const editor = useSlate()
  return (
    <Button
      className="rmdeditor-toolbar-button"
      reversed
      active={isFormatActive(editor, props.format)}
      onMouseDown={(event: any) => {
        event.preventDefault()
        toggleFormat(editor, props.format)
      }}
    >
      {props.icon}
    </Button>
  )
}

const BlockButton = (props: { format: string; icon: React.ReactNode }) => {
  const editor = useSlate()

  return (
    <Button
      className="rmdeditor-toolbar-button"
      reversed
      active={isBlockActive(editor, props.format)}
      onMouseDown={(event: any) => {
        event.preventDefault()
        toggleBlock(editor, props.format)
      }}
    >
      {props.icon}
    </Button>
  )
}

const Button = React.forwardRef(
  (
    {
      className,
      active,
      reversed,
      ...props
    }: PropsWithChildren<
      {
        active: boolean
        reversed: boolean
      } & BaseProps
    >,
    ref: any
    // ref: Ref<OrNull<HTMLSpanElement>>
  ) => <div {...props} ref={ref} className={cn('rmdeditor-toolbar-button', { active: active })} />
)

export default MyEditor

const renderElementInReact = (key: any, element: any, styles: any) => {
  let children = element.children
    ? element.children.map((child: any, k: number) => renderElementInReact(k, child, styles))
    : element.text || ''

  // is root elemnt
  if (element.type) {
    const elementStyle = {
      lineHeight: 1, // match with mjml automatic line height
      ...styles[element.type]
    }

    if (elementStyle.paddingControl === 'all') {
      delete elementStyle.paddingTop
      delete elementStyle.paddingRight
      delete elementStyle.paddingBottom
      delete elementStyle.paddingLeft
    }

    switch (element.type) {
      // case 'block-quote':
      //   return <blockquote {...attributes}>{children}</blockquote>
      case 'bulleted-list':
        return (
          <ul key={key} style={elementStyle}>
            {children}
          </ul>
        )
      case 'h1':
        return (
          <div key={key} style={elementStyle}>
            {children}
          </div>
        )
      case 'h2':
        return (
          <div key={key} style={elementStyle}>
            {children}
          </div>
        )
      case 'h3':
        return (
          <div key={key} style={elementStyle}>
            {children}
          </div>
        )
      case 'list-item':
        return (
          <li key={key} style={elementStyle}>
            {children}
          </li>
        )
      case 'numbered-list':
        return (
          <ol key={key} style={elementStyle}>
            {children}
          </ol>
        )
      default:
        return (
          <p key={key} style={elementStyle}>
            {children}
          </p>
        )
    }
  } else {
    // is text leaf
    if (element.bold) {
      children = <strong key={key}>{children}</strong>
    }

    if (element.italic) {
      children = <em key={key}>{children}</em>
    }

    if (element.underlined) {
      children = <u key={key}>{children}</u>
    }

    if (element.fontSize) {
      children = (
        <span style={{ fontSize: element.fontSize }} key={key}>
          {children}
        </span>
      )
    }

    if (element.fontFamily) {
      children = (
        <span style={{ fontFamily: element.fontFamily }} key={key}>
          {children}
        </span>
      )
    }

    if (element.fontColor) {
      children = (
        <span style={{ color: element.fontColor }} key={key}>
          {children}
        </span>
      )
    }

    return <span key={key}>{children}</span>
  }
}

export const EditorDataToReact = (data: any, styles: any) => {
  return data.map((el: any, i: number) => <div key={i}>{renderElementInReact(0, el, styles)}</div>)
}
