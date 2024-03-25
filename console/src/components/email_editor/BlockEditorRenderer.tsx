import { CSSProperties, MouseEvent, ReactNode, useRef, RefObject } from 'react'
import { DropResult } from './smooth-dnd'
import Container from './Container'
import Draggable from './Draggable'
import { BlockInterface, BlockDefinitionMap, BlockDefinitionInterface } from './Block'
import { SelectedBlockButtonsProp } from './Editor'
import cn from 'classnames'

export interface BlockEditorRendererProps {
  block: BlockInterface
  blockDefinitions: BlockDefinitionMap
  onContainerDrop: (path: string, result: DropResult) => void
  onSelectBlock: (block: BlockInterface, e: MouseEvent) => void
  selectedBlockId: string
  // focusedBlock: BlockInterface | undefined
  // onFocusBlock: (block: BlockInterface | undefined) => void
  onFocusBlock: (node: RefObject<HTMLDivElement> | undefined) => void
  renderSelectedBlockButtons: (props: SelectedBlockButtonsProp) => ReactNode
  deleteBlock: (block: BlockInterface) => void
  cloneBlock: (block: BlockInterface) => void
  updateTree: (path: string, data: any) => void
  tree: BlockInterface
  deviceWidth: number
}

const RenderChild = (props: BlockEditorRendererProps, index: number) => {
  const childBlock = props.block.children[index]

  // set child path
  childBlock.path =
    props.block.path + (props.block.path !== '' ? '.' : '') + 'children[' + index + ']'

  const childBlockProps: BlockEditorRendererProps = {
    block: childBlock,
    blockDefinitions: props.blockDefinitions,
    onContainerDrop: props.onContainerDrop,
    onSelectBlock: props.onSelectBlock,
    selectedBlockId: props.selectedBlockId,
    // focusedBlock: props.focusedBlock,
    onFocusBlock: props.onFocusBlock,
    renderSelectedBlockButtons: props.renderSelectedBlockButtons,
    deleteBlock: props.deleteBlock,
    cloneBlock: props.cloneBlock,
    updateTree: props.updateTree,
    tree: props.tree,
    deviceWidth: props.deviceWidth
  }

  return <BlockEditorRenderer key={childBlock.id} {...childBlockProps} />
}

const RenderChildrenVertically = (
  props: BlockEditorRendererProps,
  blockDefinition: BlockDefinitionInterface
) => {
  // render empty children if provided by block
  // if ((!props.block.children || props.block.children.length === 0) && blockDefinition.renderEditorEmptyChildren) {
  //     return <>{blockDefinition.renderEditorEmptyChildren(props)}</>
  // }

  if (blockDefinition.containsDraggables === true) {
    return (
      <>
        {props.block.children?.map((childBlock, i) => {
          return <Draggable key={childBlock.id}>{RenderChild(props, i)}</Draggable>
        })}
      </>
    )
  }

  // wrap array in JSX Fragment required
  return (
    <>
      {props.block.children?.map((childBlock, i) => {
        return <div key={childBlock.id}>{RenderChild(props, i)}</div>
      })}
    </>
  )
}

// const stopEvent = (event: MouseEvent) => {
//     console.log('click drag')
//     event.preventDefault()
//     event.stopPropagation()
// }

// React.memo(
export const BlockEditorRenderer = (props: BlockEditorRendererProps) => {
  const blockElement = useRef<HTMLDivElement>(null)

  // console.log('render', props)
  const blockDefinition = props.blockDefinitions[props.block.kind]
  if (!blockDefinition) {
    return (
      <div className="cmeditor-block-content">Block definition missing for {props.block.kind}</div>
    )
  }
  if (!blockDefinition.renderEditor) {
    return (
      <div className="cmeditor-block-content">
        Block editor render missing for {props.block.kind}
      </div>
    )
  }

  let childContent = <></>

  const orientation =
    blockDefinition.columns && blockDefinition.columns.length > 1 ? 'horizontal' : 'vertical'

  // horizontal columns = container with draggables inside
  if (orientation === 'horizontal') {
    if (!blockDefinition.columns || blockDefinition.columns.length === 0) {
      return <span>No columns configured for block: {blockDefinition.kind}</span>
    }

    const totalColumnsValue = blockDefinition.columns?.reduce(
      (prev: number, current: number) => prev + current,
      0
    )

    // create a container per children
    childContent = (
      <div style={{ display: 'table', width: '100%', height: '1px' }}>
        {props.block.children?.map((_childBlock, i) => {
          const col =
            blockDefinition.columns && blockDefinition.columns[i] ? blockDefinition.columns[i] : 100
          const width = col === 0 ? 0 : ((col * 100) / totalColumnsValue).toFixed(2)
          const containerStyle: CSSProperties = {
            display: 'table-cell',
            verticalAlign: 'top',
            width: width + '%'
          }

          // detect we are on mobile to stack the columns
          if (
            !props.block.data.columnsOnMobile &&
            props.block.data.stackColumnsAtWidth &&
            parseInt(props.block.data.stackColumnsAtWidth) >= props.deviceWidth
          ) {
            containerStyle.display = 'block'
            containerStyle.width = '100%'
          }

          if (props.block.children.length === 0) {
            containerStyle.background =
              "url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAoAAAAKCAYAAACNMs+9AAAAAXNSR0IArs4c6QAAAERlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAA6ABAAMAAAABAAEAAKACAAQAAAABAAAACqADAAQAAAABAAAACgAAAAA7eLj1AAAAK0lEQVQYGWP8DwQMaODZs2doIgwMTBgiOAQGUCELNodLSUlhuHQA3Ui01QDcPgnEE5wAOwAAAABJRU5ErkJggg==')"
          }

          return (
            <Container
              key={i}
              groupName={blockDefinition.containerGroup}
              onDrop={(dropResult) => props.onContainerDrop(props.block.path, dropResult)}
              getChildPayload={(i) => props.block.children[i]}
              dragHandleSelector=".cmeditor-drag-handle"
              style={containerStyle}
              dragClass="cmeditor-ghost-drag"
              dropClass="cmeditor-ghost-drop"
              dropPlaceholder={{
                animationDuration: 200,
                showOnTop: true,
                className: 'cmeditor-drop-preview'
              }}
              // onDragEnter={() => {
              //     console.log('onDragEnter 2')
              // }}
            >
              {RenderChild(props, i)}
            </Container>
          )
        })}
      </div>
    )
  } else if (blockDefinition.containsDraggables === true) {
    const containerStyle: CSSProperties = {
      height: '100%' // content level, should match other columns heights
    }

    if (props.block.children.length === 0) {
      containerStyle.background =
        "url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAoAAAAKCAYAAACNMs+9AAAAAXNSR0IArs4c6QAAAERlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAA6ABAAMAAAABAAEAAKACAAQAAAABAAAACqADAAQAAAABAAAACgAAAAA7eLj1AAAAK0lEQVQYGWP8DwQMaODZs2doIgwMTBgiOAQGUCELNodLSUlhuHQA3Ui01QDcPgnEE5wAOwAAAABJRU5ErkJggg==')"
    }

    childContent = (
      <>
        <Container
          groupName={blockDefinition.containerGroup}
          onDrop={(dropResult) => props.onContainerDrop(props.block.path, dropResult)}
          getChildPayload={(i) => props.block.children[i]}
          dragHandleSelector=".cmeditor-drag-handle"
          dragClass="cmeditor-ghost-drag"
          dropClass="cmeditor-ghost-drop"
          dropPlaceholder={{
            animationDuration: 200,
            showOnTop: true,
            className: 'cmeditor-drop-preview'
          }}
          style={containerStyle}
          // onDragEnter={() => {
          //     console.log('onDragEnter')
          // }}
        >
          {RenderChildrenVertically(props, blockDefinition)}
        </Container>
      </>
    )
  } else if (props.block.children.length > 0) {
    // render children even if block is not container
    childContent = RenderChildrenVertically(props, blockDefinition)
  }

  return (
    <div
      ref={blockElement}
      style={props.block.id === 'root' ? { minHeight: '100vh' } : {}}
      id={props.block.id}
      className={cn({
        'cmeditor-block-content': props.block.kind !== 'root',
        // 'focused': props.focusedBlock && props.focusedBlock.id === props.block.id,
        'cmeditor-selected': props.block.id === props.selectedBlockId
      })}
      onClick={props.onSelectBlock.bind(null, props.block)}
      onMouseOver={(e: MouseEvent) => {
        // avoid propagation that would trigger underlying blocks
        e.stopPropagation()
        props.onFocusBlock(blockElement)
      }}
    >
      {props.block.id !== 'root' && props.block.id === props.selectedBlockId && (
        <span
          className="cmeditor-drag-handle"
          onMouseOver={(event: MouseEvent) => {
            // stop propagation when we over the handles, to avoid focus underlying block
            // when the handles are in absolute position outside of the current block
            event.stopPropagation()
            props.onFocusBlock(undefined)
          }}
        >
          {props.renderSelectedBlockButtons({
            isDraggable: true,
            block: props.block,
            blockDefinitions: props.blockDefinitions,
            deleteBlock: props.deleteBlock,
            cloneBlock: props.cloneBlock
          })}
        </span>
      )}
      {blockDefinition.renderEditor(props, childContent)}
    </div>
  )
}
// , (props: BlockEditorRenderer, nextProps: NodeComponentProps) => {
//     // dont rerender if nodes are equals
//     const equals = _.isEqual(props.node, nextProps.node)
//     // console.log('equals', equals)
//     return equals
// })
