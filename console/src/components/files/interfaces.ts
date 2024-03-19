export interface FileManagerProps {
  foldersTree: FoldersTree
  currentPath?: string
  itemFilters?: ItemFilter[]
  onError: (error: any) => void
  onSelect: (items: Item[]) => void
  height: number
  acceptFileType: string
  acceptItem: (item: Item) => boolean
  multiple?: boolean
}

export interface FoldersTree {
  id: string
  path: string
  name: string
  files_loaded: boolean
  children: FoldersTree[]
}

export interface Item {
  id: string
  path: string
  name: string
  deleted_at?: string
  metadata: any

  // item
  url: string
  contentType: string
  size: number
  width?: number
  height?: number
  persistedAt?: number
  lastModifiedAt: number
  // updatedAt: Date
  uploadState?: string
  uploadProgress?: number
  uploadedAt?: number
  file?: File
}

export interface ItemFilter {
  key: string // item key
  value: any
  operator: string // contains equals greaterThan lessThan
}
