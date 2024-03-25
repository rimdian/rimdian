import { ReactNode } from 'react'
import { BlockEditorRendererProps } from './BlockEditorRenderer'

export type BlockDefinitionMap = {
    [index: string]: BlockDefinitionInterface
}

export interface BlockRenderSettingsProps {
    block: BlockInterface
    tree: BlockInterface
    updateTree: (path: string, data: any) => void
}

export interface BlockDefinitionInterface {
    name: string
    kind: string
    containsDraggables: boolean
    containerGroup?: string
    isDraggable: boolean
    draggableIntoGroup?: string
    isDeletable: boolean
    children?: BlockDefinitionInterface[]
    columns?: number[]
    defaultData: any
    menuSettings: any
    RenderSettings: (props: BlockRenderSettingsProps) => ReactNode // render in settings
    renderEditor: (props: BlockEditorRendererProps, content: ReactNode) => ReactNode // render in editor
    // renderEditorEmptyChildren?: (props: BlockEditorRendererProps) => ReactNode // render in editor
    renderMenu?: (blockDefinition: BlockDefinitionInterface) => ReactNode,
    // transformer: (block: BlockInterface) => any // to MJML to HTML
    // deserialize: (json: object) => BlockInterface // JSON to block
    // serialize: (block: BlockInterface) => object // block to JSON
}

// instance of block in the editor
export interface BlockInterface {
    id: string // unique id
    kind: string
    path: string // path in the tree
    children: BlockInterface[]
    data: any
}

// export interface BlockSettingsRenderProps {
//     block: BlockInterface
//     onUpdate: (block: BlockInterface) => void
// }