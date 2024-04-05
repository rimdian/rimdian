import { useState } from 'react'
import { Button, Tooltip, Space, Form, Select } from 'antd'
import _ from 'lodash'
import { useEditorContext, EditorContextValue } from '../Editor'
import Settings from './Settings'
import { Blocks, BlocksProps } from './Blocks'
import { BlockInterface } from '../Block'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faDesktop,
  faMobileAlt,
  faEye,
  faChevronLeft,
  faChevronRight
} from '@fortawesome/free-solid-svg-icons'
import Preview from './Preview'
import CSS from 'utils/css'
import AceInput from 'components/common/input_ace'

export const MobileWidth = 400
export const DesktopWidth = 960

const FindBlockById = (currentBlock: BlockInterface, id: string): BlockInterface | undefined => {
  if (currentBlock.id === id) return currentBlock
  else if (currentBlock.children) {
    let found
    currentBlock.children.forEach((child) => {
      const got = FindBlockById(child, id)
      if (got) found = got
    })
    return found
  }
  return undefined
}

export const Layout = (props: any): JSX.Element => {
  const editor: EditorContextValue = useEditorContext()
  const [isPreview, setIsPreview] = useState(false)

  // console.log('render')

  if (editor.currentTree.kind !== 'root') {
    return <>First block should be "root", got: {editor.currentTree.kind}</>
  }

  const blocksProps: BlocksProps = {
    blockDefinitions: editor.blockDefinitions,
    savedBlocks: editor.savedBlocks,
    renderBlockForMenu: editor.renderBlockForMenu,
    renderSavedBlockForMenu: editor.renderSavedBlockForMenu
  }

  const pathBlocks: BlockInterface[] = [editor.currentTree]

  let currentPath = ''

  let selectedBlock = FindBlockById(editor.currentTree, editor.selectedBlockId)

  // focus root by default
  if (!selectedBlock) {
    selectedBlock = editor.currentTree
  }

  selectedBlock.path.split('.').forEach((part) => {
    if (currentPath === '') {
      currentPath = part
    } else {
      currentPath += '.' + part
    }

    const block: BlockInterface = _.get(editor.currentTree, currentPath)

    if (block && block.kind) {
      pathBlocks.push(block)
    }
  })

  const togglePreview = () => {
    setIsPreview(!isPreview)
  }

  const toggleDevice = () => {
    if (editor.deviceWidth === MobileWidth) {
      editor.setDeviceWidth(DesktopWidth)
    } else {
      editor.setDeviceWidth(MobileWidth)
    }
  }

  const goBackHistory = () => {
    const lastHistoryIndex: number = editor.history.length - 1
    if (lastHistoryIndex > 0) {
      // console.log('back to', editor.currentHistoryIndex - 1)
      editor.setCurrentHistoryIndex(editor.currentHistoryIndex - 1)
    }
  }

  const goNextHistory = () => {
    const lastHistoryIndex: number = editor.history.length - 1
    if (editor.currentHistoryIndex < lastHistoryIndex) {
      editor.setCurrentHistoryIndex(editor.currentHistoryIndex + 1)
    }
  }
  const doc = document.querySelector('.rmdeditor-main')
  const layoutLeftHeight = doc ? parseInt(window.getComputedStyle(doc).height) - 132 : 400

  // console.log('layout props', props)

  return (
    <div className="rmdeditor-main" style={{ height: props.height || '100vh' }}>
      <div className={'rmdeditor-layout-left ' + (isPreview ? 'preview' : '')}>
        {!isPreview && <Blocks {...blocksProps} />}
        {isPreview && (
          <>
            <Form.Item
              label="Use a macros page"
              name="template_macro_id"
              className={CSS.padding_t_m + ' ' + CSS.padding_h_m}
            >
              <Select
                style={{ width: '100%' }}
                dropdownMatchSelectWidth={false}
                allowClear={true}
                size="small"
                placeholder="Select macros page"
                options={props.macros.map((x: any) => {
                  return { label: x.name, value: x.id }
                })}
                onChange={(val: any) => props.form.setFieldsValue({ template_macro_id: val })}
              />
            </Form.Item>

            <Form.Item
              label={<span className={CSS.padding_h_m}>Test data</span>}
              name="test_data"
              validateFirst={true}
              rules={[
                {
                  validator: (xxx, value) => {
                    // check if data is valid json
                    try {
                      if (JSON.parse(value)) {
                      }
                      return Promise.resolve(undefined)
                    } catch (e: any) {
                      return Promise.reject('Your test variables is not a valid JSON object!')
                    }
                  }
                },
                {
                  required: false,
                  type: 'object',
                  transform: (value: any) => {
                    try {
                      const parsed = JSON.parse(value)
                      return parsed
                    } catch (e: any) {
                      return value
                    }
                  }
                }
              ]}
            >
              <AceInput
                onChange={(val: any) => props.form.setFieldsValue({ test_data: val })}
                id="test_data"
                width="100%"
                height={layoutLeftHeight + 'px'}
                mode="json"
                theme="monokai"
              />
            </Form.Item>
          </>
        )}
      </div>

      {isPreview && (
        <>
          <Preview
            tree={editor.currentTree}
            templateData={editor.templateDataValue}
            isMobile={editor.deviceWidth === MobileWidth}
            deviceWidth={editor.deviceWidth}
            toggleDevice={toggleDevice}
            urlParams={editor.urlParams}
            closePreview={togglePreview}
          />
        </>
      )}

      {!isPreview && (
        <div className="rmdeditor-layout-middle">
          <div className="rmdeditor-topbar">
            <span className={CSS.pull_right}>
              <Space>
                <Button.Group>
                  <Button
                    size="small"
                    type="text"
                    disabled={editor.deviceWidth === MobileWidth}
                    onClick={() => toggleDevice()}
                  >
                    <FontAwesomeIcon icon={faMobileAlt} />
                  </Button>
                  <Button
                    size="small"
                    type="text"
                    disabled={editor.deviceWidth === DesktopWidth}
                    onClick={() => toggleDevice()}
                  >
                    <FontAwesomeIcon icon={faDesktop} />
                  </Button>
                </Button.Group>

                <Button type="primary" size="small" ghost onClick={() => togglePreview()}>
                  <FontAwesomeIcon icon={faEye} />
                  &nbsp; Preview
                </Button>
              </Space>
            </span>

            <Space size="large">
              <>
                <Button.Group>
                  <Tooltip title="Undo">
                    <Button
                      size="small"
                      type="text"
                      onClick={goBackHistory}
                      disabled={editor.currentHistoryIndex === 0}
                      icon={<FontAwesomeIcon icon={faChevronLeft} />}
                    />
                  </Tooltip>
                  <Tooltip title="Redo">
                    <Button
                      size="small"
                      type="text"
                      onClick={goNextHistory}
                      disabled={editor.currentHistoryIndex === editor.history.length - 1}
                      icon={<FontAwesomeIcon icon={faChevronRight} />}
                    />
                  </Tooltip>
                </Button.Group>
                <div className="rmdeditor-path">
                  {pathBlocks.map((block, i) => {
                    const isLast = i === pathBlocks.length - 1 ? true : false
                    return (
                      <span key={i}>
                        {isLast === true && (
                          <span className="rmdeditor-path-item-last">
                            {editor.blockDefinitions[block.kind]?.name}
                          </span>
                        )}
                        {isLast === false && (
                          <>
                            <span
                              className="rmdeditor-path-item"
                              onClick={editor.selectBlock.bind(null, block)}
                            >
                              {editor.blockDefinitions[block.kind]?.name}
                            </span>
                            <span className="rmdeditor-path-divider">/</span>
                          </>
                        )}
                      </span>
                    )
                  })}
                </div>
              </>
            </Space>
          </div>
          <div onClick={editor.selectBlock.bind(null, editor.currentTree)}>{editor.editor}</div>
        </div>
      )}

      <div className={'rmdeditor-layout-right ' + (isPreview ? 'preview' : '')}>
        {!isPreview && (
          <Settings
            block={selectedBlock}
            blockDefinition={editor.blockDefinitions[selectedBlock.kind]}
            tree={editor.currentTree}
            updateTree={editor.updateTree}
            // deviceType={editor.deviceWidth <= 480 ? 'mobile' : 'desktop'}
          />
        )}
      </div>
    </div>
  )
}
