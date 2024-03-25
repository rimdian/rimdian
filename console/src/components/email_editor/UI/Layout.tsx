import React, { useState } from "react"
import { Row, Col, Button, Tooltip } from 'antd'
import _ from 'lodash';
import { useEditorContext, EditorContextValue } from '../Editor'
import Settings from './Settings'
import { Blocks, BlocksProps } from './Blocks'
import { BlockInterface } from '../Block'
import { UndoOutlined, RedoOutlined, ExpandAltOutlined, ShrinkOutlined } from '@ant-design/icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faDesktop, faMobileAlt, faEye, faPen } from '@fortawesome/free-solid-svg-icons'
import Preview from './Preview'
import cn from 'classnames'


// export type DeviceType = 'mobile' | 'desktop'

export const MobileWidth = 400
export const DesktopWidth = 960

const FindBlockById = (currentBlock: BlockInterface, id: string): BlockInterface | undefined => {

    if (currentBlock.id === id) return currentBlock
    else if (currentBlock.children) {
        let found
        currentBlock.children.forEach(child => {
            const got = FindBlockById(child, id)
            if (got) found = got
        })
        return found
    }
    return undefined
}

export const Layout = (props: any): JSX.Element => {

    const editor: EditorContextValue = useEditorContext();
    const [modalPreview, setModalPreview] = useState(false)
    const [fullscreen, setFullscreen] = useState(false)

    // console.log('render')

    if (editor.currentTree.kind !== 'root') {
        return <>First block should be "root", got: {editor.currentTree.kind}</>
    }

    const blocksProps: BlocksProps = {
        blockDefinitions: editor.blockDefinitions,
        savedBlocks: editor.savedBlocks,
        renderBlockForMenu: editor.renderBlockForMenu,
        renderSavedBlockForMenu: editor.renderSavedBlockForMenu,
    }

    const pathBlocks: BlockInterface[] = [editor.currentTree]

    let currentPath = ''

    let selectedBlock = FindBlockById(editor.currentTree, editor.selectedBlockId)

    // focus root by default
    if (!selectedBlock) {
        selectedBlock = editor.currentTree
    }

    selectedBlock.path.split('.').forEach(part => {
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
        setModalPreview(!modalPreview)
    }

    const toggleFullscreen = () => {
        setFullscreen(!fullscreen)
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

    // console.log('layout props', props)

    return <div className={cn('cmeditor-main', {fullscreen})} style={{height: props.height || '100vh'}}>
        <div className="cmeditor-topbar">
            <Row>
                <Col span={6}><div className="cmeditor-title"></div></Col>
                <Col span={10} className="cmeditor-path">
                    {modalPreview === false && pathBlocks.map((block, i) => {
                        const isLast = i === (pathBlocks.length - 1) ? true : false
                        return <span key={i}>
                            {isLast === true && <span className="cmeditor-path-item-last">{editor.blockDefinitions[block.kind]?.name}</span>}
                            {isLast === false && <>
                                <span className="cmeditor-path-item" onClick={editor.selectBlock.bind(null, block)}>
                                    {editor.blockDefinitions[block.kind]?.name}
                                </span>
                                <span className="cmeditor-path-divider">/</span>
                            </>}
                        </span>
                    })}
                </Col>
                <Col span={2}>
                    {editor.history.length > 1 && <Button.Group>
                        <Tooltip title="Undo"><Button ghost onClick={goBackHistory} disabled={editor.currentHistoryIndex === 0} icon={<UndoOutlined />} /></Tooltip>
                        <Tooltip title="Redo"><Button ghost onClick={goNextHistory} disabled={editor.currentHistoryIndex === (editor.history.length - 1)} icon={<RedoOutlined />} /></Tooltip>
                    </Button.Group>}
                </Col>
                <Col span={2} style={{ textAlign: 'left' }}>
                    <Button.Group>
                        <Button type={editor.deviceWidth === MobileWidth ? 'primary' : 'default'} ghost onClick={() => editor.setDeviceWidth(MobileWidth)}>
                            <FontAwesomeIcon icon={faMobileAlt} />
                        </Button>
                        <Button type={editor.deviceWidth === DesktopWidth ? 'primary' : 'default'} ghost onClick={() => editor.setDeviceWidth(DesktopWidth)}>
                            <FontAwesomeIcon icon={faDesktop} />
                        </Button>
                    </Button.Group>

                    {/* <Radio.Group onChange={(e) => editor.setDeviceWidth(e.target.value)} buttonStyle="solid" optionType="button" value={editor.deviceWidth}>
                        <Radio.Button value={400}><MobileOutlined style={{verticalAlign: 'middle'}} /></Radio.Button>
                        <Radio.Button value={960}><DesktopOutlined style={{verticalAlign: 'middle'}} /></Radio.Button>
                    </Radio.Group> */}
                </Col>
                <Col span={4} style={{ textAlign: 'right' }}>
                    <Button type="primary" onClick={() => togglePreview()}>
                        {modalPreview && <><FontAwesomeIcon icon={faPen} />&nbsp; Edit</>}
                        {!modalPreview && <><FontAwesomeIcon icon={faEye} />&nbsp; Preview</>}
                    </Button>

                    <Button className="margin-l-s" type="default" ghost onClick={() => toggleFullscreen()} icon={fullscreen ? <ShrinkOutlined /> : <ExpandAltOutlined />} />
                </Col>
            </Row>
        </div>


        {modalPreview === true && <Preview
            tree={editor.currentTree}
            templateData={editor.templateData}
            isMobile={editor.deviceWidth === MobileWidth}
            form={props.form}
            macros={props.macros}
        />}

        {modalPreview === false && <>
            <div className="cmeditor-layout-left">
                <Blocks {...blocksProps} />
            </div>

            <div className="cmeditor-layout-middle" onClick={editor.selectBlock.bind(null, editor.currentTree)}>
                {/* <div className="cmeditor-page" style={{ maxWidth: editor.deviceWidth + 'px' }}> */}
                {editor.editor}
                {/* </div> */}
            </div>

            <div className="cmeditor-layout-right">
                <Settings
                    block={selectedBlock}
                    blockDefinition={editor.blockDefinitions[selectedBlock.kind]}
                    tree={editor.currentTree}
                    updateTree={editor.updateTree}
                // deviceType={editor.deviceWidth <= 480 ? 'mobile' : 'desktop'}
                />
            </div>
        </>}
    </div>
}
