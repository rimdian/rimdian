import {
  useState,
  MouseEvent,
  ReactNode,
  RefObject,
  useEffect,
  createContext,
  useContext
} from 'react'
import { BlockEditorRenderer } from './BlockEditorRenderer'
import { DropResult } from './smooth-dnd'
import Container from './Container'
import Draggable from './Draggable'
import uuid from 'short-uuid'
import { BlockDefinitionInterface, BlockInterface, BlockDefinitionMap } from './Block'
import { cloneDeep, get, isEqual, remove, set } from 'lodash'
import './UI/editor.css'

const EditorContext = createContext<EditorContextValue | null>(null)

export function useEditorContext(): EditorContextValue {
  const editorValue = useContext(EditorContext)
  if (!editorValue) {
    throw new Error('Missing EditorContextProvider in its parent.')
  }
  return editorValue
}

export interface EditorContextValue {
  blockDefinitions: BlockDefinitionMap
  savedBlocks: BlockDefinitionInterface[]
  templateDataValue: string
  currentTree: BlockInterface
  selectedBlockId: string
  updateTree: (path: string, data: any) => void
  selectBlock: (block: BlockInterface, event: MouseEvent) => void
  renderBlockForMenu: (blockDefinition: BlockDefinitionInterface) => ReactNode
  renderSavedBlockForMenu: (block: BlockInterface, renderMenu: ReactNode) => ReactNode
  editor: ReactNode
  history: BlockInterface[]
  currentHistoryIndex: number
  setCurrentHistoryIndex: (index: number) => void
  deviceWidth: number
  setDeviceWidth: (width: number) => void
}

export interface SelectedBlockButtonsProp {
  isDraggable: boolean
  blockDefinitions: BlockDefinitionMap
  block: BlockInterface
  cloneBlock: (block: BlockInterface) => void
  deleteBlock: (block: BlockInterface) => void
}

export interface EditorProps {
  children: ReactNode
  blockDefinitions: BlockDefinitionMap
  savedBlocks: BlockDefinitionInterface[]
  templateDataValue: string
  value: BlockInterface
  onChange: (newValue: BlockInterface) => void
  renderSelectedBlockButtons: (props: SelectedBlockButtonsProp) => ReactNode
  deviceWidth: number
  selectedBlockId?: string
}

// recursive id generation, used to clone blocks
export const generateNewBlockIds = (block: BlockInterface) => {
  block.id = uuid.generate()
  // generate new uuids for children
  if (block.children) {
    block.children.forEach((child) => {
      generateNewBlockIds(child)
    })
  }
}

export const Editor = (props: EditorProps): JSX.Element => {
  // const { color, onClick } = props;

  let focusedNode: any = undefined

  useEffect(() => {
    return () => {
      // reset focused node on cleanup
      // eslint-disable-next-line react-hooks/exhaustive-deps
      if (focusedNode && focusedNode.current)
        focusedNode.current.classList.remove('rmdeditor-focused')
    }
  })

  const recomputeBlockpaths = (block: BlockInterface) => {
    if (block.children) {
      block.children.forEach((child, i) => {
        child.path = block.path + '.children[' + i + ']'
        recomputeBlockpaths(child)
      })
    }
  }

  // set initial block paths from provided tree
  recomputeBlockpaths(props.value)

  const [history, setHistory] = useState([props.value])
  const [currentHistoryIndex, setCurrentHistoryIndex] = useState(0)
  const [tree, setTree] = useState(props.value)
  const [selectedBlockId, setSelectedBlockId] = useState(
    props.selectedBlockId ? props.selectedBlockId : tree.id
  )
  // const [focusedBlock, setFocusedBlock] = useState<RefObject<HTMLDivElement> | undefined>()
  const [deviceWidth, setDeviceWidth] = useState<number>(props.deviceWidth)
  // const [isDragging, setIsDragging] = useState(false)

  const onContainerDrop = (path: string, dropResult: DropResult) => {
    // console.log('drop path', path)
    // console.log('dropResult', dropResult)
    // console.log('dropped ' + dropResult.payload.kind + ':' + dropResult.payload.id + ' into group ' + group)

    const { removedIndex, addedIndex, payload } = dropResult

    // abort if nothing
    if (removedIndex === null && addedIndex === null) return

    // get children at path
    const finalPath = path === '' ? 'children' : path + '.children'
    // console.log('finalPath',finalPath)

    const currentTree = history[currentHistoryIndex]

    const children = get(currentTree, finalPath)
    if (!children) return

    const result = [...children]

    let itemToAdd = payload

    if (removedIndex !== null) {
      itemToAdd = result.splice(removedIndex, 1)[0]
    }

    if (addedIndex !== null) {
      result.splice(addedIndex, 0, itemToAdd)
    }

    // abort if no changes
    if (isEqual(children, result) === true) {
      return
    }

    updateTree(finalPath, result)
    selectBlock(payload)
  }

  const selectBlock = (block: BlockInterface, event?: MouseEvent) => {
    // console.log('selectBlock', block)

    if (event) {
      event.preventDefault()
      event.stopPropagation()
    }

    // console.log('click event', event)
    if (selectedBlockId !== block.id) {
      setSelectedBlockId(block.id)
      if (block.kind === 'root') {
        onFocusBlock(undefined)
      }
    }
  }

  const updateTree = (path: string, data: any) => {
    const currentTree = history[currentHistoryIndex]

    let newTree: BlockInterface

    if (path === '') {
      newTree = cloneDeep(data) as BlockInterface
    } else {
      newTree = cloneDeep(currentTree)
      set(newTree, path, data)
    }

    // newTree.lastUpdate = uuid.generate()
    setTree(newTree)
    props.onChange(newTree)

    // append to history
    const newHistory = [...history]
    newHistory.push(newTree)
    setHistory(newHistory)
    // move cursor to last version
    setCurrentHistoryIndex(newHistory.length - 1)
    // console.log('history is', newHistory)
    // console.log('new tree', JSON.stringify(newTree, undefined, 2))
  }

  const getParentBlock = (tree: BlockInterface, block: BlockInterface) => {
    const parts = block.path.split('.')

    // console.log('block.path', block.path)
    // console.log('parts', parts)

    // get parent block path
    let parentBlock = tree
    let parentPath = ''

    // find parent block
    parts.forEach((part, i) => {
      // traverse tree as long as we dont reach the last block
      if (i < parts.length - 1) {
        // console.log('qsd', parentPath + (i === 0 ? '' : '.') + part)
        parentBlock = get(tree, parentPath + (i === 0 ? '' : '.') + part)
        parentPath = parentBlock.path
      }
    })

    return parentBlock
  }

  const deleteBlock = (block: BlockInterface) => {
    if (!props.blockDefinitions[block.kind].isDeletable) {
      alert('The block ' + block.kind + ' is not deletable')
      return
    }

    const currentTree = history[currentHistoryIndex]

    const newTree = cloneDeep(currentTree)

    const parentBlock = getParentBlock(newTree, block)

    parentBlock.children = remove(parentBlock.children, (child) => child.id !== block.id)

    recomputeBlockpaths(parentBlock)

    updateTree(
      parentBlock.path + (parentBlock.path === '' ? '' : '.') + 'children',
      parentBlock.children
    )
  }

  const cloneBlock = (block: BlockInterface) => {
    const currentTree = history[currentHistoryIndex]

    const newTree = cloneDeep(currentTree)

    const parentBlock = getParentBlock(newTree, block)

    // console.log('parentBlock', JSON.stringify(parentBlock, undefined, 2))

    if (!parentBlock.children) {
      parentBlock.children = []
    }

    const newBlock = cloneDeep(block)

    // append after block
    const currentBlockIndex = parentBlock.children.findIndex((child) => child.id === block.id)
    const newBlockIndex = currentBlockIndex + 1

    newBlock.path =
      parentBlock.path + (parentBlock.path === '' ? '' : '.') + 'children[' + newBlockIndex + ']'
    generateNewBlockIds(newBlock)

    // console.log('newBlock', JSON.stringify(newBlock, undefined, 2))
    parentBlock.children.splice(newBlockIndex, 0, newBlock)
    recomputeBlockpaths(parentBlock)

    updateTree(parentBlock.path, parentBlock)
  }

  const generateBlockFromDefinition = (
    blockDefinition: BlockDefinitionInterface
  ): BlockInterface => {
    const id = uuid.generate()

    const block: BlockInterface = {
      id: id,
      kind: blockDefinition.kind,
      path: '', // path is set when rendering
      children: blockDefinition.children
        ? blockDefinition.children.map((child) => {
            return generateBlockFromDefinition(child)
          })
        : [],
      data: { ...blockDefinition.defaultData }
    }

    return block
  }

  const renderBlockForMenu = (blockDefinition: BlockDefinitionInterface) => {
    return (
      <Container
        key={blockDefinition.kind}
        groupName={blockDefinition.draggableIntoGroup}
        behaviour="copy"
        getChildPayload={generateBlockFromDefinition.bind(null, blockDefinition)}
      >
        <Draggable>
          {blockDefinition.renderMenu
            ? blockDefinition.renderMenu(blockDefinition)
            : 'renderMenu() not provided for: ' + blockDefinition.kind}
        </Draggable>
      </Container>
    )
  }

  const renderSavedBlockForMenu = (block: BlockInterface, renderMenu: ReactNode): ReactNode => {
    // find definition of block
    if (!props.blockDefinitions[block.kind]) {
      console.error('block definition not found for block', block)
      return ''
    }

    return (
      <Container
        key={block.id}
        groupName={props.blockDefinitions[block.kind].draggableIntoGroup}
        behaviour="copy"
        getChildPayload={() => {
          generateNewBlockIds(block)
          return block
        }}
      >
        <Draggable>{renderMenu}</Draggable>
      </Container>
    )
  }

  const onFocusBlock = (node: RefObject<HTMLDivElement> | undefined) => {
    const previousNode = focusedNode?.current
    const currentNode = node?.current

    // abort if the focus is on same block
    if (previousNode && currentNode && previousNode.id === currentNode.id) {
      return
    }

    // remove previous CSS if possible
    if (previousNode) {
      previousNode.classList.remove('rmdeditor-focused')
    }

    if (currentNode) {
      currentNode.classList.add('rmdeditor-focused')
    }

    focusedNode = node
  }

  // console.log('render')

  if (props.value.kind !== 'root') {
    return <>First block should be "root", got: {props.value.kind}</>
  }

  const currentTree = history[currentHistoryIndex]

  const blockEditorRendererProps = {
    block: currentTree,
    blockDefinitions: props.blockDefinitions,
    onContainerDrop: onContainerDrop,
    onSelectBlock: selectBlock,
    selectedBlockId: selectedBlockId,
    onFocusBlock: onFocusBlock,
    // focusedBlock: focusedBlock,
    renderSelectedBlockButtons: props.renderSelectedBlockButtons,
    deleteBlock: deleteBlock,
    cloneBlock: cloneBlock,
    updateTree: updateTree,
    tree: tree,
    deviceWidth: deviceWidth
  }

  const layoutProps: EditorContextValue = {
    blockDefinitions: props.blockDefinitions,
    savedBlocks: props.savedBlocks,
    templateDataValue: props.templateDataValue,
    currentTree: currentTree,
    selectedBlockId: selectedBlockId,
    updateTree: updateTree,
    selectBlock: selectBlock,
    renderBlockForMenu: renderBlockForMenu,
    renderSavedBlockForMenu: renderSavedBlockForMenu,
    editor: (
      <div
        onMouseLeave={() => {
          // remove focus when leaving the editor
          if (focusedNode) {
            onFocusBlock(undefined)
          }
        }}
      >
        <BlockEditorRenderer {...blockEditorRendererProps} />
      </div>
    ),
    history: history,
    currentHistoryIndex: currentHistoryIndex,
    setCurrentHistoryIndex: setCurrentHistoryIndex,
    deviceWidth: deviceWidth,
    setDeviceWidth: setDeviceWidth
  }

  return <EditorContext.Provider value={layoutProps}>{props.children}</EditorContext.Provider>
}
