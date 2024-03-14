import React, { Component } from 'react'
import { Button } from 'antd'
// https://github.com/ajaxorg/ace-builds/tree/master/src-noconflict
import AceEditor from 'react-ace'
// import 'ace-builds/src-noconflict/mode-nunjucks'
import 'ace-builds/src-noconflict/mode-javascript'
import 'ace-builds/src-noconflict/mode-json'
import 'ace-builds/src-noconflict/theme-chrome'
import 'ace-builds/src-noconflict/ext-language_tools'

type AceInputProps = {
  value?: string
  onChange?: (value: string) => void
  mode: string
  id: string
  height: string
  width: string
}

const AceInput = (props: AceInputProps) => {
  return (
    <AceEditor
      defaultValue={props.value}
      mode={props.mode}
      theme="chrome"
      onChange={props.onChange}
      debounceChangePeriod={300}
      name={props.id}
      editorProps={{ $blockScrolling: true }}
      fontSize="14px"
      height={props.height}
      width={props.width}
      className="ace-input"
      wrapEnabled={true}
      setOptions={
        {
          // enableBasicAutocompletion: true,
          // enableLiveAutocompletion: true,
          // enableSnippets: true
        }
      }
    />
  )
}

export default AceInput
